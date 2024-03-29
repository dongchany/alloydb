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

package field_test

import (
	"testing"

	. "github.com/pingcap/check"
	"github.com/Dong-Chan/alloydb/expression/expressions"
	"github.com/Dong-Chan/alloydb/field"
	mysql "github.com/Dong-Chan/alloydb/mysqldef"
	"github.com/Dong-Chan/alloydb/util/types"
)

func TestT(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&testFieldSuite{})

type testFieldSuite struct {
}

func (*testFieldSuite) TestField(c *C) {
	f := &field.Field{
		Expr: expressions.Value{Val: "c1+1"},
		Name: "a",
	}
	s := f.String()
	c.Assert(len(s), Greater, 0)

	ft := types.NewFieldType(mysql.TypeLong)
	ft.Flen = 20
	ft.Flag |= mysql.UnsignedFlag | mysql.ZerofillFlag
	c.Assert(ft.String(), Equals, "INT (20) UNSIGNED ZEROFILL")

	ft = types.NewFieldType(mysql.TypeFloat)
	ft.Flen = 20
	ft.Decimal = 10
	c.Assert(ft.String(), Equals, "FLOAT (20, 10)")

	ft = types.NewFieldType(mysql.TypeTimestamp)
	ft.Decimal = 8
	c.Assert(ft.String(), Equals, "TIMESTAMP (8)")

	ft = types.NewFieldType(mysql.TypeVarchar)
	ft.Flag |= mysql.BinaryFlag
	ft.Charset = "utf8"
	ft.Collate = "utf8_unicode_gi"
	c.Assert(ft.String(), Equals, "VARCHAR BINARY CHARACTER SET utf8 COLLATE utf8_unicode_gi")
}
