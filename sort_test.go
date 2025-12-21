// Copyright 2025 me.fndo.xb
//
// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package xb

import (
	"testing"
)

// Test ASC function
func TestASC(t *testing.T) {
	result := ASC()
	if result != "ASC" {
		t.Errorf("ASC() = %v, want ASC", result)
	}
}

// Test DESC function
func TestDESC(t *testing.T) {
	result := DESC()
	if result != "DESC" {
		t.Errorf("DESC() = %v, want DESC", result)
	}
}

// Test Sort structure
type testStruct struct {
	Id        int64  `db:"id"`
	CreatedAt int64  `db:"created_at"`
	Status    string `db:"status"`
}

func (*testStruct) TableName() string {
	return "test_table"
}

func TestSort(t *testing.T) {
	// Test ASC sorting
	builder := Of(&testStruct{}).
		Sort("id", ASC).
		Build()

	sql, _, _ := builder.SqlOfSelect()
	if !containsString(sql, "ORDER BY id ASC") {
		t.Errorf("Sort ASC failed, got: %s", sql)
	}

	// Test DESC sorting
	builder = Of(&testStruct{}).
		Sort("created_at", DESC).
		Build()

	sql, _, _ = builder.SqlOfSelect()
	if !containsString(sql, "ORDER BY created_at DESC") {
		t.Errorf("Sort DESC failed, got: %s", sql)
	}

	// Test multiple sorting
	builder = Of(&testStruct{}).
		Sort("status", ASC).
		Sort("id", DESC).
		Build()

	sql, _, _ = builder.SqlOfSelect()
	if !containsString(sql, "ORDER BY status ASC, id DESC") {
		t.Errorf("Multiple Sort failed, got: %s", sql)
	}
}
