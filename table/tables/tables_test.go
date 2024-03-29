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

package tables_test

import (
	"testing"

	. "github.com/pingcap/check"
	"github.com/Dong-Chan/alloydb"
	"github.com/Dong-Chan/alloydb/column"
	"github.com/Dong-Chan/alloydb/context"
	"github.com/Dong-Chan/alloydb/kv"
	"github.com/Dong-Chan/alloydb/model"
	"github.com/Dong-Chan/alloydb/sessionctx"
	"github.com/Dong-Chan/alloydb/store/localstore"
	"github.com/Dong-Chan/alloydb/store/localstore/goleveldb"
	"github.com/Dong-Chan/alloydb/table"
	"github.com/Dong-Chan/alloydb/table/tables"
)

func TestT(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&testSuite{})

type testSuite struct {
	store kv.Storage
	se    alloydb.Session
}

func (ts *testSuite) SetUpSuite(c *C) {
	table.TableFromMeta = tables.TableFromMeta
	driver := localstore.Driver{goleveldb.MemoryDriver{}}
	store, err := driver.Open("memory")
	c.Check(err, IsNil)
	ts.store = store
	ts.se, err = alloydb.CreateSession(ts.store)
	c.Assert(err, IsNil)
	_, err = ts.se.Execute("CREATE DATABASE test")
	c.Assert(err, IsNil)
}

func (ts *testSuite) TestBasic(c *C) {
	_, err := ts.se.Execute("CREATE TABLE test.t (a int primary key auto_increment, b varchar(255) unique)")
	c.Assert(err, IsNil)
	ctx := ts.se.(context.Context)
	dom := sessionctx.GetDomain(ctx)
	tb, err := dom.InfoSchema().TableByName(model.NewCIStr("test"), model.NewCIStr("t"))
	c.Assert(err, IsNil)
	c.Assert(tb.TableID(), Greater, int64(0))
	c.Assert(tb.TableName().L, Equals, "t")
	c.Assert(tb.Meta(), NotNil)
	c.Assert(tb.Indices(), NotNil)
	c.Assert(tb.FirstKey(), Not(Equals), "")
	c.Assert(tb.IndexPrefix(), Not(Equals), "")
	c.Assert(tb.KeyPrefix(), Not(Equals), "")
	c.Assert(tb.FindIndexByColName("b"), NotNil)

	autoid, err := tb.AllocAutoID()
	c.Assert(err, IsNil)
	c.Assert(autoid, Greater, int64(0))

	rid, err := tb.AddRecord(ctx, []interface{}{1, "abc"})
	c.Assert(err, IsNil)
	c.Assert(rid, Greater, int64(0))
	row, err := tb.Row(ctx, rid)
	c.Assert(err, IsNil)
	c.Assert(len(row), Equals, 2)
	c.Assert(row[0].(int32), Equals, int32(1))

	_, err = tb.AddRecord(ctx, []interface{}{1, "aba"})
	c.Assert(err, NotNil)
	_, err = tb.AddRecord(ctx, []interface{}{2, "abc"})
	c.Assert(err, NotNil)

	c.Assert(tb.UpdateRecord(ctx, rid, []interface{}{1, "abc"}, []interface{}{1, "cba"}, []bool{false, true}), IsNil)

	tb.IterRecords(ctx, tb.FirstKey(), tb.Cols(), func(h int64, data []interface{}, cols []*column.Col) (bool, error) {
		return true, nil
	})

	c.Assert(tb.RemoveRowAllIndex(ctx, rid, []interface{}{1, "cba"}), IsNil)

	c.Assert(tb.RemoveRow(ctx, rid), IsNil)
	c.Assert(tb.Truncate(ctx), IsNil)
	_, err = ts.se.Execute("drop table test.t")
	c.Assert(err, IsNil)
}

func (ts *testSuite) TestTypes(c *C) {
	_, err := ts.se.Execute("CREATE TABLE test.t (c1 tinyint, c2 smallint, c3 int, c4 bigint, c5 text, c6 blob, c7 varchar(64), c8 time, c9 timestamp not null default CURRENT_TIMESTAMP, c10 decimal)")
	c.Assert(err, IsNil)
	ctx := ts.se.(context.Context)
	dom := sessionctx.GetDomain(ctx)
	_, err = dom.InfoSchema().TableByName(model.NewCIStr("test"), model.NewCIStr("t"))
	c.Assert(err, IsNil)
	_, err = ts.se.Execute("insert test.t values (1, 2, 3, 4, '5', '6', '7', '10:10:10', null, 1.4)")
	c.Assert(err, IsNil)
	rs, err := ts.se.Execute("select * from test.t where c1 = 1")
	row, err := rs[0].FirstRow()
	c.Assert(err, IsNil)
	c.Assert(row, NotNil)
	_, err = ts.se.Execute("drop table test.t")
	c.Assert(err, IsNil)

	_, err = ts.se.Execute("CREATE TABLE test.t (c1 tinyint unsigned, c2 smallint unsigned, c3 int unsigned, c4 bigint unsigned, c5 double)")
	c.Assert(err, IsNil)
	_, err = ts.se.Execute("insert test.t values (1, 2, 3, 4, 5)")
	c.Assert(err, IsNil)
	rs, err = ts.se.Execute("select * from test.t where c1 = 1")
	row, err = rs[0].FirstRow()
	c.Assert(err, IsNil)
	c.Assert(row, NotNil)
	_, err = ts.se.Execute("drop table test.t")
	c.Assert(err, IsNil)
}
