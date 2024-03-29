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

package printer

import (
	"bytes"
	"fmt"
)

// Version information.
var (
	TiDBBuildTS = "None"
	TiDBGitHash = "None"
)

// PrintTiDBInfo prints the TiDB version information.
func PrintTiDBInfo() {
	fmt.Printf("Welcome to the TiDB.\n")
	fmt.Printf("Version:\n")
	fmt.Printf("Git Commit Hash: %s\n", TiDBGitHash)
	fmt.Printf("UTC Build Time:  %s\n", TiDBBuildTS)
	fmt.Printf("\n")
}

// checkValidity checks whether cols and every data have the same length.
func checkValidity(cols []string, datas [][]string) bool {
	colLen := len(cols)
	if len(datas) == 0 || colLen == 0 {
		return false
	}

	for _, data := range datas {
		if colLen != len(data) {
			return false
		}
	}

	return true
}

func getMaxColLen(cols []string, datas [][]string) []int {
	maxColLen := make([]int, len(cols))
	for i, col := range cols {
		maxColLen[i] = len(col)
	}

	for _, data := range datas {
		for i, v := range data {
			if len(v) > maxColLen[i] {
				maxColLen[i] = len(v)
			}
		}
	}

	return maxColLen
}

func getPrintDivLine(maxColLen []int) []byte {
	var value []byte
	for _, v := range maxColLen {
		value = append(value, '+')
		value = append(value, bytes.Repeat([]byte{'-'}, v+2)...)
	}
	value = append(value, '+')
	value = append(value, '\n')
	return value
}

func getPrintCol(cols []string, maxColLen []int) []byte {
	var value []byte
	for i, v := range cols {
		value = append(value, '|')
		value = append(value, ' ')
		value = append(value, []byte(v)...)
		value = append(value, bytes.Repeat([]byte{' '}, maxColLen[i]+1-len(v))...)
	}
	value = append(value, '|')
	value = append(value, '\n')
	return value
}

func getPrintRow(data []string, maxColLen []int) []byte {
	var value []byte
	for i, v := range data {
		value = append(value, '|')
		value = append(value, ' ')
		value = append(value, []byte(v)...)
		value = append(value, bytes.Repeat([]byte{' '}, maxColLen[i]+1-len(v))...)
	}
	value = append(value, '|')
	value = append(value, '\n')
	return value
}

func getPrintRows(datas [][]string, maxColLen []int) []byte {
	var value []byte
	for _, data := range datas {
		value = append(value, getPrintRow(data, maxColLen)...)
	}
	return value
}

// GetPrintResult gets a result with a formatted string.
func GetPrintResult(cols []string, datas [][]string) (string, bool) {
	if !checkValidity(cols, datas) {
		return "", false
	}

	var value []byte
	maxColLen := getMaxColLen(cols, datas)

	value = append(value, getPrintDivLine(maxColLen)...)
	value = append(value, getPrintCol(cols, maxColLen)...)
	value = append(value, getPrintDivLine(maxColLen)...)
	value = append(value, getPrintRows(datas, maxColLen)...)
	value = append(value, getPrintDivLine(maxColLen)...)
	return string(value), true
}
