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
	"github.com/Dong-Chan/alloydb/parser/opcode"
	"github.com/Dong-Chan/alloydb/plan/plans"
)

var _ = Suite(&testHavingRsetSuite{})

type testHavingRsetSuite struct {
	r *HavingRset
}

func (s *testHavingRsetSuite) SetUpSuite(c *C) {
	names := []string{"id", "name"}
	tblPlan := newTestTablePlan(testData, names)

	// expr `id > 1`
	expr := expressions.NewBinaryOperation(opcode.GT, &expressions.Ident{model.NewCIStr("id")}, expressions.Value{1})

	s.r = &HavingRset{Src: tblPlan, Expr: expr}
}

func (s *testHavingRsetSuite) TestHavingRsetCheckAndUpdateSelectList(c *C) {
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

	groupBy := []expression.Expression{}

	// `select id, name from t having id > 1`
	err := s.r.CheckAndUpdateSelectList(selectList, groupBy, resultFields)
	c.Assert(err, IsNil)

	// `select name from t group by id having id > 1`
	selectList.ResultFields = selectList.ResultFields[1:]
	selectList.Fields = selectList.Fields[1:]

	groupBy = []expression.Expression{&expressions.Ident{model.NewCIStr("id")}}
	err = s.r.CheckAndUpdateSelectList(selectList, groupBy, resultFields)
	c.Assert(err, IsNil)

	// `select name from t group by id + 1 having id > 1`
	expr := expressions.NewBinaryOperation(opcode.Plus, &expressions.Ident{model.NewCIStr("id")}, expressions.Value{1})

	groupBy = []expression.Expression{expr}
	err = s.r.CheckAndUpdateSelectList(selectList, groupBy, resultFields)
	c.Assert(err, IsNil)

	// `select name from t group by id + 1 having count(1) > 1`
	aggExpr, err := expressions.NewCall("count", []expression.Expression{expressions.Value{1}}, false)
	c.Assert(err, IsNil)

	s.r.Expr = aggExpr

	err = s.r.CheckAndUpdateSelectList(selectList, groupBy, resultFields)
	c.Assert(err, IsNil)

	// `select name from t group by id + 1 having count(xxx) > 1`
	aggExpr, err = expressions.NewCall("count", []expression.Expression{&expressions.Ident{model.NewCIStr("xxx")}}, false)
	c.Assert(err, IsNil)

	s.r.Expr = aggExpr

	err = s.r.CheckAndUpdateSelectList(selectList, groupBy, resultFields)
	c.Assert(err, NotNil)

	// `select name from t group by id having xxx > 1`
	expr = expressions.NewBinaryOperation(opcode.GT, &expressions.Ident{model.NewCIStr("xxx")}, expressions.Value{1})

	s.r.Expr = expr

	err = s.r.CheckAndUpdateSelectList(selectList, groupBy, resultFields)
	c.Assert(err, NotNil)
}

func (s *testHavingRsetSuite) TestHavingRsetPlan(c *C) {
	p, err := s.r.Plan(nil)
	c.Assert(err, IsNil)

	_, ok := p.(*plans.HavingPlan)
	c.Assert(ok, IsTrue)
}

func (s *testHavingRsetSuite) TestHavingRsetString(c *C) {
	str := s.r.String()
	c.Assert(len(str), Greater, 0)
}
