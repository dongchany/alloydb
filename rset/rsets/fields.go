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
	"strings"

	"github.com/Dong-Chan/alloydb/context"
	"github.com/Dong-Chan/alloydb/expression/expressions"
	"github.com/Dong-Chan/alloydb/field"
	"github.com/Dong-Chan/alloydb/plan"
	"github.com/Dong-Chan/alloydb/plan/plans"
)

var (
	_ plan.Planner = (*SelectFieldsRset)(nil)
	_ plan.Planner = (*FieldRset)(nil)
)

// SelectFieldsRset is record set to select fields.
type SelectFieldsRset struct {
	Src        plan.Plan
	SelectList *plans.SelectList
}

// Plan gets SrcPlan/SelectFieldsDefaultPlan.
// If all fields are equal to src plan fields, then gets SrcPlan.
// Default gets SelectFieldsDefaultPlan.
func (r *SelectFieldsRset) Plan(ctx context.Context) (plan.Plan, error) {
	fields := r.SelectList.Fields
	srcFields := r.Src.GetFields()
	if len(fields) == len(srcFields) {
		match := true
		for i, v := range fields {
			// TODO: is it this check enough? e.g, the ident field is t.c.
			if x, ok := v.Expr.(*expressions.Ident); ok && strings.EqualFold(x.L, srcFields[i].Name) && strings.EqualFold(v.Name, srcFields[i].Name) {
				continue
			}

			match = false
			break
		}

		if match {
			return r.Src, nil
		}
	}

	src := r.Src
	if x, ok := src.(*plans.TableDefaultPlan); ok {
		// check whether src plan will be set TableNilPlan, like `select 1, 2 from t`.
		isConst := true
		for _, v := range fields {
			if expressions.FastEval(v.Expr) == nil {
				isConst = false
				break
			}
		}
		if isConst {
			src = &plans.TableNilPlan{x.T}
		}
	}

	p := &plans.SelectFieldsDefaultPlan{Src: src, SelectList: r.SelectList}
	return p, nil
}

// FieldRset is Recordset for select expression, like `select 1, 1+1`.
type FieldRset struct {
	Fields []*field.Field
}

// Plan gets SelectExprPlan.
func (r *FieldRset) Plan(ctx context.Context) (plan.Plan, error) {
	return &plans.SelectEmptyFieldListPlan{Fields: r.Fields}, nil
}
