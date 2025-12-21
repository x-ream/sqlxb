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

// Test all Pointer helpers
func TestPointerHelpers(t *testing.T) {
	// Bool
	b := Bool(true)
	if b == nil || *b != true {
		t.Error("Bool() failed")
	}

	// Int
	i := Int(42)
	if i == nil || *i != 42 {
		t.Error("Int() failed")
	}

	// Int64
	i64 := Int64(int64(123))
	if i64 == nil || *i64 != 123 {
		t.Error("Int64() failed")
	}

	// Int32
	i32 := Int32(int32(456))
	if i32 == nil || *i32 != 456 {
		t.Error("Int32() failed")
	}

	// Int16
	i16 := Int16(int16(789))
	if i16 == nil || *i16 != 789 {
		t.Error("Int16() failed")
	}

	// Int8
	i8 := Int8(int8(12))
	if i8 == nil || *i8 != 12 {
		t.Error("Int8() failed")
	}

	// Byte
	by := Byte(byte(255))
	if by == nil || *by != 255 {
		t.Error("Byte() failed")
	}

	// Float64
	f64 := Float64(3.14)
	if f64 == nil || *f64 != 3.14 {
		t.Error("Float64() failed")
	}

	// Float32
	f32 := Float32(float32(2.71))
	if f32 == nil || *f32 != float32(2.71) {
		t.Error("Float32() failed")
	}

	// Uint64
	u64 := Uint64(uint64(999))
	if u64 == nil || *u64 != 999 {
		t.Error("Uint64() failed")
	}

	// Uint
	u := Uint(uint(888))
	if u == nil || *u != 888 {
		t.Error("Uint() failed")
	}
}

// Test Np2s function (Nullable Pointer to String)
func TestNp2s(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
		valid    bool
	}{
		// Non nil pointer
		{"uint64", Uint64(123), "123", true},
		{"uint", Uint(456), "456", true},
		{"int64", Int64(-789), "-789", true},
		{"int", Int(-999), "-999", true},
		{"int32", Int32(111), "111", true},
		{"int16", Int16(222), "222", true},
		{"int8", Int8(-33), "-33", true},
		{"byte", Byte(255), "255", true},
		{"float64", Float64(3.14), "3.14", true},
		{"float32", Float32(2.71), "2.71", true},

		// Nil pointer
		{"nil uint64", (*uint64)(nil), "", false},
		{"nil uint", (*uint)(nil), "", false},
		{"nil int64", (*int64)(nil), "", false},
		{"nil int", (*int)(nil), "", false},
		{"nil int32", (*int32)(nil), "", false},
		{"nil int16", (*int16)(nil), "", false},
		{"nil int8", (*int8)(nil), "", false},
		{"nil byte", (*byte)(nil), "", false},
		{"nil float64", (*float64)(nil), "", false},
		{"nil float32", (*float32)(nil), "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, valid := Np2s(tt.input)
			if valid != tt.valid {
				t.Errorf("Np2s(%v) valid = %v, want %v", tt.name, valid, tt.valid)
			}
			if result != tt.expected {
				t.Errorf("Np2s(%v) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

// Test N2s function (Number to String)
func TestN2s(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"uint64", uint64(123), "123"},
		{"uint", uint(456), "456"},
		{"int64", int64(-789), "-789"},
		{"int", int(-999), "-999"},
		{"int32", int32(111), "111"},
		{"int16", int16(222), "222"},
		{"int8", int8(-33), "-33"},
		{"byte", byte(255), "255"},
		{"float64", float64(3.14), "3.14"},
		{"float32", float32(2.71), "2.71"},
		{"unknown type", "string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := N2s(tt.input)
			if result != tt.expected {
				t.Errorf("N2s(%v) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

// Test NilOrNumber function
func TestNilOrNumber(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		expectNil bool
		expected  interface{}
	}{
		// Non nil pointer
		{"uint64", Uint64(123), false, uint64(123)},
		{"uint", Uint(456), false, uint(456)},
		{"int64", Int64(-789), false, int64(-789)},
		{"int", Int(-999), false, int(-999)},
		{"int32", Int32(111), false, int32(111)},
		{"int16", Int16(222), false, int16(222)},
		{"int8", Int8(-33), false, int8(-33)},
		{"byte", Byte(255), false, byte(255)},
		{"float64", Float64(3.14), false, float64(3.14)},
		{"float32", Float32(2.71), false, float32(2.71)},
		{"bool", Bool(true), false, true},

		// Nil pointer
		{"nil uint64", (*uint64)(nil), true, nil},
		{"nil uint", (*uint)(nil), true, nil},
		{"nil int64", (*int64)(nil), true, nil},
		{"nil int", (*int)(nil), true, nil},
		{"nil int32", (*int32)(nil), true, nil},
		{"nil int16", (*int16)(nil), true, nil},
		{"nil int8", (*int8)(nil), true, nil},
		{"nil byte", (*byte)(nil), true, nil},
		{"nil float64", (*float64)(nil), true, nil},
		{"nil float32", (*float32)(nil), true, nil},
		{"nil bool", (*bool)(nil), true, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isNil, value := NilOrNumber(tt.input)
			if isNil != tt.expectNil {
				t.Errorf("NilOrNumber(%v) isNil = %v, want %v", tt.name, isNil, tt.expectNil)
			}
			if !tt.expectNil && value != tt.expected {
				t.Errorf("NilOrNumber(%v) value = %v, want %v", tt.name, value, tt.expected)
			}
		})
	}
}

// Test NilOrNumber panic case
func TestNilOrNumber_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for unsupported type")
		}
	}()

	// Unsupported type should panic
	NilOrNumber("unsupported")
}
