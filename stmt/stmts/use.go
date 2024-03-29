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

package stmts

import (
	"github.com/Dong-Chan/alloydb/context"
	"github.com/Dong-Chan/alloydb/model"
	"github.com/Dong-Chan/alloydb/rset"
	"github.com/Dong-Chan/alloydb/sessionctx"
	"github.com/Dong-Chan/alloydb/sessionctx/db"
	"github.com/Dong-Chan/alloydb/stmt"
	"github.com/Dong-Chan/alloydb/util/errors"
	"github.com/Dong-Chan/alloydb/util/format"
)

var _ stmt.Statement = (*UseStmt)(nil)

// UseStmt is a statement to use the DBName database as the current database.
// See: https://dev.mysql.com/doc/refman/5.7/en/use.html
type UseStmt struct {
	DBName string

	Text string
}

// Explain implements the stmt.Statement Explain interface.
func (s *UseStmt) Explain(ctx context.Context, w format.Formatter) {
	w.Format("%s\n", s.Text)
}

// IsDDL implements the stmt.Statement IsDDL interface.
func (s *UseStmt) IsDDL() bool {
	return false
}

// OriginText implements the stmt.Statement OriginText interface.
func (s *UseStmt) OriginText() string {
	return s.Text
}

// SetText implements the stmt.Statement SetText interface.
func (s *UseStmt) SetText(text string) {
	s.Text = text
}

// Exec implements the stmt.Statement Exec interface.
func (s *UseStmt) Exec(ctx context.Context) (_ rset.Recordset, err error) {
	dbname := model.NewCIStr(s.DBName)
	if !sessionctx.GetDomain(ctx).InfoSchema().SchemaExists(dbname) {
		return nil, errors.ErrDatabaseNotExist
	}
	db.BindCurrentSchema(ctx, dbname.O)
	return nil, nil
}
