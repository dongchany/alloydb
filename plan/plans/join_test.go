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
	. "github.com/pingcap/check"
	"github.com/Dong-Chan/alloydb/column"
	"github.com/Dong-Chan/alloydb/expression/expressions"
	"github.com/Dong-Chan/alloydb/field"
	"github.com/Dong-Chan/alloydb/model"
	mysql "github.com/Dong-Chan/alloydb/mysqldef"
	"github.com/Dong-Chan/alloydb/parser/opcode"
	"github.com/Dong-Chan/alloydb/util/types"
)

type testJoinSuit struct {
	cols []*column.Col
}

var _ = Suite(&testJoinSuit{})

func (s *testJoinSuit) TestJoin(c *C) {
	s.cols = []*column.Col{
		&column.Col{
			ColumnInfo: model.ColumnInfo{
				ID:           0,
				Name:         model.NewCIStr("id"),
				Offset:       0,
				DefaultValue: 0,
				FieldType:    *types.NewFieldType(mysql.TypeLonglong),
			},
		},
		&column.Col{
			ColumnInfo: model.ColumnInfo{
				ID:           1,
				Name:         model.NewCIStr("name"),
				Offset:       1,
				DefaultValue: nil,
				FieldType:    *types.NewFieldType(mysql.TypeVarchar),
			},
		},
	}

	var testData1 = []*testRowData{
		&testRowData{1, []interface{}{10, "10"}},
		&testRowData{2, []interface{}{10, "20"}},
		&testRowData{3, []interface{}{10, "30"}},
		&testRowData{4, []interface{}{40, "40"}},
		&testRowData{6, []interface{}{60, "60"}},
	}

	var testData2 = []*testRowData{
		&testRowData{1, []interface{}{10, "10"}},
		&testRowData{2, []interface{}{10, "20"}},
		&testRowData{3, []interface{}{10, "30"}},
		&testRowData{4, []interface{}{40, "40"}},
		&testRowData{6, []interface{}{60, "60"}},
	}

	tblPlan1 := &testTablePlan{testData1, []string{"id", "name"}}
	tblPlan2 := &testTablePlan{testData2, []string{"id", "name"}}

	joinPlan := &JoinPlan{
		Left:   tblPlan1,
		Right:  tblPlan2,
		Type:   "CROSS",
		Fields: []*field.ResultField{},
		On:     expressions.Value{true},
	}

	joinPlan.Do(nil, func(id interface{}, data []interface{}) (bool, error) {
		return true, nil
	})

	joinPlan = &JoinPlan{
		Left:   tblPlan1,
		Right:  tblPlan2,
		Type:   "LEFT",
		Fields: []*field.ResultField{},
		On:     expressions.Value{true},
	}

	joinPlan.Do(nil, func(id interface{}, data []interface{}) (bool, error) {
		return true, nil
	})

	joinPlan = &JoinPlan{
		Left:   tblPlan1,
		Right:  tblPlan2,
		Type:   "RIGHT",
		Fields: []*field.ResultField{},
		On:     expressions.Value{true},
	}

	expr := &expressions.BinaryOperation{
		Op: opcode.LT,
		L: &expressions.Ident{
			CIStr: model.NewCIStr("id"),
		},
		R: expressions.Value{
			Val: 100,
		},
	}
	np, _, err := joinPlan.Filter(nil, expr)
	c.Assert(np, NotNil)
	c.Assert(err, IsNil)

	joinPlan.Do(nil, func(id interface{}, data []interface{}) (bool, error) {
		return true, nil
	})

	joinPlan = &JoinPlan{
		Left:   tblPlan1,
		Right:  tblPlan2,
		Type:   "FULL",
		Fields: []*field.ResultField{},
		On:     expressions.Value{true},
	}

	joinPlan.Do(nil, func(id interface{}, data []interface{}) (bool, error) {
		return true, nil
	})

}
