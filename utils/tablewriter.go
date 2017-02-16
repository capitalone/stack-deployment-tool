//
// Copyright 2016 Capital One Services, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and limitations under the License.
//
// SPDX-Copyright: Copyright (c) Capital One Services, LLC
// SPDX-License-Identifier: Apache-2.0
//
package utils

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/width"
)

const (
	trimSize   = 3
	AlignRight = 1
	AlignLeft  = 2
)

type TableWriter struct {
	colLens []int
	writer  io.Writer
	corner  string
	line    string
	trimSep string
	colSep  string
	Align   int
}

func NewTableWriter(w io.Writer, colLens ...int) *TableWriter {
	return &TableWriter{colLens: colLens, writer: w, corner: "+", line: "-", trimSep: "...",
		colSep: "|", Align: AlignRight}
}

func (t *TableWriter) WriteHeader(headers ...string) {
	t.writeLine()
	t.WriteRow(headers...)
	t.writeLine()
}

func (t *TableWriter) WriteRow(columns ...string) {
	for i, col := range columns {
		t.writeCol(i, col)
	}
	fmt.Fprintf(t.writer, t.colSep+"\n")
}

func (t *TableWriter) Footer() {
	t.writeLine()
}

func (t *TableWriter) writeCol(colNum int, val string) {
	if colNum > len(t.colLens)-1 {
		return
	}
	col := val
	if utf8.RuneCountInString(val) > t.colLens[colNum] {
		lastIndex := 0
		for runes, w := 0, 0; lastIndex < len(val) && runes < t.colLens[colNum]-trimSize; lastIndex += w {
			_, width := utf8.DecodeRuneInString(val[lastIndex:])
			w = width
			runes++
		}
		col = val[0:lastIndex] + t.trimSep
	}

	cnt := utf8.RuneCountInString(val)
	if cnt < t.colLens[colNum] {
		if t.Align == AlignRight {
			col = strings.Repeat(" ", t.colLens[colNum]-cnt) + col
		} else {
			col += strings.Repeat(" ", t.colLens[colNum]-cnt)
		}
	}

	fmt.Fprintf(t.writer, t.colSep+" %s ", width.Narrow.String(col))
}

func (t *TableWriter) writeLine() {
	for _, l := range t.colLens {
		fmt.Fprintf(t.writer, t.corner+"%s", strings.Repeat(t.line, l+2))
	}
	if len(t.colLens) > 0 {
		fmt.Fprintf(t.writer, t.corner+"\n")
	}
}
