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

package codec

import (
	"bytes"
	"math"
	"testing"

	. "github.com/pingcap/check"
)

func TestT(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&testCodecSuite{})

type testCodecSuite struct {
}

func (s *testCodecSuite) TestCodecKey(c *C) {
	table := []struct {
		Input  []interface{}
		Expect []interface{}
	}{
		{
			[]interface{}{int64(1)},
			[]interface{}{int64(1)},
		},

		{
			[]interface{}{float32(1), float64(3.15), []byte("123"), "123"},
			[]interface{}{float64(1), float64(3.15), []byte("123"), "123"},
		},

		{
			[]interface{}{uint64(1), float64(3.15), []byte("123"), int64(-1)},
			[]interface{}{uint64(1), float64(3.15), []byte("123"), int64(-1)},
		},

		{
			[]interface{}{true, false},
			[]interface{}{int64(1), int64(0)},
		},

		{
			[]interface{}{int(1), int8(1), int16(1), int32(1), int64(1)},
			[]interface{}{int64(1), int64(1), int64(1), int64(1), int64(1)},
		},

		{
			[]interface{}{uint(1), uint8(1), uint16(1), uint32(1), uint64(1)},
			[]interface{}{uint64(1), uint64(1), uint64(1), uint64(1), uint64(1)},
		},

		{
			[]interface{}{nil},
			[]interface{}{nil},
		},
	}

	for _, t := range table {
		b, err := EncodeKey(t.Input...)
		c.Assert(err, IsNil)
		args, err := DecodeKey(b)
		c.Assert(err, IsNil)

		c.Assert(args, DeepEquals, t.Expect)
	}
}

func (s *testCodecSuite) TestCodecKeyCompare(c *C) {
	table := []struct {
		Left   []interface{}
		Right  []interface{}
		Expect int
	}{
		{
			[]interface{}{1},
			[]interface{}{1},
			0,
		},
		{
			[]interface{}{-1},
			[]interface{}{1},
			-1,
		},
		{
			[]interface{}{3.15},
			[]interface{}{3.12},
			1,
		},
		{
			[]interface{}{"abc"},
			[]interface{}{"abcd"},
			-1,
		},
		{
			[]interface{}{1, "abc"},
			[]interface{}{1, "abcd"},
			-1,
		},
		{
			[]interface{}{1, "abc", "def"},
			[]interface{}{1, "abcd", "af"},
			-1,
		},
		{
			[]interface{}{3.12, "ebc", "def"},
			[]interface{}{2.12, "abcd", "af"},
			1,
		},
		{
			[]interface{}{[]byte{0x01, 0x00}, []byte{0xFF}},
			[]interface{}{[]byte{0x01, 0x00, 0xFF}},
			-1,
		},

		{
			[]interface{}{[]byte{0x01}, 0xFFFFFFFFFFFFFFF},
			[]interface{}{[]byte{0x01, 0x10}, 0},
			-1,
		},
		{
			[]interface{}{0},
			[]interface{}{nil},
			1,
		},
		{
			[]interface{}{[]byte{0x00}},
			[]interface{}{nil},
			1,
		},
		{
			[]interface{}{math.SmallestNonzeroFloat64},
			[]interface{}{nil},
			1,
		},
		{
			[]interface{}{math.MinInt64},
			[]interface{}{nil},
			1,
		},
		{
			[]interface{}{1, math.MinInt64, nil},
			[]interface{}{1, nil, uint64(math.MaxUint64)},
			1,
		},
		{
			[]interface{}{1, []byte{}, nil},
			[]interface{}{1, nil, 123},
			1,
		},
	}

	for _, t := range table {
		b1, err := EncodeKey(t.Left...)
		c.Assert(err, IsNil)

		b2, err := EncodeKey(t.Right...)
		c.Assert(err, IsNil)

		c.Assert(bytes.Compare(b1, b2), Equals, t.Expect)
	}
}

func (s *testCodecSuite) TestNumberCodec(c *C) {
	tblInt64 := []int64{
		math.MinInt64,
		math.MinInt32,
		math.MinInt16,
		math.MinInt8,
		0,
		math.MaxInt8,
		math.MaxInt16,
		math.MaxInt32,
		math.MaxInt64,
		1<<47 - 1,
		-1 << 47,
		1<<23 - 1,
		-1 << 23,
		1<<55 - 1,
		-1 << 55,
		1,
		-1,
	}

	for _, t := range tblInt64 {
		b := EncodeInt(nil, t)
		_, v, err := DecodeInt(b)
		c.Assert(err, IsNil)
		c.Assert(v, Equals, t)

		b = EncodeIntDesc(nil, t)
		_, v, err = DecodeIntDesc(b)
		c.Assert(err, IsNil)
		c.Assert(v, Equals, t)
	}

	tblUint64 := []uint64{
		0,
		math.MaxUint8,
		math.MaxUint16,
		math.MaxUint32,
		math.MaxUint64,
		1<<24 - 1,
		1<<48 - 1,
		1<<56 - 1,
		1,
		math.MaxInt16,
		math.MaxInt8,
		math.MaxInt32,
		math.MaxInt64,
	}

	for _, t := range tblUint64 {
		b := EncodeUint(nil, t)
		_, v, err := DecodeUint(b)
		c.Assert(err, IsNil)
		c.Assert(v, Equals, t)

		b = EncodeUintDesc(nil, t)
		_, v, err = DecodeUintDesc(b)
		c.Assert(err, IsNil)
		c.Assert(v, Equals, t)
	}
}

func (s *testCodecSuite) TestNumberOrder(c *C) {
	tblInt64 := []struct {
		Arg1 int64
		Arg2 int64
		Ret  int
	}{
		{-1, 1, -1},
		{math.MaxInt64, math.MinInt64, 1},
		{math.MaxInt64, math.MaxInt32, 1},
		{math.MinInt32, math.MaxInt16, -1},
		{math.MinInt64, math.MaxInt8, -1},
		{0, math.MaxInt8, -1},
		{math.MinInt8, 0, -1},
		{math.MinInt16, math.MaxInt16, -1},
		{1, -1, 1},
		{1, 0, 1},
		{-1, 0, -1},
		{0, 0, 0},
		{math.MaxInt16, math.MaxInt16, 0},
	}

	for _, t := range tblInt64 {
		b1 := EncodeInt(nil, t.Arg1)
		b2 := EncodeInt(nil, t.Arg2)

		ret := bytes.Compare(b1, b2)
		c.Assert(ret, Equals, t.Ret)

		b1 = EncodeIntDesc(nil, t.Arg1)
		b2 = EncodeIntDesc(nil, t.Arg2)

		ret = bytes.Compare(b1, b2)
		c.Assert(ret, Equals, -t.Ret)
	}

	tblUint64 := []struct {
		Arg1 uint64
		Arg2 uint64
		Ret  int
	}{
		{0, 0, 0},
		{1, 0, 1},
		{0, 1, -1},
		{math.MaxInt8, math.MaxInt16, -1},
		{math.MaxUint32, math.MaxInt32, 1},
		{math.MaxUint8, math.MaxInt8, 1},
		{math.MaxUint16, math.MaxInt32, -1},
		{math.MaxUint64, math.MaxInt64, 1},
		{math.MaxInt64, math.MaxUint32, 1},
		{math.MaxUint64, 0, 1},
		{0, math.MaxUint64, -1},
	}

	for _, t := range tblUint64 {
		b1 := EncodeUint(nil, t.Arg1)
		b2 := EncodeUint(nil, t.Arg2)

		ret := bytes.Compare(b1, b2)
		c.Assert(ret, Equals, t.Ret)

		b1 = EncodeUintDesc(nil, t.Arg1)
		b2 = EncodeUintDesc(nil, t.Arg2)

		ret = bytes.Compare(b1, b2)
		c.Assert(ret, Equals, -t.Ret)
	}
}

func (s *testCodecSuite) TestFloatCodec(c *C) {
	tblFloat := []float64{
		-1,
		0,
		1,
		math.MaxFloat64,
		math.MaxFloat32,
		math.SmallestNonzeroFloat32,
		math.SmallestNonzeroFloat64}

	for _, t := range tblFloat {
		b := EncodeFloat(nil, t)
		_, v, err := DecodeFloat(b)
		c.Assert(err, IsNil)
		c.Assert(v, Equals, t)

		b = EncodeFloatDesc(nil, t)
		_, v, err = DecodeFloatDesc(b)
		c.Assert(err, IsNil)
		c.Assert(v, Equals, t)
	}

	tblCmp := []struct {
		Arg1 float64
		Arg2 float64
		Ret  int
	}{
		{1, -1, 1},
		{1, 0, 1},
		{0, -1, 1},
		{0, 0, 0},
		{math.MaxFloat64, 1, 1},
		{math.MaxFloat32, math.MaxFloat64, -1},
		{math.MaxFloat64, 0, 1},
		{math.MaxFloat64, math.SmallestNonzeroFloat64, 1},
	}

	for _, t := range tblCmp {
		b1 := EncodeFloat(nil, t.Arg1)
		b2 := EncodeFloat(nil, t.Arg2)

		ret := bytes.Compare(b1, b2)
		c.Assert(ret, Equals, t.Ret)

		b1 = EncodeFloatDesc(nil, t.Arg1)
		b2 = EncodeFloatDesc(nil, t.Arg2)

		ret = bytes.Compare(b1, b2)
		c.Assert(ret, Equals, -t.Ret)
	}
}

func (s *testCodecSuite) TestBytes(c *C) {
	tblBytes := [][]byte{
		[]byte{},
		[]byte{0x00, 0x01},
		[]byte{0xff, 0xff},
		[]byte{0x01, 0x00},
		[]byte("abc"),
		[]byte("hello world"),
	}

	for _, t := range tblBytes {
		b := EncodeBytes(nil, t)
		_, v, err := DecodeBytes(b)
		c.Assert(err, IsNil)
		c.Assert(t, DeepEquals, v)

		b = EncodeBytesDesc(nil, t)
		_, v, err = DecodeBytesDesc(b)
		c.Assert(err, IsNil)
		c.Assert(t, DeepEquals, v)
	}

	tblCmp := []struct {
		Arg1 []byte
		Arg2 []byte
		Ret  int
	}{
		{[]byte{}, []byte{0x00}, -1},
		{[]byte{0x00}, []byte{0x00}, 0},
		{[]byte{0xFF}, []byte{0x00}, 1},
		{[]byte{0xFF}, []byte{0xFF, 0x00}, -1},
		{[]byte("a"), []byte("b"), -1},
		{[]byte("a"), []byte{0x00}, 1},
	}

	for _, t := range tblCmp {
		b1 := EncodeBytes(nil, t.Arg1)
		b2 := EncodeBytes(nil, t.Arg2)

		ret := bytes.Compare(b1, b2)
		c.Assert(ret, Equals, t.Ret)

		b1 = EncodeBytesDesc(nil, t.Arg1)
		b2 = EncodeBytesDesc(nil, t.Arg2)

		ret = bytes.Compare(b1, b2)
		c.Assert(ret, Equals, -t.Ret)
	}
}
