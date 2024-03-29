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
	. "github.com/pingcap/check"
	"github.com/Dong-Chan/alloydb/model"
)

var _ = Suite(&testIdentSuite{})

type testIdentSuite struct {
}

func (s *testIdentSuite) TestIdent(c *C) {
	e := Ident{
		model.NewCIStr("id"),
	}

	c.Assert(e.IsStatic(), IsFalse)
	c.Assert(e.String(), Equals, "id")

	ec, err := e.Clone()
	e2, ok := ec.(*Ident)
	c.Assert(ok, IsTrue)
	e2.O = "ID"
	c.Assert(e.Equal(e2), IsTrue)

	m := map[interface{}]interface{}{}
	v, err := e.Eval(nil, m)
	c.Assert(err, NotNil)

	m[ExprEvalArgAggEmpty] = struct{}{}
	v, err = e.Eval(nil, m)
	c.Assert(err, IsNil)
	c.Assert(v, IsNil)

	delete(m, ExprEvalArgAggEmpty)
	m[ExprEvalIdentFunc] = func(string) (interface{}, error) {
		return 1, nil
	}

	v, err = e.Eval(nil, m)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, 1)
}
