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

package table_test

import (
	"testing"

	. "github.com/pingcap/check"
	"github.com/Dong-Chan/alloydb"
	"github.com/Dong-Chan/alloydb/context"
	"github.com/Dong-Chan/alloydb/model"
	"github.com/Dong-Chan/alloydb/sessionctx/db"
	"github.com/Dong-Chan/alloydb/store/localstore"
	"github.com/Dong-Chan/alloydb/store/localstore/goleveldb"
	"github.com/Dong-Chan/alloydb/table"
)

func TestT(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&testSuite{})

type testSuite struct {
}

func (*testSuite) TestT(c *C) {
	var ident = table.Ident{
		Name: model.NewCIStr("t"),
	}
	c.Assert(ident.String(), Not(Equals), "")
	driver := localstore.Driver{goleveldb.MemoryDriver{}}
	store, err := driver.Open("memory")
	c.Assert(err, IsNil)
	se, err := alloydb.CreateSession(store)
	c.Assert(err, IsNil)
	ctx := se.(context.Context)
	db.BindCurrentSchema(ctx, "test")
	fullIdent := ident.Full(ctx)
	c.Assert(fullIdent.Schema.L, Equals, "test")
	c.Assert(fullIdent.Name.L, Equals, "t")
	c.Assert(fullIdent.String(), Not(Equals), "")
	fullIdent2 := fullIdent.Full(ctx)
	c.Assert(fullIdent2.Schema.L, Equals, fullIdent.Schema.L)
}
