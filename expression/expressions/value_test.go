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
)

var _ = Suite(&testValueSuite{})

type testValueSuite struct {
}

func (s *testValueSuite) TestValue(c *C) {
	e := Value{nil}
	c.Assert(e.IsStatic(), IsTrue)
	c.Assert(e.String(), Equals, "NULL")

	_, err := e.Clone()
	c.Assert(err, IsNil)

	v, err := e.Eval(nil, nil)
	c.Assert(err, IsNil)
	c.Assert(v, IsNil)

	e.Val = "1"
	c.Assert(e.String(), Equals, "\"1\"")

	e.Val = 1
	c.Assert(e.String(), Equals, "1")
}
