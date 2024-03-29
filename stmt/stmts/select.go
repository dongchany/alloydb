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

package stmts

import (
	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/Dong-Chan/alloydb/context"
	"github.com/Dong-Chan/alloydb/expression"
	"github.com/Dong-Chan/alloydb/field"
	"github.com/Dong-Chan/alloydb/parser/coldef"
	"github.com/Dong-Chan/alloydb/plan"
	"github.com/Dong-Chan/alloydb/plan/plans"
	"github.com/Dong-Chan/alloydb/rset"
	"github.com/Dong-Chan/alloydb/rset/rsets"
	"github.com/Dong-Chan/alloydb/sessionctx/variable"
	"github.com/Dong-Chan/alloydb/stmt"
	"github.com/Dong-Chan/alloydb/util/format"
)

var _ stmt.Statement = (*SelectStmt)(nil)

// SelectStmt is a statement to retrieve rows selected from one or more tables.
// See: https://dev.mysql.com/doc/refman/5.7/en/select.html
type SelectStmt struct {
	Distinct bool
	Fields   []*field.Field
	From     *rsets.JoinRset
	GroupBy  *rsets.GroupByRset
	Having   *rsets.HavingRset
	Limit    *rsets.LimitRset
	Offset   *rsets.OffsetRset
	OrderBy  *rsets.OrderByRset
	Where    *rsets.WhereRset
	// TODO: rename Lock
	Lock coldef.LockType

	Text string
}

// Explain implements the stmt.Statement Explain interface.
func (s *SelectStmt) Explain(ctx context.Context, w format.Formatter) {
	p, err := s.Plan(ctx)
	if err != nil {
		w.Format("ERROR: %v\n", err)
		return
	}

	p.Explain(w)
}

// IsDDL implements the stmt.Statement IsDDL interface.
func (s *SelectStmt) IsDDL() bool {
	return false
}

// OriginText implements the stmt.Statement OriginText interface.
func (s *SelectStmt) OriginText() string {
	return s.Text
}

// SetText implements the stmt.Statement SetText interface.
func (s *SelectStmt) SetText(text string) {
	s.Text = text
}

// Plan implements the plan.Planner interface.
// The whole phase for select is
// `from -> where -> lock -> group by -> having -> select fields -> distinct -> order by -> limit -> final`
func (s *SelectStmt) Plan(ctx context.Context) (plan.Plan, error) {
	var (
		r   plan.Plan
		err error
	)

	if s.From != nil {
		r, err = s.From.Plan(ctx)
		if err != nil {
			return nil, err
		}
	} else if s.Fields != nil {
		// Only evaluate fields values.
		fr := &rsets.FieldRset{Fields: s.Fields}
		r, err = fr.Plan(ctx)
		if err != nil {
			return nil, err
		}

	}

	if w := s.Where; w != nil {
		r, err = (&rsets.WhereRset{Expr: w.Expr, Src: r}).Plan(ctx)
		if err != nil {
			return nil, err
		}
	}
	lock := s.Lock
	if variable.IsAutocommit(ctx) {
		// Locking of rows for update using SELECT FOR UPDATE only applies when autocommit
		// is disabled (either by beginning transaction with START TRANSACTION or by setting
		// autocommit to 0. If autocommit is enabled, the rows matching the specification are not locked.
		// See: https://dev.mysql.com/doc/refman/5.7/en/innodb-locking-reads.html
		lock = coldef.SelectLockNone
	}
	r, err = (&rsets.SelectLockRset{Src: r, Lock: lock}).Plan(ctx)
	if err != nil {
		return nil, err
	}

	// Get select list for futher field values evaluation.
	selectList, err := plans.ResolveSelectList(s.Fields, r.GetFields())
	if err != nil {
		return nil, errors.Trace(err)
	}

	var groupBy []expression.Expression
	if s.GroupBy != nil {
		groupBy = s.GroupBy.By
	}

	if s.Having != nil {
		// `having` may contain aggregate functions, and we will add this to hidden fields.
		if err = s.Having.CheckAndUpdateSelectList(selectList, groupBy, r.GetFields()); err != nil {
			return nil, errors.Trace(err)
		}
	}

	if s.OrderBy != nil {
		// `order by` may contain aggregate functions, and we will add this to hidden fields.
		if err = s.OrderBy.CheckAndUpdateSelectList(selectList, r.GetFields()); err != nil {
			return nil, errors.Trace(err)
		}
	}

	switch {
	case !rsets.HasAggFields(selectList.Fields) && s.GroupBy == nil:
		// If no group by and no aggregate functions, we will use SelectFieldsPlan.
		if r, err = (&rsets.SelectFieldsRset{Src: r,
			SelectList: selectList}).Plan(ctx); err != nil {
			return nil, err
		}
	default:
		if r, err = (&rsets.GroupByRset{By: groupBy,
			Src:        r,
			SelectList: selectList}).Plan(ctx); err != nil {
			return nil, err
		}
	}

	if s := s.Having; s != nil {
		if r, err = (&rsets.HavingRset{
			Src:  r,
			Expr: s.Expr}).Plan(ctx); err != nil {
			return nil, err
		}
	}

	if s.Distinct {
		if r, err = (&rsets.DistinctRset{Src: r,
			SelectList: selectList}).Plan(ctx); err != nil {
			return nil, err
		}
	}

	if s := s.OrderBy; s != nil {
		if r, err = (&rsets.OrderByRset{By: s.By,
			Src:        r,
			SelectList: selectList}).Plan(ctx); err != nil {
			return nil, err
		}
	}

	if s := s.Offset; s != nil {
		if r, err = (&rsets.OffsetRset{s.Count, r}).Plan(ctx); err != nil {
			return nil, err
		}
	}
	if s := s.Limit; s != nil {
		if r, err = (&rsets.LimitRset{s.Count, r}).Plan(ctx); err != nil {
			return nil, err
		}
	}

	if r, err = (&rsets.SelectFinalRset{Src: r,
		SelectList: selectList}).Plan(ctx); err != nil {
		return nil, err
	}

	return r, nil
}

// Exec implements the stmt.Statement Exec interface.
func (s *SelectStmt) Exec(ctx context.Context) (rs rset.Recordset, err error) {
	log.Info("SelectStmt trx:")
	r, err := s.Plan(ctx)
	if err != nil {
		return nil, err
	}

	return rsets.Recordset{ctx, r}, nil
}
