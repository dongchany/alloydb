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

package mysqldef

// GetDefaultFieldLength is used for Interger Types, Flen is the display length.
// Call this when no Flen assigned in ddl.
// or column value is calculated from an expression.
// For example: "select count(*) from t;", the column type is int64 and Flen in ResultField will be 21.
func GetDefaultFieldLength(tp byte) int {
	switch tp {
	case TypeTiny:
		return 4
	case TypeShort:
		return 6
	case TypeInt24:
		return 9
	case TypeLong:
		return 11
	case TypeLonglong:
		return 21
	default:
		//TODO: add more types
		return 0
	}
}
