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

package expressions

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/juju/errors"
	"github.com/Dong-Chan/alloydb/kv/memkv"
	mysql "github.com/Dong-Chan/alloydb/mysqldef"
	"github.com/Dong-Chan/alloydb/util/types"
)

// see https://dev.mysql.com/doc/refman/5.7/en/group-by-functions.html

type aggregateDistinct struct {
	// now we have to use memkv Temp, later may be use map directly
	Distinct memkv.Temp
}

func (c *Call) createDistinct() *aggregateDistinct {
	a := new(aggregateDistinct)

	switch strings.ToLower(c.F) {
	case "count", "sum", "avg", "group_concat":
		// only these aggregate functions support distinct
		if c.Distinct {
			a.Distinct, _ = memkv.CreateTemp(true)
		}
	}

	return a
}

// check whether v is distinct or not, return true for distinct
func (a *aggregateDistinct) isDistinct(v ...interface{}) (bool, error) {
	// no distinct flag
	if a.Distinct == nil {
		return true, nil
	}

	k := v
	r, err := a.Distinct.Get(k)
	if err != nil {
		return false, nil
	}

	if len(r) > 0 {
		// we save a same value before
		return false, nil
	}

	if err := a.Distinct.Set(k, []interface{}{true}); err != nil {
		return false, err
	}

	return true, nil
}

func (a *aggregateDistinct) clear() {
	if a.Distinct == nil {
		return
	}

	// drop does nothing, no need to check error
	a.Distinct.Drop()
	// CreateTemp returns no error, no need to check error
	// later we may use another better way instead of memkv
	a.Distinct, _ = memkv.CreateTemp(true)
}

func getDistinct(ctx map[interface{}]interface{}, fn interface{}) *aggregateDistinct {
	c, ok := fn.(*Call)
	if !ok {
		// if fn is not a Call, maybe error
		// but now we just return a dummpy aggregate distinct
		return new(aggregateDistinct)
	}

	// we may have multi aggregate function in one query
	// e.g, select sum(c1) + count(*) from t
	// so here we use a map to keep all aggregate disctinct
	m := map[interface{}]interface{}{}
	if v, ok := ctx[ExprAggDistinct]; ok {
		m = v.(map[interface{}]interface{})
	}

	if d, ok := m[c]; ok {
		return d.(*aggregateDistinct)
	}

	d := c.createDistinct()
	m[c] = d

	ctx[ExprAggDistinct] = m
	return d
}

func calculateSum(sum interface{}, v interface{}) (interface{}, error) {
	// for avg and sum calculation
	// avg and sum use decimal for integer and decimal type, use float for others
	// see https://dev.mysql.com/doc/refman/5.7/en/group-by-functions.html
	var (
		data interface{}
		err  error
	)

	switch y := v.(type) {
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
		data, err = mysql.ConvertToDecimal(v)
	case mysql.Decimal:
		data = y
	default:
		data, err = types.ToFloat64(v)
	}

	if err != nil {
		return nil, err
	}

	switch x := sum.(type) {
	case nil:
		return data, nil
	case float64:
		return x + data.(float64), nil
	case mysql.Decimal:
		return x.Add(data.(mysql.Decimal)), nil
	default:
		return nil, errors.Errorf("invalid value %v(%T) for aggregate", x, x)
	}
}

func builtinAvg(args []interface{}, ctx map[interface{}]interface{}) (v interface{}, err error) {
	// avg use decimal for integer and decimal type, use float for others
	// see https://dev.mysql.com/doc/refman/5.7/en/group-by-functions.html
	type avg struct {
		sum           interface{}
		n             uint64
		decimalResult bool
	}

	if _, ok := ctx[ExprEvalArgAggEmpty]; ok {
		return
	}

	fn := ctx[ExprEvalFn]
	distinct := getDistinct(ctx, fn)

	if _, ok := ctx[ExprAggDone]; ok {
		distinct.clear()

		data, ok := ctx[fn].(avg)
		if !ok {
			return
		}

		switch x := data.sum.(type) {
		case float64:
			return float64(x) / float64(data.n), nil
		case mysql.Decimal:
			return x.Div(mysql.NewDecimalFromUint(data.n, 0)), nil
		}

		panic("should not happend")
	}

	data, _ := ctx[fn].(avg)
	y := args[0]
	if y == nil {
		return
	}

	ok, err := distinct.isDistinct(args...)
	if err != nil || !ok {
		// if err or not distinct, return
		return nil, err
	}

	if data.sum == nil {
		data.n = 0
	}

	data.sum, err = calculateSum(data.sum, y)
	if err != nil {
		return nil, errors.Errorf("eval AVG aggregate err: %v", err)
	}

	data.n++
	ctx[fn] = data
	return
}

func builtinCount(args []interface{}, ctx map[interface{}]interface{}) (v interface{}, err error) {
	if _, ok := ctx[ExprEvalArgAggEmpty]; ok {
		return int64(0), nil
	}

	fn := ctx[ExprEvalFn]
	distinct := getDistinct(ctx, fn)

	if _, ok := ctx[ExprAggDone]; ok {
		distinct.clear()
		return ctx[fn].(int64), nil
	}

	n, _ := ctx[fn].(int64)

	if args[0] != nil {
		ok, err := distinct.isDistinct(args...)
		if err != nil || !ok {
			// if err or not distinct, return
			return nil, err
		}
		n++
	}

	ctx[fn] = n
	return
}

func builtinMax(args []interface{}, ctx map[interface{}]interface{}) (v interface{}, err error) {
	if _, ok := ctx[ExprEvalArgAggEmpty]; ok {
		return
	}

	fn := ctx[ExprEvalFn]
	if _, ok := ctx[ExprAggDone]; ok {
		if v, ok = ctx[fn]; ok {
			return
		}

		return nil, nil
	}

	max := ctx[fn]
	y := args[0]
	if y == nil {
		return
	}

	// Notice: for max, `nil < non nil`
	if max == nil {
		max = y
	} else {
		if types.Compare(max, y) < 0 {
			max = y
		}
	}

	ctx[fn] = max
	return
}

func builtinMin(args []interface{}, ctx map[interface{}]interface{}) (v interface{}, err error) {
	if _, ok := ctx[ExprEvalArgAggEmpty]; ok {
		return
	}

	fn := ctx[ExprEvalFn]
	if _, ok := ctx[ExprAggDone]; ok {
		if v, ok = ctx[fn]; ok {
			return
		}

		return nil, nil
	}

	min := ctx[fn]
	y := args[0]
	if y == nil {
		return
	}

	// Notice: for min, `nil > non nil`
	if min == nil {
		min = y
	} else {
		if types.Compare(min, y) > 0 {
			min = y
		}
	}

	ctx[fn] = min
	return
}

func builtinSum(args []interface{}, ctx map[interface{}]interface{}) (v interface{}, err error) {
	if _, ok := ctx[ExprEvalArgAggEmpty]; ok {
		return
	}

	fn := ctx[ExprEvalFn]
	distinct := getDistinct(ctx, fn)

	if _, ok := ctx[ExprAggDone]; ok {
		distinct.clear()
		if v, ok = ctx[fn]; ok {
			return
		}

		return nil, nil
	}

	sum := ctx[fn]
	y := args[0]
	if y == nil {
		return
	}

	ok, err := distinct.isDistinct(args...)
	if err != nil || !ok {
		// if err or not distinct, return
		return nil, err
	}

	sum, err = calculateSum(sum, y)
	if err != nil {
		return nil, errors.Errorf("eval SUM aggregate err: %v", err)
	}

	ctx[fn] = sum
	return
}

func builtinGroupConcat(args []interface{}, ctx map[interface{}]interface{}) (v interface{}, err error) {
	// TODO: the real group_concat is very complex, here we just support the simplest one.
	if _, ok := ctx[ExprEvalArgAggEmpty]; ok {
		return nil, nil
	}

	fn := ctx[ExprEvalFn]
	distinct := getDistinct(ctx, fn)
	if _, ok := ctx[ExprAggDone]; ok {
		distinct.clear()
		if v, _ := ctx[fn]; v != nil {
			return v.(string), nil
		}
		return nil, nil
	}

	var buf bytes.Buffer
	if v := ctx[fn]; v != nil {
		s := v.(string)
		// now use comma separator
		buf.WriteString(s)
		buf.WriteString(",")
	}

	ok, err := distinct.isDistinct(args...)
	if err != nil || !ok {
		// if err or not distinct, return
		return nil, err
	}

	for i := 0; i < len(args); i++ {
		if args[i] == nil {
			// if any is nil, we will not concat
			return
		}

		buf.WriteString(fmt.Sprintf("%v", args[i]))
	}

	// TODO: if total length is greater than global var group_concat_max_len, truncate it.
	ctx[fn] = buf.String()
	return
}
