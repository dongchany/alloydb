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
	"time"

	"github.com/juju/errors"
	mysql "github.com/Dong-Chan/alloydb/mysqldef"
	"github.com/Dong-Chan/alloydb/util/types"
)

func convertToTime(arg interface{}, tp byte) (interface{}, error) {
	f := types.NewFieldType(tp)
	f.Decimal = mysql.MaxFsp

	v, err := types.Convert(arg, f)
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, nil
	}

	t, ok := v.(mysql.Time)
	if !ok {
		return nil, errors.Errorf("need time type, but got %T", v)
	}

	return t, nil
}

func convertToDuration(arg interface{}) (interface{}, error) {
	f := types.NewFieldType(mysql.TypeDuration)
	f.Decimal = mysql.MaxFsp

	v, err := types.Convert(arg, f)
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, nil
	}

	t, ok := v.(mysql.Duration)
	if !ok {
		return nil, errors.Errorf("need duration type, but got %T", v)
	}

	return t, nil
}

func builtinDate(args []interface{}, _ map[interface{}]interface{}) (interface{}, error) {
	return convertToTime(args[0], mysql.TypeDate)
}

// See http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_day
// day is a synonym for DayOfMonth
func builtinDay(args []interface{}, ctx map[interface{}]interface{}) (interface{}, error) {
	return builtinDayOfMonth(args, ctx)
}

// See http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_hour
func builtinHour(args []interface{}, ctx map[interface{}]interface{}) (interface{}, error) {
	v, err := convertToDuration(args[0])
	if err != nil || v == nil {
		return v, err
	}

	// No need to check type here.
	d := v.(mysql.Duration)
	return int64(d.Hour()), nil
}

// See http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_minute
func builtinMinute(args []interface{}, ctx map[interface{}]interface{}) (interface{}, error) {
	v, err := convertToDuration(args[0])
	if err != nil || v == nil {
		return v, err
	}

	// No need to check type here.
	d := v.(mysql.Duration)
	return int64(d.Minute()), nil
}

// See http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_second
func builtinSecond(args []interface{}, ctx map[interface{}]interface{}) (interface{}, error) {
	v, err := convertToDuration(args[0])
	if err != nil || v == nil {
		return v, err
	}

	// No need to check type here.
	d := v.(mysql.Duration)
	return int64(d.Second()), nil
}

// See http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_microsecond
func builtinMicroSecond(args []interface{}, ctx map[interface{}]interface{}) (interface{}, error) {
	v, err := convertToDuration(args[0])
	if err != nil || v == nil {
		return v, err
	}

	// No need to check type here.
	d := v.(mysql.Duration)
	return int64(d.MicroSecond()), nil
}

// See http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_month
func builtinMonth(args []interface{}, ctx map[interface{}]interface{}) (interface{}, error) {
	v, err := convertToTime(args[0], mysql.TypeDate)
	if err != nil || v == nil {
		return v, err
	}

	// No need to check type here.
	t := v.(mysql.Time)
	if t.IsZero() {
		return int64(0), nil
	}

	return int64(t.Month()), nil
}

func builtinNow(args []interface{}, ctx map[interface{}]interface{}) (interface{}, error) {
	fsp := int64(0)
	if len(args) == 1 {
		var err error
		fsp, err = types.ToInt64(args[0])
		if err != nil {
			return nil, errors.Trace(err)
		}
		if int(fsp) > mysql.MaxFsp {
			return nil, errors.Errorf("Too big precision %d specified. Maximum is 6.", fsp)
		} else if fsp < 0 {
			return nil, errors.Errorf("Invalid negative %d specified, must in [0, 6].", fsp)
		}
	}

	t := mysql.Time{
		Time: time.Now(),
		Type: mysql.TypeDatetime,
		// set unspecified for later round
		Fsp: mysql.UnspecifiedFsp,
	}

	return t.RoundFrac(int(fsp))
}

// See http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayofmonth
func builtinDayOfMonth(args []interface{}, ctx map[interface{}]interface{}) (interface{}, error) {
	// TODO: some invalid format like 2000-00-00 will return 0 too.
	v, err := convertToTime(args[0], mysql.TypeDate)
	if err != nil || v == nil {
		return v, err
	}

	// No need to check type here.
	t := v.(mysql.Time)
	if t.IsZero() {
		return int64(0), nil
	}

	return int64(t.Day()), nil
}

// See http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayofweek
func builtinDayOfWeek(args []interface{}, ctx map[interface{}]interface{}) (interface{}, error) {
	v, err := convertToTime(args[0], mysql.TypeDate)
	if err != nil || v == nil {
		return v, err
	}

	// No need to check type here.
	t := v.(mysql.Time)
	if t.IsZero() {
		// TODO: log warning or return error?
		return nil, nil
	}

	// 1 is Sunday, 2 is Monday, .... 7 is Saturday
	return int64(t.Weekday()) + 1, nil
}

// See http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_dayofyear
func builtinDayOfYear(args []interface{}, ctx map[interface{}]interface{}) (interface{}, error) {
	v, err := convertToTime(args[0], mysql.TypeDate)
	if err != nil || v == nil {
		return v, err
	}

	t := v.(mysql.Time)
	if t.IsZero() {
		// TODO: log warning or return error?
		return nil, nil
	}

	return int64(t.YearDay()), nil
}

// See http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_week
func builtinWeek(args []interface{}, ctx map[interface{}]interface{}) (interface{}, error) {
	v, err := convertToTime(args[0], mysql.TypeDate)
	if err != nil || v == nil {
		return v, err
	}

	// No need to check type here.
	t := v.(mysql.Time)
	if t.IsZero() {
		// TODO: log warning or return error?
		return nil, nil
	}

	// TODO: support multi mode for week
	_, week := t.ISOWeek()
	return int64(week), nil
}

// See http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_weekday
func builtinWeekDay(args []interface{}, ctx map[interface{}]interface{}) (interface{}, error) {
	v, err := convertToTime(args[0], mysql.TypeDate)
	if err != nil || v == nil {
		return v, err
	}

	// No need to check type here.
	t := v.(mysql.Time)
	if t.IsZero() {
		// TODO: log warning or return error?
		return nil, nil
	}

	// Monday is 0, ... Sunday = 6 in MySQL
	// but in go, Sunday is 0, ... Saturday is 6
	// w will do a conversion.
	w := (int64(t.Weekday()) + 6) % 7
	return w, nil
}

// See http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_weekofyear
func builtinWeekOfYear(args []interface{}, ctx map[interface{}]interface{}) (interface{}, error) {
	// WeekOfYear is equivalent to to Week(date, 3)
	return builtinWeek([]interface{}{args[0], 3}, ctx)
}

// See http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_year
func builtinYear(args []interface{}, ctx map[interface{}]interface{}) (interface{}, error) {
	v, err := convertToTime(args[0], mysql.TypeDate)
	if err != nil || v == nil {
		return v, err
	}

	// No need to check type here.
	t := v.(mysql.Time)
	if t.IsZero() {
		return int64(0), nil
	}

	return int64(t.Year()), nil
}

// See http://dev.mysql.com/doc/refman/5.7/en/date-and-time-functions.html#function_yearweek
func builtinYearWeek(args []interface{}, ctx map[interface{}]interface{}) (interface{}, error) {
	v, err := convertToTime(args[0], mysql.TypeDate)
	if err != nil || v == nil {
		return v, err
	}

	// No need to check type here.
	t := v.(mysql.Time)
	if t.IsZero() {
		// TODO: log warning or return error?
		return nil, nil
	}

	// TODO: support multi mode for week
	year, week := t.ISOWeek()
	return int64(year*100 + week), nil
}
