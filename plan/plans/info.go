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
	"github.com/Dong-Chan/alloydb/infoschema"
	"github.com/Dong-Chan/alloydb/model"
	mysql "github.com/Dong-Chan/alloydb/mysqldef"
	"github.com/Dong-Chan/alloydb/plan"
	"github.com/Dong-Chan/alloydb/sessionctx"
	"github.com/Dong-Chan/alloydb/util/charset"
	"github.com/Dong-Chan/alloydb/util/format"
	"github.com/Dong-Chan/alloydb/util/types"
)

var _ = (*InfoSchemaPlan)(nil)

// InfoSchemaPlan handles information_schema query, simulates the behavior of
// MySQL.
type InfoSchemaPlan struct {
	TableName string
}

var (
	schemataFields       = buildResultFieldsForSchemata()
	tablesFields         = buildResultFieldsForTables()
	columnsFields        = buildResultFieldsForColumns()
	statisticsFields     = buildResultFieldsForStatistics()
	characterSetsFields  = buildResultFieldsForCharacterSets()
	characterSetsRecords = buildCharacterSetsRecords()
)

const (
	tableSchemata      = "SCHEMATA"
	tableTables        = "TABLES"
	tableColumns       = "COLUMNS"
	tableStatistics    = "STATISTICS"
	tableCharacterSets = "CHARACTER_SETS"
	catalogVal         = "def"
)

// NewInfoSchemaPlan returns new InfoSchemaPlan instance, and checks if the
// given table name is valid.
func NewInfoSchemaPlan(tableName string) (isp *InfoSchemaPlan, err error) {
	switch strings.ToUpper(tableName) {
	case tableSchemata:
	case tableTables:
	case tableColumns:
	case tableStatistics:
	case tableCharacterSets:
	default:
		return nil, errors.Errorf("table INFORMATION_SCHEMA.%s does not exist", tableName)
	}
	isp = &InfoSchemaPlan{
		TableName: strings.ToUpper(tableName),
	}
	return
}

func buildResultFieldsForSchemata() (rfs []*field.ResultField) {
	tbName := tableSchemata
	rfs = append(rfs, buildResultField(tbName, "CATALOG_NAME", mysql.TypeVarchar, 512))
	rfs = append(rfs, buildResultField(tbName, "SCHEMA_NAME", mysql.TypeVarchar, 64))
	rfs = append(rfs, buildResultField(tbName, "DEFAULT_CHARACTER_SET_NAME", mysql.TypeVarchar, 64))
	rfs = append(rfs, buildResultField(tbName, "DEFAULT_COLLATION_NAME", mysql.TypeVarchar, 32))
	rfs = append(rfs, buildResultField(tbName, "SQL_PATH", mysql.TypeVarchar, 512))
	return rfs
}

func (isp *InfoSchemaPlan) doSchemata(schemas []string, iterFunc plan.RowIterFunc) error {
	sort.Strings(schemas)
	for _, schema := range schemas {
		record := []interface{}{
			catalogVal,                 // CATALOG_NAME
			schema,                     // SCHEMA_NAME
			mysql.DefaultCharset,       // DEFAULT_CHARACTER_SET_NAME
			mysql.DefaultCollationName, // DEFAULT_COLLATION_NAME
			nil,
		}
		if more, err := iterFunc(0, record); !more || err != nil {
			return err
		}
	}
	return nil
}

func buildResultFieldsForTables() (rfs []*field.ResultField) {
	tbName := tableTables
	rfs = append(rfs, buildResultField(tbName, "TABLE_CATALOG", mysql.TypeVarchar, 512))
	rfs = append(rfs, buildResultField(tbName, "TABLE_SCHEMA", mysql.TypeVarchar, 64))
	rfs = append(rfs, buildResultField(tbName, "TABLE_NAME", mysql.TypeVarchar, 64))
	rfs = append(rfs, buildResultField(tbName, "TABLE_TYPE", mysql.TypeVarchar, 64))
	rfs = append(rfs, buildResultField(tbName, "ENGINE", mysql.TypeVarchar, 64))
	rfs = append(rfs, buildResultField(tbName, "VERSION", mysql.TypeLonglong, 21))
	rfs = append(rfs, buildResultField(tbName, "ROW_FORMAT", mysql.TypeVarchar, 10))
	rfs = append(rfs, buildResultField(tbName, "TABLE_ROWS", mysql.TypeLonglong, 21))
	rfs = append(rfs, buildResultField(tbName, "AVG_ROW_LENGTH", mysql.TypeLonglong, 21))
	rfs = append(rfs, buildResultField(tbName, "DATA_LENGTH", mysql.TypeLonglong, 21))
	rfs = append(rfs, buildResultField(tbName, "MAX_DATA_LENGTH", mysql.TypeLonglong, 21))
	rfs = append(rfs, buildResultField(tbName, "INDEX_LENGTH", mysql.TypeLonglong, 21))
	rfs = append(rfs, buildResultField(tbName, "DATA_FREE", mysql.TypeLonglong, 21))
	rfs = append(rfs, buildResultField(tbName, "AUTO_INCREMENT", mysql.TypeLonglong, 21))
	rfs = append(rfs, buildResultField(tbName, "CREATE_TIME", mysql.TypeDatetime, 19))
	rfs = append(rfs, buildResultField(tbName, "UPDATE_TIME", mysql.TypeDatetime, 19))
	rfs = append(rfs, buildResultField(tbName, "CHECK_TIME", mysql.TypeDatetime, 19))
	rfs = append(rfs, buildResultField(tbName, "TABLE_COLLATION", mysql.TypeVarchar, 32))
	rfs = append(rfs, buildResultField(tbName, "CHECK_SUM", mysql.TypeLonglong, 21))
	rfs = append(rfs, buildResultField(tbName, "CREATE_OPTIONS", mysql.TypeVarchar, 255))
	rfs = append(rfs, buildResultField(tbName, "TABLE_COMMENT", mysql.TypeVarchar, 2048))
	for i, f := range rfs {
		f.Offset = i
	}
	return
}

func (isp *InfoSchemaPlan) doTables(schemas []*model.DBInfo, iterFunc plan.RowIterFunc) error {
	for _, schema := range schemas {
		for _, table := range schema.Tables {
			record := []interface{}{
				catalogVal,          // TABLE_CATALOG
				schema.Name.O,       // TABLE_SCHEMA
				table.Name.O,        // TABLE_NAME
				"BASE_TABLE",        // TABLE_TYPE
				"InnoDB",            // ENGINE
				uint64(10),          // VERSION
				"Compact",           // ROW_FORMAT
				uint64(0),           // TABLE_ROWS
				uint64(0),           // AVG_ROW_LENGTH
				uint64(16384),       // DATA_LENGTH
				uint64(0),           // MAX_DATA_LENGTH
				uint64(0),           // INDEX_LENGTH
				uint64(0),           // DATA_FREE
				nil,                 // AUTO_INCREMENT
				nil,                 // CREATE_TIME
				nil,                 // UPDATE_TIME
				nil,                 // CHECK_TIME
				"latin1_swedish_ci", // TABLE_COLLATION
				nil,                 // CHECKSUM
				"",                  // CREATE_OPTIONS
				"",                  // TABLE_COMMENT
			}
			if more, err := iterFunc(0, record); !more || err != nil {
				return err
			}
		}
	}
	return nil
}

func buildResultFieldsForColumns() (rfs []*field.ResultField) {
	tbName := tableColumns
	rfs = append(rfs, buildResultField(tbName, "TABLE_CATALOG", mysql.TypeVarchar, 512))
	rfs = append(rfs, buildResultField(tbName, "TABLE_SCHEMA", mysql.TypeVarchar, 64))
	rfs = append(rfs, buildResultField(tbName, "TABLE_NAME", mysql.TypeVarchar, 64))
	rfs = append(rfs, buildResultField(tbName, "COLUMN_NAME", mysql.TypeVarchar, 64))
	rfs = append(rfs, buildResultField(tbName, "ORIGINAL_POSITION", mysql.TypeLonglong, 64))
	rfs = append(rfs, buildResultField(tbName, "COLUMN_DEFAULT", mysql.TypeBlob, 196606))
	rfs = append(rfs, buildResultField(tbName, "IS_NULLABLE", mysql.TypeVarchar, 3))
	rfs = append(rfs, buildResultField(tbName, "DATA_TYPE", mysql.TypeVarchar, 64))
	rfs = append(rfs, buildResultField(tbName, "CHARACTER_MAXIMUM_LENGTH", mysql.TypeLonglong, 21))
	rfs = append(rfs, buildResultField(tbName, "CHARACTOR_OCTET_LENGTH", mysql.TypeLonglong, 21))
	rfs = append(rfs, buildResultField(tbName, "NUMERIC_PRECISION", mysql.TypeLonglong, 21))
	rfs = append(rfs, buildResultField(tbName, "NUMERIC_SCALE", mysql.TypeLonglong, 21))
	rfs = append(rfs, buildResultField(tbName, "DATETIME_PRECISION", mysql.TypeLonglong, 21))
	rfs = append(rfs, buildResultField(tbName, "CHARACTER_SET_NAME", mysql.TypeVarchar, 32))
	rfs = append(rfs, buildResultField(tbName, "COLLATION_NAME", mysql.TypeVarchar, 32))
	rfs = append(rfs, buildResultField(tbName, "COLUMN_TYPE", mysql.TypeBlob, 196606))
	rfs = append(rfs, buildResultField(tbName, "COLUMN_KEY", mysql.TypeVarchar, 3))
	rfs = append(rfs, buildResultField(tbName, "EXTRA", mysql.TypeVarchar, 30))
	rfs = append(rfs, buildResultField(tbName, "PRIVILEGES", mysql.TypeVarchar, 80))
	rfs = append(rfs, buildResultField(tbName, "COLUMN_COMMENT", mysql.TypeVarchar, 1024))
	for i, f := range rfs {
		f.Offset = i
	}
	return
}

func (isp *InfoSchemaPlan) doColumns(schemas []*model.DBInfo, iterFunc plan.RowIterFunc) error {
	for _, schema := range schemas {
		for _, table := range schema.Tables {
			for i, col := range table.Columns {
				colLen := col.Flen
				if colLen == types.UnspecifiedLength {
					colLen = mysql.GetDefaultFieldLength(col.Tp)
				}
				decimal := col.Decimal
				if decimal == types.UnspecifiedLength {
					decimal = 0
				}
				dataType := types.TypeToStr(col.Tp, col.Charset == charset.CharsetBin)
				columnType := fmt.Sprintf("%s(%d)", dataType, colLen)
				columnDesc := column.NewColDesc(&column.Col{ColumnInfo: *col})
				var columnDefault interface{}
				if columnDesc.DefaultValue != nil {
					columnDefault = fmt.Sprintf("%v", columnDesc.DefaultValue)
				}
				record := []interface{}{
					catalogVal,                                                 // TABLE_CATALOG
					schema.Name.O,                                              // TABLE_SCHEMA
					table.Name.O,                                               // TABLE_NAME
					col.Name.O,                                                 // COLUMN_NAME
					i + 1,                                                      // ORIGINAL_POSITION
					columnDefault,                                              // COLUMN_DEFAULT
					columnDesc.Null,                                            // IS_NULLABLE
					types.TypeToStr(col.Tp, col.Charset == charset.CharsetBin), // DATA_TYPE
					colLen,                            // CHARACTER_MAXIMUM_LENGTH
					colLen,                            // CHARACTOR_OCTET_LENGTH
					decimal,                           // NUMERIC_PRECISION
					0,                                 // NUMERIC_SCALE
					0,                                 // DATETIME_PRECISION
					col.Charset,                       // CHARACTER_SET_NAME
					col.Collate,                       // COLLATION_NAME
					columnType,                        // COLUMN_TYPE
					columnDesc.Key,                    // COLUMN_KEY
					columnDesc.Extra,                  // EXTRA
					"select,insert,update,references", // PRIVILEGES
					"", // COLUMN_COMMENT
				}
				if more, err := iterFunc(0, record); !more || err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func buildResultFieldsForStatistics() (rfs []*field.ResultField) {
	tbName := tableStatistics
	rfs = append(rfs, buildResultField(tbName, "TABLE_CATALOG", mysql.TypeVarchar, 512))
	rfs = append(rfs, buildResultField(tbName, "TABLE_SCHEMA", mysql.TypeVarchar, 64))
	rfs = append(rfs, buildResultField(tbName, "TABLE_NAME", mysql.TypeVarchar, 64))
	rfs = append(rfs, buildResultField(tbName, "NON_UNIQUE", mysql.TypeVarchar, 1))
	rfs = append(rfs, buildResultField(tbName, "INDEX_SCHEMA", mysql.TypeVarchar, 64))
	rfs = append(rfs, buildResultField(tbName, "INDEX_NAME", mysql.TypeVarchar, 64))
	rfs = append(rfs, buildResultField(tbName, "SEQ_IN_INDEX", mysql.TypeLonglong, 2))
	rfs = append(rfs, buildResultField(tbName, "COLUMN_NAME", mysql.TypeVarchar, 21))
	rfs = append(rfs, buildResultField(tbName, "COLLATION", mysql.TypeVarchar, 1))
	rfs = append(rfs, buildResultField(tbName, "CARDINALITY", mysql.TypeLonglong, 21))
	rfs = append(rfs, buildResultField(tbName, "SUB_PART", mysql.TypeLonglong, 3))
	rfs = append(rfs, buildResultField(tbName, "PACKED", mysql.TypeVarchar, 10))
	rfs = append(rfs, buildResultField(tbName, "NULLABLE", mysql.TypeVarchar, 3))
	rfs = append(rfs, buildResultField(tbName, "INDEX_TYPE", mysql.TypeVarchar, 16))
	rfs = append(rfs, buildResultField(tbName, "COMMENT", mysql.TypeVarchar, 16))
	rfs = append(rfs, buildResultField(tbName, "INDEX_COMMENT", mysql.TypeVarchar, 1024))
	for i, f := range rfs {
		f.Offset = i
	}
	return
}

func (isp *InfoSchemaPlan) doStatistics(is infoschema.InfoSchema, schemas []*model.DBInfo, iterFunc plan.RowIterFunc) error {
	for _, schema := range schemas {
		for _, table := range schema.Tables {
			for _, index := range table.Indices {
				nonUnique := "1"
				if index.Unique {
					nonUnique = "0"
				}
				for i, key := range index.Columns {
					col, _ := is.ColumnByName(schema.Name, table.Name, key.Name)
					nullable := "YES"
					if mysql.HasNotNullFlag(col.Flag) {
						nullable = ""
					}
					record := []interface{}{
						catalogVal,    // TABLE_CATALOG
						schema.Name.O, // TABLE_SCHEMA
						table.Name.O,  // TABLE_NAME
						nonUnique,     // NON_UNIQUE
						schema.Name.O, // INDEX_SCHEMA
						index.Name.O,  // INDEX_NAME
						i + 1,         // SEQ_IN_INDEX
						key.Name.O,    // COLUMN_NAME
						"A",           // COLLATION
						0,             // CARDINALITY
						nil,           // SUB_PART
						nil,           // PACKED
						nullable,      // NULLABLE
						"BTREE",       // INDEX_TYPE
						"",            // COMMENT
						"",            // INDEX_COMMENT
					}
					if more, err := iterFunc(0, record); !more || err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func buildResultFieldsForCharacterSets() (rfs []*field.ResultField) {
	tbName := tableCharacterSets
	rfs = append(rfs, buildResultField(tbName, "CHARACTER_SET_NAME", mysql.TypeVarchar, 32))
	rfs = append(rfs, buildResultField(tbName, "DEFAULT_COLLATE_NAME", mysql.TypeVarchar, 32))
	rfs = append(rfs, buildResultField(tbName, "DESCRIPTION", mysql.TypeVarchar, 60))
	rfs = append(rfs, buildResultField(tbName, "MAXLEN", mysql.TypeLonglong, 3))
	return rfs
}

func buildCharacterSetsRecords() (records [][]interface{}) {
	records = append(records,
		[]interface{}{"ascii", "ascii_general_ci", "US ASCII", 1},
		[]interface{}{"binary", "binary", "Binary pseudo charset", 1},
		[]interface{}{"latin1", "latin1_swedish_ci", "cp1252 West European", 1},
		[]interface{}{"utf8", "utf8_general_ci", "UTF-8 Unicode", 3},
		[]interface{}{"utf8mb4", "utf8mb4_general_ci", "UTF-8 Unicode", 4},
	)
	return records
}

func (isp *InfoSchemaPlan) doCharacterSets(iterFunc plan.RowIterFunc) error {
	for _, record := range characterSetsRecords {
		if more, err := iterFunc(0, record); !more || err != nil {
			return err
		}
	}
	return nil
}

// Do implements plan.Plan Do interface, constructs result data.
func (isp *InfoSchemaPlan) Do(ctx context.Context, iterFunc plan.RowIterFunc) error {
	is := sessionctx.GetDomain(ctx).InfoSchema()
	schemas := is.AllSchemas()
	switch isp.TableName {
	case tableSchemata:
		return isp.doSchemata(is.AllSchemaNames(), iterFunc)
	case tableTables:
		return isp.doTables(schemas, iterFunc)
	case tableColumns:
		return isp.doColumns(schemas, iterFunc)
	case tableStatistics:
		return isp.doStatistics(is, schemas, iterFunc)
	case tableCharacterSets:
		return isp.doCharacterSets(iterFunc)
	}
	return nil
}

// Explain implements plan.Plan Explain interface.
func (isp *InfoSchemaPlan) Explain(w format.Formatter) {}

// Filter implements plan.Plan Filter interface.
func (isp *InfoSchemaPlan) Filter(ctx context.Context, expr expression.Expression) (p plan.Plan, filtered bool, err error) {
	return isp, false, nil
}

// GetFields implements plan.Plan GetFields interface, simulates MySQL's output.
func (isp *InfoSchemaPlan) GetFields() []*field.ResultField {
	switch isp.TableName {
	case tableSchemata:
		return schemataFields
	case tableTables:
		return tablesFields
	case tableColumns:
		return columnsFields
	case tableStatistics:
		return statisticsFields
	case tableCharacterSets:
		return characterSetsFields
	}
	return nil
}

func buildResultField(tableName, name string, tp byte, size int) *field.ResultField {
	mCharset := charset.CharsetBin
	mCollation := charset.CharsetBin
	mFlag := mysql.UnsignedFlag
	if tp == mysql.TypeVarchar || tp == mysql.TypeBlob {
		mCharset = mysql.DefaultCharset
		mCollation = mysql.DefaultCollationName
		mFlag = 0
	}
	fieldType := types.FieldType{
		Charset: mCharset,
		Collate: mCollation,
		Tp:      tp,
		Flen:    size,
		Flag:    uint(mFlag),
	}
	colInfo := model.ColumnInfo{
		Name:      model.NewCIStr(name),
		FieldType: fieldType,
	}
	field := &field.ResultField{
		Col:       column.Col{ColumnInfo: colInfo},
		DBName:    infoschema.Name,
		TableName: tableName,
		Name:      colInfo.Name.O,
	}
	return field
}
