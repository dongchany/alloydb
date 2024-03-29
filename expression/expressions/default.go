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
	"strings"

	"github.com/juju/errors"
	"github.com/Dong-Chan/alloydb/context"
	"github.com/Dong-Chan/alloydb/expression"
)

var (
	_ expression.Expression = (*Default)(nil)
)

// Default is the default expression using default value for a column.
type Default struct {
	// Name is the column name
	Name string
}

// Clone implements the Expression Clone interface.
func (v *Default) Clone() (expression.Expression, error) {
	newV := *v
	return &newV, nil
}

// IsStatic implements the Expression IsStatic interface, always returns false.
func (v *Default) IsStatic() bool {
	return false
}

// String implements the Expression String interface.
func (v *Default) String() string {
	if v.Name == "" {
		return "default"
	}

	return fmt.Sprintf("default (%s)", strings.ToLower(v.Name))
}

// Eval implements the Expression Eval interface.
func (v *Default) Eval(ctx context.Context, args map[interface{}]interface{}) (interface{}, error) {
	name := strings.ToLower(v.Name)
	if name == "" {
		// if name is empty, the stmt may like "insert into t values (default)"
		// we will use the corresponding column name
		colName, ok := args[ExprEvalDefaultName]
		if !ok {
			return nil, errors.Errorf("default column not found - %s", name)
		}
		name = colName.(string)
	}

	vv, ok := args[name]
	if ok {
		return vv, nil
	}

	return nil, errors.Errorf("default column not found - %s", name)
}
