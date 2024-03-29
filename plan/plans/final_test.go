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
	"github.com/Dong-Chan/alloydb/field"
	"github.com/Dong-Chan/alloydb/model"
	mysql "github.com/Dong-Chan/alloydb/mysqldef"
	"github.com/Dong-Chan/alloydb/util/charset"
)

type testFinalPlan struct{}

var _ = Suite(&testFinalPlan{})

var finalTestData = []*testRowData{
	&testRowData{1, []interface{}{10, "hello", true}},
	&testRowData{2, []interface{}{10, "hello", true}},
	&testRowData{3, []interface{}{10, "hello", true}},
	&testRowData{4, []interface{}{40, "hello", true}},
	&testRowData{6, []interface{}{60, "hello", false}},
}

func (t *testFinalPlan) TestFinalPlan(c *C) {
	col1 := &column.Col{
		ColumnInfo: model.ColumnInfo{
			ID:           0,
			Name:         model.NewCIStr("id"),
			Offset:       0,
			DefaultValue: 0,
		},
	}

	col2 := &column.Col{
		ColumnInfo: model.ColumnInfo{
			ID:           1,
			Name:         model.NewCIStr("name"),
			Offset:       1,
			DefaultValue: nil,
		},
	}

	col3 := &column.Col{
		ColumnInfo: model.ColumnInfo{
			ID:           2,
			Name:         model.NewCIStr("ok"),
			Offset:       2,
			DefaultValue: false,
		},
	}

	tblPlan := &testTablePlan{finalTestData, []string{"id", "name", "ok"}}

	p := &SelectFinalPlan{
		SelectList: &SelectList{
			HiddenFieldOffset: len(tblPlan.GetFields()),
			ResultFields: []*field.ResultField{
				field.ColToResultField(col1, "t"),
				field.ColToResultField(col2, "t"),
				field.ColToResultField(col3, "t"),
			},
		},
		Src: tblPlan,
	}

	for _, rf := range p.ResultFields {
		c.Assert(rf.Col.Flag, Equals, uint(0))
		c.Assert(rf.Col.Tp, Equals, byte(0))
		c.Assert(rf.Col.Charset, Equals, "")
	}

	p.Do(nil, func(id interface{}, in []interface{}) (bool, error) {
		return true, nil
	})

	for _, rf := range p.ResultFields {
		if rf.Name == "id" {
			c.Assert(rf.Col.Tp, Equals, mysql.TypeLonglong)
			c.Assert(rf.Col.Charset, Equals, charset.CharsetBin)
		}
		if rf.Name == "name" {
			c.Assert(rf.Col.Tp, Equals, mysql.TypeVarchar)
		}
		// TODO add more type tests
	}
}
