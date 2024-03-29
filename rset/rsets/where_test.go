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
	"github.com/Dong-Chan/alloydb/model"
	"github.com/Dong-Chan/alloydb/parser/opcode"
	"github.com/Dong-Chan/alloydb/plan/plans"
)

var _ = Suite(&testWhereRsetSuite{})

type testWhereRsetSuite struct {
	r *WhereRset
}

func (s *testWhereRsetSuite) SetUpSuite(c *C) {
	names := []string{"id", "name"}
	tblPlan := newTestTablePlan(testData, names)

	expr := expressions.NewBinaryOperation(opcode.Plus, &expressions.Ident{model.NewCIStr("id")}, expressions.Value{1})

	s.r = &WhereRset{Src: tblPlan, Expr: expr}
}

func (s *testWhereRsetSuite) TestShowRsetPlan(c *C) {
	// `select id, name from t where id + 1`
	p, err := s.r.Plan(nil)
	c.Assert(err, IsNil)

	_, ok := p.(*plans.FilterDefaultPlan)
	c.Assert(ok, IsTrue)

	// `select id, name from t where 1`
	s.r.Expr = expressions.Value{int64(1)}

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	// `select id, name from t where 0`
	s.r.Expr = expressions.Value{int64(0)}

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	// `select id, name from t where null`
	s.r.Expr = expressions.Value{}

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	// `select id, name from t where id < 10`
	expr := expressions.NewBinaryOperation(opcode.LT, &expressions.Ident{model.NewCIStr("id")}, expressions.Value{10})

	s.r.Expr = expr

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	src := s.r.Src.(*testTablePlan)
	src.SetFilter(true)

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	// `select id, name from t where null && 1`
	src.SetFilter(false)

	expr = expressions.NewBinaryOperation(opcode.AndAnd, expressions.Value{}, expressions.Value{1})

	s.r.Expr = expr

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	// `select id, name from t where id && 1`
	expr = expressions.NewBinaryOperation(opcode.AndAnd, &expressions.Ident{model.NewCIStr("id")}, expressions.Value{1})

	s.r.Expr = expr

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	// `select id, name from t where id && (id < 10)`
	exprx := expressions.NewBinaryOperation(opcode.LT, &expressions.Ident{model.NewCIStr("id")}, expressions.Value{10})
	expr = expressions.NewBinaryOperation(opcode.AndAnd, &expressions.Ident{model.NewCIStr("id")}, exprx)

	s.r.Expr = expr

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	src.SetFilter(true)

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	// `select id, name from t where abc`
	src.SetFilter(false)

	expr = &expressions.Ident{model.NewCIStr("abc")}

	s.r.Expr = expr

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	src.SetFilter(true)

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	// `select id, name from t where 1 is null`
	src.SetFilter(false)

	exprx = expressions.Value{1}
	expr = &expressions.IsNull{Expr: exprx}

	s.r.Expr = expr

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	src.SetFilter(true)

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	// `select id, name from t where id is null`
	src.SetFilter(false)

	exprx = &expressions.Ident{model.NewCIStr("id")}
	expr = &expressions.IsNull{Expr: exprx}

	s.r.Expr = expr

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	src.SetFilter(true)

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	// `select id, name from t where +id`
	src.SetFilter(false)

	exprx = &expressions.Ident{model.NewCIStr("id")}
	expr = expressions.NewUnaryOperation(opcode.Plus, exprx)

	s.r.Expr = expr

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	src.SetFilter(true)

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	// `select id, name from t where id in (id)`
	expr = &expressions.PatternIn{Expr: exprx, List: []expression.Expression{exprx}}
	s.r.Expr = expr

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	// `select id, name from t where id like '%s'`
	expry := expressions.Value{"%s"}
	expr = &expressions.PatternLike{Expr: exprx, Pattern: expry}
	s.r.Expr = expr

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)

	// `select id, name from t where id is true`
	expr = &expressions.IsTruth{Expr: exprx}
	s.r.Expr = expr

	_, err = s.r.Plan(nil)
	c.Assert(err, IsNil)
}
