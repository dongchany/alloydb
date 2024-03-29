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

package plans

import (
	"github.com/juju/errors"
	"github.com/Dong-Chan/alloydb/context"
	"github.com/Dong-Chan/alloydb/expression"
	"github.com/Dong-Chan/alloydb/field"
	"github.com/Dong-Chan/alloydb/kv/memkv"
	"github.com/Dong-Chan/alloydb/plan"
	"github.com/Dong-Chan/alloydb/util/format"
)

var (
	_ plan.Plan = (*UnionPlan)(nil)
)

// UnionPlan handles UNION query, merges results from different plans.
type UnionPlan struct {
	Srcs      []plan.Plan
	Distincts []bool
	RFields   []*field.ResultField
}

// Do implements plan.Plan Do interface.
// In union plan, we should evaluate each selects from left to right, and union them according to distinct option
// Use the first select ResultFields as the final ResultFields but we should adjust its Flen for the rest Select ResultFields.
func (p *UnionPlan) Do(ctx context.Context, f plan.RowIterFunc) error {
	if len(p.Srcs) == 0 {
		return nil
	}
	t, err := memkv.CreateTemp(true)
	if err != nil {
		return err
	}

	defer func() {
		if derr := t.Drop(); derr != nil && err == nil {
			err = derr
		}
	}()
	// The final result
	var rows [][]interface{}

	// Eval the first select statement
	src := p.Srcs[0]
	src.Do(ctx, func(rid interface{}, in []interface{}) (bool, error) {
		rows = append(rows, in)
		if err := t.Set(in, []interface{}{true}); err != nil {
			return false, err
		}
		return true, nil
	})
	// Use the ResultFields of the first select statement as the final ResultFields
	rfs := src.GetFields()

	// Eval the following select statements
	for i, distinct := range p.Distincts {
		var err error
		src = p.Srcs[i+1]
		// Eval src
		if len(src.GetFields()) != len(rfs) {
			return errors.New("The used SELECT statements have a different number of columns")
		}
		src.Do(ctx, func(rid interface{}, in []interface{}) (bool, error) {
			srcRfs := src.GetFields()
			for i := range in {
				// The column value should be casted as the same type of the first select statement in corresponding position
				srcRf := srcRfs[i]
				rf := rfs[i]
				/*
				 * The lengths of the columns in the UNION result take into account the values retrieved by all of the SELECT statements
				 * SELECT REPEAT('a',1) UNION SELECT REPEAT('b',10);
				 * +---------------+
				 * | REPEAT('a',1) |
				 * +---------------+
				 * | a             |
				 * | bbbbbbbbbb    |
				 * +---------------+
				 */
				if srcRf.Flen > rf.Col.Flen {
					rf.Col.Flen = srcRf.Col.Flen
				}
				in[i], err = rf.Col.CastValue(ctx, in[i])
				if err != nil {
					return false, err
				}
			}
			if distinct {
				// distinct union, check duplicate
				v, err := t.Get(in)
				if err != nil {
					return false, err
				}
				if len(v) > 0 {
					// Find duplicate, ignore it
					return true, nil
				}
			}
			rows = append(rows, in)
			if err := t.Set(in, []interface{}{true}); err != nil {
				return false, err
			}
			return true, nil
		})
	}
	var more bool
	for _, row := range rows {
		if more, err = f(nil, row); !more || err != nil {
			break
		}
	}
	return nil
}

// Explain implements plan.Plan Explain interface.
func (p *UnionPlan) Explain(w format.Formatter) {
	// TODO: Finish this
}

// GetFields implements plan.Plan GetFields interface.
func (p *UnionPlan) GetFields() []*field.ResultField {
	return p.Srcs[0].GetFields()
}

// Filter implements plan.Plan Filter interface.
func (p *UnionPlan) Filter(ctx context.Context, expr expression.Expression) (plan.Plan, bool, error) {
	return p, false, nil
}
