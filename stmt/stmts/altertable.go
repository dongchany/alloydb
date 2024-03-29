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
	"github.com/Dong-Chan/alloydb/ddl"
	"github.com/Dong-Chan/alloydb/rset"
	"github.com/Dong-Chan/alloydb/sessionctx"
	"github.com/Dong-Chan/alloydb/stmt"
	"github.com/Dong-Chan/alloydb/table"
	"github.com/Dong-Chan/alloydb/util/format"
)

var _ stmt.Statement = (*AlterTableStmt)(nil)

// AlterTableStmt is a statement to change the structure of a table.
// See: https://dev.mysql.com/doc/refman/5.7/en/alter-table.html
type AlterTableStmt struct {
	Ident table.Ident
	Specs []*ddl.AlterSpecification

	Text string
}

// Explain implements the stmt.Statement Explain interface.
func (s *AlterTableStmt) Explain(ctx context.Context, w format.Formatter) {
	w.Format("%s\n", s.Text)
}

// IsDDL implements the stmt.Statement IsDDL interface.
func (s *AlterTableStmt) IsDDL() bool {
	return true
}

// OriginText implements the stmt.Statement OriginText interface.
func (s *AlterTableStmt) OriginText() string {
	return s.Text
}

// SetText implements the stmt.Statement SetText interface.
func (s *AlterTableStmt) SetText(text string) {
	s.Text = text
}

// Exec implements the stmt.Statement Exec interface.
func (s *AlterTableStmt) Exec(ctx context.Context) (_ rset.Recordset, err error) {
	err = sessionctx.GetDomain(ctx).DDL().AlterTable(ctx, s.Ident.Full(ctx), s.Specs)
	return nil, err
}
