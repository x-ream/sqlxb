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

import "testing"

// Test Po interface
type TestPo struct {
	ID   uint64 `db:"id"`
	Name string `db:"name"`
}

func (*TestPo) TableName() string {
	return "test_table"
}

func TestPoInterface(t *testing.T) {
	po := &TestPo{}
	tableName := po.TableName()

	if tableName != "test_table" {
		t.Errorf("TableName() = %v, want test_table", tableName)
	}
}

// Test LongId interface
type TestLongId struct {
	ID uint64 `db:"id"`
}

func (t *TestLongId) GetId() uint64 {
	return t.ID
}

func TestLongIdInterface(t *testing.T) {
	entity := &TestLongId{ID: 12345}
	id := entity.GetId()

	if id != 12345 {
		t.Errorf("GetId() = %v, want 12345", id)
	}
}

// Test StringId interface
type TestStringId struct {
	ID string `db:"id"`
}

func (t *TestStringId) GetId() string {
	return t.ID
}

func TestStringIdInterface(t *testing.T) {
	entity := &TestStringId{ID: "abc123"}
	id := entity.GetId()

	if id != "abc123" {
		t.Errorf("GetId() = %v, want abc123", id)
	}
}
