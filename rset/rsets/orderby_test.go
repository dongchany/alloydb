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
	"github.com/Dong-Chan/alloydb/expression"
	"github.com/Dong-Chan/alloydb/expression/expressions"
	"github.com/Dong-Chan/alloydb/field"
	"github.com/Dong-Chan/alloydb/model"
	"github.com/Dong-Chan/alloydb/plan/plans"
)

var _ = Suite(&testOrderByRsetSuite{})

type testOrderByRsetSuite struct {
	r *OrderByRset
}

func (s *testOrderByRsetSuite) SetUpSuite(c *C) {
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

	s.r = &OrderByRset{Src: tblPlan, SelectList: selectList}
}

func (s *testOrderByRsetSuite) TestOrderByRsetCheckAndUpdateSelectList(c *C) {
	resultFields := s.r.Src.GetFields()

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

	expr := &expressions.Ident{model.NewCIStr("id")}
	orderByItem := OrderByItem{Expr: expr, Asc: true}
	by := []OrderByItem{orderByItem}
	r := &OrderByRset{By: by, SelectList: selectList}

	// `select id, name from t order by id`
	err := r.CheckAndUpdateSelectList(selectList, resultFields)
	c.Assert(err, IsNil)

	// `select id, name as id from t order by id`
	selectList.Fields[1].Name = "id"
	selectList.ResultFields[1].Name = "id"

	err = r.CheckAndUpdateSelectList(selectList, resultFields)
	c.Assert(err, NotNil)

	// `select id, name from t order by count(1) > 1`
	aggExpr, err := expressions.NewCall("count", []expression.Expression{expressions.Value{1}}, false)
	c.Assert(err, IsNil)

	r.By[0].Expr = aggExpr

	err = r.CheckAndUpdateSelectList(selectList, resultFields)
	c.Assert(err, IsNil)

	// `select id, name from t order by count(xxx) > 1`
	aggExpr, err = expressions.NewCall("count", []expression.Expression{&expressions.Ident{model.NewCIStr("xxx")}}, false)
	c.Assert(err, IsNil)

	r.By[0].Expr = aggExpr

	err = r.CheckAndUpdateSelectList(selectList, resultFields)
	c.Assert(err, NotNil)

	// `select id, name from t order by xxx`
	r.By[0].Expr = &expressions.Ident{model.NewCIStr("xxx")}
	selectList.Fields[1].Name = "name"
	selectList.ResultFields[1].Name = "name"

	err = r.CheckAndUpdateSelectList(selectList, resultFields)
	c.Assert(err, NotNil)
}

func (s *testOrderByRsetSuite) TestOrderByRsetPlan(c *C) {
	// `select id, name from t`
	_, err := s.r.Plan(nil)
	c.Assert(err, IsNil)

	// `select id, name from t order by id`
	expr := &expressions.Ident{model.NewCIStr("id")}
	orderByItem := OrderByItem{Expr: expr}
	by := []OrderByItem{orderByItem}

	s.r.By = by

	p, err := s.r.Plan(nil)
	c.Assert(err, IsNil)

	_, ok := p.(*plans.OrderByDefaultPlan)
	c.Assert(ok, IsTrue)

	// `select id, name from t order by 1`
	s.r.By[0].Expr = expressions.Value{int64(1)}

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	s.r.By[0].Expr = expressions.Value{uint64(1)}

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	// `select id, name from t order by 0`
	s.r.By[0].Expr = expressions.Value{int64(0)}

	_, err = s.r.Plan(nil)
	c.Assert(err, NotNil)

	s.r.By[0].Expr = expressions.Value{0}

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	// `select id, name from t order by xxx`
	s.r.By[0].Expr = &expressions.Ident{model.NewCIStr("xxx")}

	_, err = s.r.Plan(nil)
	c.Assert(err, NotNil)

	// check src plan is NullPlan
	s.r.Src = &plans.NullPlan{s.r.SelectList.ResultFields}

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)
}

func (s *testOrderByRsetSuite) TestOrderByRsetString(c *C) {
	str := s.r.String()
	c.Assert(len(str), Greater, 0)

	s.r.By[0].Asc = true
	str = s.r.String()
	c.Assert(len(str), Greater, 0)

	s.r.By = nil
	str = s.r.String()
	c.Assert(len(str), Equals, 0)
}
