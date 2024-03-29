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
	"strings"

	"github.com/juju/errors"
	"github.com/Dong-Chan/alloydb/context"
	"github.com/Dong-Chan/alloydb/expression"
	"github.com/Dong-Chan/alloydb/sessionctx/variable"
)

var (
	_ expression.Expression = (*Variable)(nil)
)

// Variable is the expression for variable.
type Variable struct {
	// Name is the variable name.
	Name string
	// IsGlobal indicates whether this variable is global.
	IsGlobal bool
	// IsSystem indicates whether this variable is a global variable in current session.
	IsSystem bool
}

// Clone implements the Expression Clone interface.
func (v *Variable) Clone() (expression.Expression, error) {
	newVar := *v
	return &newVar, nil
}

// IsStatic implements the Expression IsStatic interface, always returns false.
func (v *Variable) IsStatic() bool {
	return false
}

// String implements the Expression String interface.
func (v *Variable) String() string {
	if !v.IsSystem {
		return "@" + v.Name
	}
	if v.IsGlobal {
		return "@@GLOBAL." + v.Name
	}
	return "@@" + v.Name
}

// Eval implements the Expression Eval interface.
func (v *Variable) Eval(ctx context.Context, args map[interface{}]interface{}) (interface{}, error) {
	name := strings.ToLower(v.Name)
	sessionVars := variable.GetSessionVars(ctx)
	if !v.IsSystem {
		// user vars
		if value, ok := sessionVars.Users[name]; ok {
			return value, nil
		}
		// select null user vars is permitted.
		return nil, nil
	}

	sysVar, ok := variable.SysVars[name]
	if !ok {
		// select null sys vars is not permitted
		return nil, errors.Errorf("Unknown system variable '%s'", name)
	}

	if !v.IsGlobal {
		if value, ok := sessionVars.Systems[name]; ok {
			return value, nil
		}
	}

	// TODO: select global sys var from kv store, now we only get it from memory
	return sysVar.Value, nil
}
