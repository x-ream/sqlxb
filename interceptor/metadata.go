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
package interceptor

import "time"

// Metadata query metadata
// Does not affect query logic, only for observation
type Metadata struct {
	// Trace related
	TraceID   string `json:"trace_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`

	// User related
	UserID   int64 `json:"user_id,omitempty"`
	TenantID int64 `json:"tenant_id,omitempty"`

	// Performance related
	StartTime time.Time `json:"start_time,omitempty"`

	// Extension point (custom metadata)
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// Set set custom metadata
func (m *Metadata) Set(key string, value interface{}) {
	if m.Custom == nil {
		m.Custom = make(map[string]interface{})
	}
	m.Custom[key] = value
}

// Get get custom metadata
func (m *Metadata) Get(key string) interface{} {
	if m.Custom == nil {
		return nil
	}
	return m.Custom[key]
}

// GetString get string type metadata
func (m *Metadata) GetString(key string) string {
	v := m.Get(key)
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// GetInt64 get int64 type metadata
func (m *Metadata) GetInt64(key string) int64 {
	v := m.Get(key)
	if v == nil {
		return 0
	}
	if i, ok := v.(int64); ok {
		return i
	}
	return 0
}

// GetFloat64 get float64 type metadata
func (m *Metadata) GetFloat64(key string) float64 {
	v := m.Get(key)
	if v == nil {
		return 0
	}
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

// GetBool get bool type metadata
func (m *Metadata) GetBool(key string) bool {
	v := m.Get(key)
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}
