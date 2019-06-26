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

package ddl_test

import (
	"testing"

	. "github.com/pingcap/check"
	"github.com/Dong-Chan/alloydb"
	"github.com/Dong-Chan/alloydb/context"
	"github.com/Dong-Chan/alloydb/ddl"
	"github.com/Dong-Chan/alloydb/infoschema"
	"github.com/Dong-Chan/alloydb/kv"
	"github.com/Dong-Chan/alloydb/model"
	"github.com/Dong-Chan/alloydb/parser"
	"github.com/Dong-Chan/alloydb/stmt"
	"github.com/Dong-Chan/alloydb/stmt/stmts"
	"github.com/Dong-Chan/alloydb/table"
	qerror "github.com/Dong-Chan/alloydb/util/errors"
	"github.com/Dong-Chan/alloydb/util/errors2"
)

func TestT(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&testSuite{})

type testSuite struct {
	store kv.Storage
}

func (ts *testSuite) SetUpSuite(c *C) {
	store, err := alloydb.NewStore(alloydb.EngineGoLevelDBMemory)
	c.Assert(err, IsNil)
	ts.store = store
}

func (ts *testSuite) TestT(c *C) {
	handle := infoschema.NewHandle(ts.store)
	handle.Set(nil)
	dd := ddl.NewDDL(ts.store, handle)
	se, _ := alloydb.CreateSession(ts.store)
	ctx := se.(context.Context)
	schemaName := model.NewCIStr("test")
	tblName := model.NewCIStr("t")
	tbIdent := table.Ident{
		Schema: schemaName,
		Name:   tblName,
	}
	noExist := model.NewCIStr("noexist")

	err := dd.CreateSchema(ctx, tbIdent.Schema)
	c.Assert(err, IsNil)

	err = dd.CreateSchema(ctx, tbIdent.Schema)
	c.Assert(errors2.ErrorEqual(err, ddl.ErrExists), IsTrue)

	tbStmt := statement("create table t (a int primary key not null, b varchar(255), key idx_b (b), c int, d int unique)").(*stmts.CreateTableStmt)

	err = dd.CreateTable(ctx, table.Ident{Schema: noExist, Name: tbIdent.Name}, tbStmt.Cols, tbStmt.Constraints)
	c.Assert(errors2.ErrorEqual(err, qerror.ErrDatabaseNotExist), IsTrue)
	err = dd.CreateTable(ctx, tbIdent, tbStmt.Cols, tbStmt.Constraints)
	c.Assert(err, IsNil)
	err = dd.CreateTable(ctx, tbIdent, tbStmt.Cols, tbStmt.Constraints)
	c.Assert(errors2.ErrorEqual(err, ddl.ErrExists), IsTrue)

	tbIdent2 := tbIdent
	tbIdent2.Name = model.NewCIStr("t2")
	tbStmt = statement("create table t2 (a int unique not null)").(*stmts.CreateTableStmt)
	err = dd.CreateTable(ctx, tbIdent2, tbStmt.Cols, tbStmt.Constraints)
	c.Assert(err, IsNil)

	tb, err := handle.Get().TableByName(tbIdent.Schema, tbIdent.Name)
	c.Assert(err, IsNil)
	_, err = tb.AddRecord(ctx, []interface{}{1, "b", 2, 4})
	c.Assert(err, IsNil)

	alterStmt := statement("alter table t add column aa int first").(*stmts.AlterTableStmt)
	dd.AlterTable(ctx, tbIdent, alterStmt.Specs)
	c.Assert(alterStmt.Specs[0].String(), Not(Equals), "")
	// Check indices info
	is := handle.Get()
	tbl, err := is.TableByName(schemaName, tblName)
	c.Assert(err, IsNil)
	c.Assert(tbl, NotNil)
	expectedOffset := make(map[string]int)
	expectedOffset["a"] = 1
	expectedOffset["b"] = 2
	expectedOffset["d"] = 4
	for _, idx := range tbl.Indices() {
		for _, col := range idx.Columns {
			o, ok := expectedOffset[col.Name.L]
			c.Assert(ok, IsTrue)
			c.Assert(col.Offset, Equals, o)
		}
	}
	alterStmt = statement("alter table t add column bb int after b").(*stmts.AlterTableStmt)
	dd.AlterTable(ctx, tbIdent, alterStmt.Specs)
	c.Assert(alterStmt.Specs[0].String(), Not(Equals), "")

	idxStmt := statement("CREATE INDEX idx_c ON t (c)").(*stmts.CreateIndexStmt)
	idxName := model.NewCIStr(idxStmt.IndexName)
	err = dd.CreateIndex(ctx, tbIdent, idxStmt.Unique, idxName, idxStmt.IndexColNames)
	c.Assert(err, IsNil)
	tbs := handle.Get().SchemaTables(tbIdent.Schema)
	c.Assert(len(tbs), Equals, 2)
	err = dd.DropIndex(ctx, tbIdent.Schema, tbIdent.Name, idxName)
	c.Assert(err, IsNil)
	err = dd.DropTable(ctx, tbIdent)
	c.Assert(err, IsNil)
	tbs = handle.Get().SchemaTables(tbIdent.Schema)
	c.Assert(len(tbs), Equals, 1)

	err = dd.DropSchema(ctx, noExist)
	c.Assert(errors2.ErrorEqual(err, ddl.ErrNotExists), IsTrue)
	err = dd.DropSchema(ctx, tbIdent.Schema)
	c.Assert(err, IsNil)
}

func statement(sql string) stmt.Statement {
	lexer := parser.NewLexer(sql)
	parser.YYParse(lexer)
	return lexer.Stmts()[0]
}
