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

package stmts_test

import (
	. "github.com/pingcap/check"
	"github.com/Dong-Chan/alloydb"
	"github.com/Dong-Chan/alloydb/stmt/stmts"
)

func (s *testStmtSuite) TestUnion(c *C) {
	testSQL := `select 1 union select 0;`
	mustExec(c, s.testDB, testSQL)

	stmtList, err := alloydb.Compile(testSQL)
	c.Assert(err, IsNil)
	c.Assert(stmtList, HasLen, 1)

	testStmt, ok := stmtList[0].(*stmts.UnionStmt)
	c.Assert(ok, IsTrue)

	c.Assert(testStmt.IsDDL(), IsFalse)
	c.Assert(len(testStmt.OriginText()), Greater, 0)

	mf := newMockFormatter()
	testStmt.Explain(nil, mf)
	c.Assert(mf.Len(), Greater, 0)

	testSQL = `drop table if exists union_test; create table union_test(id int);
    insert union_test values (1),(2); select id from union_test union select 1;`
	mustExec(c, s.testDB, testSQL)

	testSQL = `select id from union_test union select id from union_test;`
	tx := mustBegin(c, s.testDB)
	rows, err := tx.Query(testSQL)
	c.Assert(err, IsNil)

	i := 1
	for rows.Next() {
		var id int
		rows.Scan(&id)
		c.Assert(id, Equals, i)

		i++
	}

	rows.Close()
	mustCommit(c, tx)
}
