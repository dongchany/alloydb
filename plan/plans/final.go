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
	"github.com/Dong-Chan/alloydb/field"
	"github.com/Dong-Chan/alloydb/plan"
	"github.com/Dong-Chan/alloydb/util/format"
)

var (
	_ plan.Plan = (*SelectFinalPlan)(nil)
)

// SelectFinalPlan sets info field for resuilt.
type SelectFinalPlan struct {
	*SelectList

	Src     plan.Plan
	infered bool // If result field info is already infered
}

// Do implements the plan.Plan Do interface, and sets result info field.
func (r *SelectFinalPlan) Do(ctx context.Context, f plan.RowIterFunc) error {
	// Reset infered. For prepared statements, this plan may run many times.
	r.infered = false
	return r.Src.Do(ctx, func(rid interface{}, in []interface{}) (bool, error) {
		// we should not output hidden fields to client
		out := in[0:r.HiddenFieldOffset]
		for i, o := range out {
			switch v := o.(type) {
			case bool:
				// Convert bool field to int
				if v {
					out[i] = uint8(1)
				} else {
					out[i] = uint8(0)
				}
			}
		}
		if !r.infered {
			setResultFieldInfo(r.ResultFields[0:r.HiddenFieldOffset], out)
			r.infered = true
		}
		return f(rid, out)
	})
}

// Explain implements the plan.Plan Explain interface.
func (r *SelectFinalPlan) Explain(w format.Formatter) {
	r.Src.Explain(w)
	if r.HiddenFieldOffset == len(r.Src.GetFields()) {
		// we have no hidden fields, can return.
		return
	}
	w.Format("┌Evaluate\n└Output field names %v\n", field.RFQNames(r.ResultFields[0:r.HiddenFieldOffset]))
}

// GetFields implements the plan.Plan GetFields interface.
func (r *SelectFinalPlan) GetFields() []*field.ResultField {
	return r.ResultFields[0:r.HiddenFieldOffset]
}

// Filter implements the plan.Plan Filter interface.
func (r *SelectFinalPlan) Filter(ctx context.Context, expr expression.Expression) (p plan.Plan, filtered bool, err error) {
	return r, false, nil
}
