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
	"github.com/Dong-Chan/alloydb/context"
	"github.com/Dong-Chan/alloydb/plan"
	"github.com/Dong-Chan/alloydb/plan/plans"
)

var (
	_ plan.Planner = (*DistinctRset)(nil)
)

// DistinctRset is record set for distinct fields.
type DistinctRset struct {
	Src        plan.Plan
	SelectList *plans.SelectList
}

// Plan gets DistinctDefaultPlan.
func (r *DistinctRset) Plan(ctx context.Context) (plan.Plan, error) {
	return &plans.DistinctDefaultPlan{Src: r.Src, SelectList: r.SelectList}, nil
}
