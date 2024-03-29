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

package expressions

import (
	"fmt"

	"github.com/Dong-Chan/alloydb/context"
	"github.com/Dong-Chan/alloydb/expression"
	"github.com/Dong-Chan/alloydb/util/types"
)

var (
	_ expression.Expression = (*IsTruth)(nil)
)

// IsTruth is the expression for true/false check.
type IsTruth struct {
	// Expr is the expression to be checked.
	Expr expression.Expression
	// Not is true, the expression is "is not true/false".
	Not bool
	// True indicates checking true or false.
	True int8
}

// Clone implements the Expression Clone interface.
func (is *IsTruth) Clone() (expression.Expression, error) {
	expr, err := is.Expr.Clone()
	if err != nil {
		return nil, err
	}

	return &IsTruth{Expr: expr, Not: is.Not, True: is.True}, nil
}

// IsStatic implements the Expression IsStatic interface.
func (is *IsTruth) IsStatic() bool {
	return is.Expr.IsStatic()
}

// String implements the Expression String interface.
func (is *IsTruth) String() string {
	not := ""
	if is.Not {
		not = "NOT "
	}

	truth := "TRUE"
	if is.True == 0 {
		truth = "FALSE"
	}

	return fmt.Sprintf("%s IS %s%s", is.Expr, not, truth)
}

// Eval implements the Expression Eval interface.
func (is *IsTruth) Eval(ctx context.Context, args map[interface{}]interface{}) (v interface{}, err error) {
	val, err := is.Expr.Eval(ctx, args)
	if err != nil {
		return
	}

	if val == nil {
		// null is true/false -> false
		// null is not true/false -> true
		return is.Not, nil
	}

	b, err := types.ToBool(val)
	if err != nil {
		return
	}

	if !is.Not {
		// true/false is true/false
		return b == is.True, nil
	}

	// true/false is not true/false
	return b != is.True, nil
}
