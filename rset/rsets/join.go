//
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package rsets

import (
	"fmt"
	"strings"

	"github.com/juju/errors"
	"github.com/Dong-Chan/alloydb/context"
	"github.com/Dong-Chan/alloydb/expression"
	"github.com/Dong-Chan/alloydb/field"
	"github.com/Dong-Chan/alloydb/plan"
	"github.com/Dong-Chan/alloydb/plan/plans"
	"github.com/Dong-Chan/alloydb/stmt"
	"github.com/Dong-Chan/alloydb/table"
)

var (
	_ plan.Planner = (*JoinRset)(nil)
)

// JoinType is join type, including cross/left/right/full.
type JoinType int

// String gets string value of join type.
func (t JoinType) String() string {
	switch t {
	case CrossJoin:
		return plans.CrossJoin
	case LeftJoin:
		return plans.LeftJoin
	case RightJoin:
		return plans.RightJoin
	case FullJoin:
		return plans.FullJoin
	}

	return "Unknown"
}

const (
	// CrossJoin is cross join type.
	CrossJoin JoinType = iota + 1
	// LeftJoin is left Join type.
	LeftJoin
	// RightJoin is right Join type.
	RightJoin
	// FullJoin is full Join type.
	FullJoin
)

// JoinRset is record set for table join.
type JoinRset struct {
	Left  interface{}
	Right interface{}
	Type  JoinType
	On    expression.Expression

	tableNames map[string]bool
}

func (r *JoinRset) String() string {
	left := r.formatNode(r.Left)

	if r.Right == nil {
		return left
	}

	onStr := ""
	if r.On != nil {
		onStr = fmt.Sprintf(" ON %s", r.On.String())
	}

	right := r.formatNode(r.Right)
	return fmt.Sprintf("%v %s JOIN %v%s", left, r.Type, right, onStr)
}

func (r *JoinRset) formatNode(node interface{}) string {
	switch x := node.(type) {
	case *TableSource:
		return x.String()
	case *JoinRset:
		return fmt.Sprintf("(%s)", x.String())
	}

	panic(fmt.Sprintf("invalid join source %T", node))
}

func (r *JoinRset) buildJoinPlan(ctx context.Context, p *plans.JoinPlan, s *JoinRset) error {
	left, leftFields, err := r.buildPlan(ctx, s.Left)
	if err != nil {
		return errors.Trace(err)
	}

	right, rightFields, err := r.buildPlan(ctx, s.Right)
	if err != nil {
		return errors.Trace(err)
	}

	p.Left = left
	p.Right = right
	p.On = s.On
	p.Type = s.Type.String()

	// if the left is not JoinPlan, the left is the leaf node,
	// we can stop recursion.
	if pl, ok := left.(*plans.JoinPlan); ok {
		if err := r.buildJoinPlan(ctx, pl, s.Left.(*JoinRset)); err != nil {
			return errors.Trace(err)
		}

		// use left JoinPlan fields.
		p.Fields = append(p.Fields, pl.GetFields()...)
	} else {
		// use left fields directly.
		p.Fields = append(p.Fields, leftFields...)
	}

	// if the right is not JoinPlan, the right is the leaf node,
	// we can stop recursion.
	if pr, ok := right.(*plans.JoinPlan); ok {
		if err := r.buildJoinPlan(ctx, pr, s.Right.(*JoinRset)); err != nil {
			return errors.Trace(err)
		}
		// use right JoinPlan fields.
		p.Fields = append(p.Fields, pr.GetFields()...)
	} else {
		// use right fields directly.
		p.Fields = append(p.Fields, rightFields...)
	}

	return nil
}

func (r *JoinRset) buildPlan(ctx context.Context, node interface{}) (plan.Plan, []*field.ResultField, error) {
	switch t := node.(type) {
	case *JoinRset:
		return &plans.JoinPlan{}, nil, nil
	case *TableSource:
		return r.buildSourcePlan(ctx, t)
	case nil:
		return nil, nil, nil
	default:
		return nil, nil, errors.Errorf("invalid join node %T", t)
	}
}

func (r *JoinRset) buildSourcePlan(ctx context.Context, t *TableSource) (plan.Plan, []*field.ResultField, error) {
	var (
		src interface{}
		tr  *TableRset
		err error
	)
	switch s := t.Source.(type) {
	case table.Ident:
		fullIdent := s.Full(ctx)
		tr, err = newTableRset(fullIdent.Schema.O, fullIdent.Name.O)
		if err != nil {
			return nil, nil, errors.Trace(err)
		}
		src = tr
		if t.Name == "" {
			qualifiedTableName := tr.Schema + "." + tr.Name
			if r.tableNames[qualifiedTableName] {
				return nil, nil, errors.Errorf("%s: duplicate name %s", r.String(), s)
			}
			r.tableNames[qualifiedTableName] = true
		}
	case stmt.Statement:
		src = s
	default:
		return nil, nil, errors.Errorf("invalid table source %T", t.Source)
	}

	var p plan.Plan
	switch x := src.(type) {
	case plan.Planner:
		if p, err = x.Plan(ctx); err != nil {
			return nil, nil, errors.Trace(err)
		}
	case plan.Plan:
		p = x
	default:
		return nil, nil, errors.Errorf("invalid table source %T, no Plan interface", t.Source)
	}

	var fields []*field.ResultField
	dupNames := make(map[string]struct{}, len(p.GetFields()))
	for _, nf := range p.GetFields() {
		f := nf.Clone()
		if t.Name != "" {
			f.TableName = t.Name
		}

		// duplicate column name in one table is not allowed.
		name := strings.ToLower(f.Name)
		if _, ok := dupNames[name]; ok {
			return nil, nil, errors.Errorf("Duplicate column name '%s'", name)
		}
		dupNames[name] = struct{}{}

		fields = append(fields, f)
	}

	return p, fields, nil
}

// Plan gets JoinPlan.
func (r *JoinRset) Plan(ctx context.Context) (plan.Plan, error) {
	r.tableNames = make(map[string]bool)
	p := &plans.JoinPlan{}
	if err := r.buildJoinPlan(ctx, p, r); err != nil {
		return nil, errors.Trace(err)
	}

	return p, nil
}
