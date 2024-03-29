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
	"github.com/juju/errors"
	"github.com/Dong-Chan/alloydb/context"
	"github.com/Dong-Chan/alloydb/expression"
	"github.com/Dong-Chan/alloydb/model"
)

var (
	_ expression.Expression = (*Ident)(nil)
)

// Ident is the identifier expression.
type Ident struct {
	// model.CIStr contains origin identifier name and its lowercase name.
	model.CIStr
}

// Clone implements the Expression Clone interface.
func (i *Ident) Clone() (expression.Expression, error) {
	newI := *i
	return &newI, nil
}

// IsStatic implements the Expression IsStatic interface, always returns false.
func (i *Ident) IsStatic() bool {
	return false
}

// String implements the Expression String interface.
func (i *Ident) String() string {
	return i.O
}

// Equal checks equation with another Ident expression using lowercase identifier name.
func (i *Ident) Equal(x *Ident) bool {
	return i.L == x.L
}

// Eval implements the Expression Eval interface.
func (i *Ident) Eval(ctx context.Context, args map[interface{}]interface{}) (v interface{}, err error) {
	if _, ok := args[ExprEvalArgAggEmpty]; ok {
		// select c1, max(c1) from t where c1 = null, must get "NULL", "NULL" for empty table
		return nil, nil
	}

	if f, ok := args[ExprEvalIdentFunc]; ok {
		if got, ok := f.(func(string) (interface{}, error)); ok {
			return got(i.L)
		}
	}

	// defer func() { log.Errorf("Ident %q -> %v %v", i.S, v, err) }()
	v, ok := args[i.L]
	if !ok {
		err = errors.Errorf("unknown field %s %v", i.O, args)
	}
	return
}
