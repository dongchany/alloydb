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

package plans

import (
	"fmt"
	"sort"
	"strings"

	"github.com/juju/errors"
	"github.com/Dong-Chan/alloydb/column"
	"github.com/Dong-Chan/alloydb/context"
	"github.com/Dong-Chan/alloydb/expression"
	"github.com/Dong-Chan/alloydb/field"
	"github.com/Dong-Chan/alloydb/model"
	"github.com/Dong-Chan/alloydb/plan"
	"github.com/Dong-Chan/alloydb/sessionctx"
	"github.com/Dong-Chan/alloydb/stmt"
	"github.com/Dong-Chan/alloydb/util/charset"
	"github.com/Dong-Chan/alloydb/util/format"
)

var (
	_ plan.Plan = (*ShowPlan)(nil)
)

// ShowPlan is used for show statements
type ShowPlan struct {
	Target     int
	DBName     string
	TableName  string
	ColumnName string
	Flag       int
	Full       bool
}

func (s *ShowPlan) isColOK(c *column.Col) bool {
	// support `desc tableName columnName`
	// TODO: columnName can be a regular
	if s.ColumnName == "" {
		return true
	}

	if strings.EqualFold(s.ColumnName, c.Name.L) {
		return true
	}

	return false
}

// Do implements plan.Plan Do interface.
func (s *ShowPlan) Do(ctx context.Context, f plan.RowIterFunc) (err error) {
	is := sessionctx.GetDomain(ctx).InfoSchema()
	dbName := model.NewCIStr(s.DBName)
	switch s.Target {
	case stmt.ShowEngines:
		f(0, []interface{}{"InnoDB", "DEFAULT", "Supports transactions, row-level locking, and foreign keys", "YES", "YES", "YES"})
	case stmt.ShowDatabases:
		dbs := sessionctx.GetDomain(ctx).InfoSchema().AllSchemaNames()

		// TODO: let information_schema be the first database
		sort.Strings(dbs)

		for _, d := range dbs {
			f(0, []interface{}{d})
		}
	case stmt.ShowTables:

		if !is.SchemaExists(dbName) {
			return errors.Errorf("Can not find DB: %s", dbName)
		}

		// sort for tables
		var tableNames []string
		for _, v := range is.SchemaTables(dbName) {
			tableNames = append(tableNames, v.TableName().L)
		}

		sort.Strings(tableNames)

		for _, v := range tableNames {
			f(0, []interface{}{v})
		}
	case stmt.ShowColumns:
		if !is.SchemaExists(dbName) {
			return errors.Errorf("Can not find DB: %s", dbName)
		}
		tbName := model.NewCIStr(s.TableName)
		tb, err := is.TableByName(dbName, tbName)
		if err != nil {
			return errors.Errorf("Can not find table: %s", s.TableName)
		}
		cols := tb.Cols()

		for _, col := range cols {
			if !s.isColOK(col) {
				continue
			}

			desc := column.NewColDesc(col)

			// The FULL keyword causes the output to include the column collation and comments,
			// as well as the privileges you have for each column.
			if s.Full {
				f(0, []interface{}{
					desc.Field,
					desc.Type,
					desc.Collation,
					desc.Null,
					desc.Key,
					desc.DefaultValue,
					desc.Extra,
					desc.Privileges,
					desc.Comment,
				})
			} else {
				f(0, []interface{}{
					desc.Field,
					desc.Type,
					desc.Null,
					desc.Key,
					desc.DefaultValue,
					desc.Extra,
				})
			}
		}
	case stmt.ShowWarnings:
		// empty result
	case stmt.ShowCharset:
		// See: http://dev.mysql.com/doc/refman/5.7/en/show-character-set.html
		descs := charset.GetAllCharsets()
		for _, desc := range descs {
			row := []interface{}{desc.Name, desc.Desc, desc.DefaultCollation, desc.Maxlen}
			f(0, row)
		}
	}
	return nil
}

// Explain implements plan.Plan Explain interface.
func (s *ShowPlan) Explain(w format.Formatter) {
	// TODO: finish this
}

// GetFields implements plan.Plan GetFields interface.
func (s *ShowPlan) GetFields() []*field.ResultField {
	var names []string

	switch s.Target {
	case stmt.ShowEngines:
		names = []string{"Engine", "Support", "Comment", "Transactions", "XA", "Savepoints"}
	case stmt.ShowDatabases:
		names = []string{"Database"}
	case stmt.ShowTables:
		names = []string{fmt.Sprintf("Tables_in_%s", s.DBName)}
	case stmt.ShowColumns:
		names = column.ColDescFieldNames(s.Full)
	case stmt.ShowWarnings:
		names = []string{"Level", "Code", "Message"}
	case stmt.ShowCharset:
		names = []string{"Charset", "Description", "Default collation", "Maxlen"}
	}

	fields := make([]*field.ResultField, 0, len(names))
	for _, name := range names {
		fields = append(fields, &field.ResultField{Name: name})
	}

	return fields
}

// Filter implements plan.Plan Filter interface.
func (s *ShowPlan) Filter(ctx context.Context, expr expression.Expression) (plan.Plan, bool, error) {
	return s, false, nil
}
