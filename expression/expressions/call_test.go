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
	"errors"

	. "github.com/pingcap/check"
	"github.com/Dong-Chan/alloydb/expression"
)

var _ = Suite(&testCallSuite{})

type testCallSuite struct {
}

func (s *testCallSuite) TestCall(c *C) {
	f, err := NewCall("abs", []expression.Expression{Value{1}}, false)
	c.Assert(err, IsNil)

	c.Assert(f.IsStatic(), IsTrue)
	_, err = f.Clone()
	c.Assert(err, IsNil)

	c.Assert(len(f.String()), Greater, 0)
	m := map[interface{}]interface{}{}
	_, err = f.Eval(nil, m)
	c.Assert(err, IsNil)

	mock := mockExpr{
		isStatic: false,
	}

	f, err = NewCall("abs", []expression.Expression{mock}, true)
	c.Assert(err, IsNil)
	c.Assert(len(f.String()), Greater, 0)
	c.Assert(f.IsStatic(), IsFalse)
	_, err = f.Clone()
	c.Assert(err, IsNil)

	// test error
	f, err = NewCall("abs", []expression.Expression{Value{1}, Value{2}}, true)
	c.Assert(err, NotNil)

	_, err = NewCall("must_error_func", nil, false)
	c.Assert(err, NotNil)

	fc := Call{
		F:        "must_error_func",
		Args:     nil,
		Distinct: false,
	}
	_, err = fc.Eval(nil, nil)
	c.Assert(err, NotNil)

	mock.err = errors.New("must error")
	mock.isStatic = true
	f, err = NewCall("abs", []expression.Expression{mock}, true)
	c.Assert(err, NotNil)

	mock.isStatic = false
	f, err = NewCall("abs", []expression.Expression{mock}, true)
	c.Assert(err, IsNil)

	_, err = f.Clone()
	c.Assert(err, NotNil)
	_, err = f.Eval(nil, nil)
	c.Assert(err, NotNil)

}
