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
package utils

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWritingTable(t *testing.T) {
	buf := &bytes.Buffer{}
	tbl := NewTableWriter(buf, 5, 10)
	assert.NotNil(t, tbl)
	tbl.WriteHeader("testing", "something")
	tbl.Align = AlignLeft
	tbl.WriteRow("foo", "bar", "boo")
	tbl.WriteRow("ok", "corral")
	tbl.Footer()

	expected := `+-------+------------+
| te... |  something |
+-------+------------+
| foo   | bar        |
| ok    | corral     |
+-------+------------+
`
	fmt.Printf(buf.String())
	assert.Equal(t, expected, buf.String())
}

func TestWritingTableWithRune(t *testing.T) {
	buf := &bytes.Buffer{}
	tbl := NewTableWriter(buf, 5, 10)
	assert.NotNil(t, tbl)
	tbl.WriteHeader("世界日本語世界", "something")
	tbl.Align = AlignLeft
	tbl.WriteRow("foo", "bar", "boo")
	tbl.WriteRow("ok", "corral")
	tbl.Footer()

	expected := `+-------+------------+
| 世界... |  something |
+-------+------------+
| foo   | bar        |
| ok    | corral     |
+-------+------------+
`
	fmt.Printf(buf.String())
	assert.Equal(t, expected, buf.String())
}
