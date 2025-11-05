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
	"strings"
	"testing"
)

// TestQdrantBuilder_HnswEf_Validation 测试 HnswEf 参数校验
func TestQdrantBuilder_HnswEf_Validation(t *testing.T) {
	tests := []struct {
		name      string
		ef        int
		wantPanic bool
	}{
		{"Valid: ef=1", 1, false},
		{"Valid: ef=64", 64, false},
		{"Valid: ef=128", 128, false},
		{"Valid: ef=512", 512, false},
		{"Invalid: ef=0", 0, true},
		{"Invalid: ef=-1", -1, true},
		{"Invalid: ef=-100", -100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.wantPanic && r == nil {
					t.Errorf("Expected panic for ef=%d, but didn't panic", tt.ef)
				}
				if !tt.wantPanic && r != nil {
					t.Errorf("Unexpected panic for ef=%d: %v", tt.ef, r)
				}
				if r != nil {
					msg := r.(string)
					if !strings.Contains(msg, "HnswEf must be >= 1") {
						t.Errorf("Panic message doesn't contain expected text: %s", msg)
					}
					t.Logf("✅ Validation panic: %v", r)
				}
			}()

			qb := NewQdrantBuilder().HnswEf(tt.ef)
			if !tt.wantPanic {
				if qb.custom.DefaultHnswEf != tt.ef {
					t.Errorf("HnswEf = %d, want %d", qb.custom.DefaultHnswEf, tt.ef)
				}
				t.Logf("✅ Valid ef=%d accepted", tt.ef)
			}
		})
	}
}

// TestQdrantBuilder_ScoreThreshold_Validation 测试 ScoreThreshold 参数校验
func TestQdrantBuilder_ScoreThreshold_Validation(t *testing.T) {
	tests := []struct {
		name      string
		threshold float32
		wantPanic bool
	}{
		{"Valid: threshold=0.0", 0.0, false},
		{"Valid: threshold=0.5", 0.5, false},
		{"Valid: threshold=0.85", 0.85, false},
		{"Valid: threshold=1.0", 1.0, false},
		{"Invalid: threshold=-0.1", -0.1, true},
		{"Invalid: threshold=-1.0", -1.0, true},
		{"Invalid: threshold=1.1", 1.1, true},
		{"Invalid: threshold=2.0", 2.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.wantPanic && r == nil {
					t.Errorf("Expected panic for threshold=%f, but didn't panic", tt.threshold)
				}
				if !tt.wantPanic && r != nil {
					t.Errorf("Unexpected panic for threshold=%f: %v", tt.threshold, r)
				}
				if r != nil {
					msg := r.(string)
					if !strings.Contains(msg, "ScoreThreshold must be in [0, 1]") {
						t.Errorf("Panic message doesn't contain expected text: %s", msg)
					}
					t.Logf("✅ Validation panic: %v", r)
				}
			}()

			qb := NewQdrantBuilder().ScoreThreshold(tt.threshold)
			if !tt.wantPanic {
				if qb.custom.DefaultScoreThreshold != tt.threshold {
					t.Errorf("ScoreThreshold = %f, want %f", qb.custom.DefaultScoreThreshold, tt.threshold)
				}
				t.Logf("✅ Valid threshold=%f accepted", tt.threshold)
			}
		})
	}
}

// TestQdrantBuilder_ChainedValidation 测试链式调用中的参数校验
func TestQdrantBuilder_ChainedValidation(t *testing.T) {
	t.Run("Valid chain", func(t *testing.T) {
		qb := NewQdrantBuilder().
			HnswEf(512).
			ScoreThreshold(0.8).
			WithVector(false).
			Build()

		if qb.DefaultHnswEf != 512 {
			t.Errorf("DefaultHnswEf = %d, want 512", qb.DefaultHnswEf)
		}
		if qb.DefaultScoreThreshold != 0.8 {
			t.Errorf("DefaultScoreThreshold = %f, want 0.8", qb.DefaultScoreThreshold)
		}
		if qb.DefaultWithVector != false {
			t.Errorf("DefaultWithVector = %v, want false", qb.DefaultWithVector)
		}
		t.Log("✅ Valid chained configuration accepted")
	})

	t.Run("Invalid ef in chain", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for invalid ef")
			} else {
				t.Logf("✅ Validation panic in chain: %v", r)
			}
		}()

		NewQdrantBuilder().
			HnswEf(0). // ❌ Invalid
			ScoreThreshold(0.8).
			Build()
	})

	t.Run("Invalid threshold in chain", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for invalid threshold")
			} else {
				t.Logf("✅ Validation panic in chain: %v", r)
			}
		}()

		NewQdrantBuilder().
			HnswEf(512).
			ScoreThreshold(1.5). // ❌ Invalid
			Build()
	})
}
