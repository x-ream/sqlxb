// Copyright 2020 io.xream.sqlxb
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

// Metadata 查询元数据
// 不影响查询逻辑，只用于观察
type Metadata struct {
	// 追踪相关
	TraceID   string `json:"trace_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`

	// 用户相关
	UserID   int64 `json:"user_id,omitempty"`
	TenantID int64 `json:"tenant_id,omitempty"`

	// 性能相关
	StartTime time.Time `json:"start_time,omitempty"`

	// 扩展点（自定义元数据）
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// Set 设置自定义元数据
func (m *Metadata) Set(key string, value interface{}) {
	if m.Custom == nil {
		m.Custom = make(map[string]interface{})
	}
	m.Custom[key] = value
}

// Get 获取自定义元数据
func (m *Metadata) Get(key string) interface{} {
	if m.Custom == nil {
		return nil
	}
	return m.Custom[key]
}

// GetString 获取字符串类型的元数据
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

// GetInt64 获取 int64 类型的元数据
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

// GetFloat64 获取 float64 类型的元数据
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

// GetBool 获取 bool 类型的元数据
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

