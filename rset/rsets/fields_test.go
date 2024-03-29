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

package rsets

import (
	. "github.com/pingcap/check"
	"github.com/Dong-Chan/alloydb/expression/expressions"
	"github.com/Dong-Chan/alloydb/field"
	"github.com/Dong-Chan/alloydb/model"
	"github.com/Dong-Chan/alloydb/plan/plans"
	"github.com/Dong-Chan/alloydb/table/tables"
)

var _ = Suite(&testSelectFieldsRsetSuite{})

type testSelectFieldsRsetSuite struct {
	sr *SelectFieldsRset
	fr *FieldRset
}

func (s *testSelectFieldsRsetSuite) SetUpSuite(c *C) {
	names := []string{"id", "name"}
	tblPlan := newTestTablePlan(testData, names)
	resultFields := tblPlan.GetFields()

	fields := make([]*field.Field, len(resultFields))
	for i, resultField := range resultFields {
		name := resultField.Name
		fields[i] = &field.Field{Expr: &expressions.Ident{CIStr: model.NewCIStr(name)}, Name: name}
	}

	selectList := &plans.SelectList{
		HiddenFieldOffset: len(resultFields),
		ResultFields:      resultFields,
		Fields:            fields,
	}

	s.sr = &SelectFieldsRset{Src: tblPlan, SelectList: selectList}
	s.fr = &FieldRset{Fields: fields}
}

func (s *testSelectFieldsRsetSuite) TestDistinctRsetPlan(c *C) {
	// check src plan, like `select c1, c2 from t`.
	p, err := s.sr.Plan(nil)
	c.Assert(err, IsNil)

	_, ok := p.(*testTablePlan)
	c.Assert(ok, IsTrue)

	// check SelectFieldsDefaultPlan, like `select c1, 1 from t`.
	fld := &field.Field{Expr: expressions.Value{Val: 1}}
	s.sr.SelectList.Fields[0] = fld

	p, err = s.sr.Plan(nil)
	c.Assert(err, IsNil)

	_, ok = p.(*plans.SelectFieldsDefaultPlan)
	c.Assert(ok, IsTrue)

	oldFld := s.sr.SelectList.Fields[1]
	s.sr.SelectList.Fields = s.sr.SelectList.Fields[:1]

	p, err = s.sr.Plan(nil)
	c.Assert(err, IsNil)

	_, ok = p.(*plans.SelectFieldsDefaultPlan)
	c.Assert(ok, IsTrue)

	// check src plan set TableNilPlan, like `select 1 from t`.
	tdp := &plans.TableDefaultPlan{T: &tables.Table{}, Fields: s.sr.SelectList.ResultFields}
	s.sr.Src = tdp

	s.sr.SelectList.Fields = []*field.Field{fld}

	p, err = s.sr.Plan(nil)
	c.Assert(err, IsNil)

	sp, ok := p.(*plans.SelectFieldsDefaultPlan)
	c.Assert(ok, IsTrue)
	c.Assert(sp, NotNil)

	_, ok = sp.Src.(*plans.TableNilPlan)
	c.Assert(ok, IsTrue)

	// cover isConst check, like `select 1, c1 from t`
	s.sr.SelectList.Fields = []*field.Field{fld, oldFld}
	p, err = s.sr.Plan(nil)
	c.Assert(err, IsNil)

	_, ok = sp.Src.(*plans.TableNilPlan)
	c.Assert(ok, IsTrue)
}

func (s *testSelectFieldsRsetSuite) TestFieldRsetPlan(c *C) {
	p, err := s.fr.Plan(nil)
	c.Assert(err, IsNil)

	_, ok := p.(*plans.SelectEmptyFieldListPlan)
	c.Assert(ok, IsTrue)
}
