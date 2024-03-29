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
	"github.com/Dong-Chan/alloydb/context"
	"github.com/Dong-Chan/alloydb/expression"
	"github.com/Dong-Chan/alloydb/expression/expressions"
	"github.com/Dong-Chan/alloydb/field"
	"github.com/Dong-Chan/alloydb/plan"
	"github.com/Dong-Chan/alloydb/util/format"
	"github.com/Dong-Chan/alloydb/util/types"
)

var (
	_ plan.Plan = (*FilterDefaultPlan)(nil)
)

// FilterDefaultPlan handles WHERE statement, filters rows by specific
// expressions.
type FilterDefaultPlan struct {
	plan.Plan
	Expr expression.Expression
}

// Explain implements plan.Plan Explain interface.
func (r *FilterDefaultPlan) Explain(w format.Formatter) {
	r.Plan.Explain(w)
	w.Format("┌FilterDefaultPlan Filter on %v\n", r.Expr)
	w.Format("└Output field names %v\n", field.RFQNames(r.GetFields()))
}

// Do implements plan.Plan Do interface.
func (r *FilterDefaultPlan) Do(ctx context.Context, f plan.RowIterFunc) (err error) {
	m := map[interface{}]interface{}{}
	fields := r.GetFields()
	return r.Plan.Do(ctx, func(rid interface{}, data []interface{}) (bool, error) {
		m[expressions.ExprEvalIdentFunc] = func(name string) (interface{}, error) {
			return getIdentValue(name, fields, data, field.DefaultFieldFlag)
		}
		val, err := r.Expr.Eval(ctx, m)
		if err != nil {
			return false, err
		}

		if val == nil {
			return true, nil
		}

		// Evaluate the expression, if the result is true, go on, otherwise
		// skip this row.
		x, err := types.ToBool(val)
		if err != nil {
			return false, err
		}

		if x == 0 {
			return true, nil
		}
		return f(rid, data)
	})
}
