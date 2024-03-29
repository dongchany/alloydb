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

// Package mock is just for test only.
package mock

import (
	"fmt"

	"github.com/Dong-Chan/alloydb/context"
	"github.com/Dong-Chan/alloydb/kv"
)

var _ context.Context = (*Context)(nil)

// Context is a mock context.Context for test.
type Context struct {
	values map[fmt.Stringer]interface{}
}

// SetValue implements context.Context SetValue interface.
func (c *Context) SetValue(key fmt.Stringer, value interface{}) {
	c.values[key] = value
}

// Value implements context.Context Value interface.
func (c *Context) Value(key fmt.Stringer) interface{} {
	value := c.values[key]
	return value
}

// ClearValue implements context.Context ClearValue interface.
func (c *Context) ClearValue(key fmt.Stringer) {
	delete(c.values, key)
}

// GetTxn implements context.Context GetTxn interface.
func (c *Context) GetTxn(forceNew bool) (kv.Transaction, error) {
	return nil, nil
}

// FinishTxn implements context.Context FinishTxn interface.
func (c *Context) FinishTxn(rollback bool) error {
	return nil
}

// NewContext creates a mock context.Context.
func NewContext() context.Context {
	return &Context{
		values: make(map[fmt.Stringer]interface{}),
	}
}
