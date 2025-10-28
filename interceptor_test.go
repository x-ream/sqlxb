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
	"time"

	"github.com/fndome/xb/interceptor"
)

// 测试元数据设置
func TestInterceptor_Metadata(t *testing.T) {
	// 清空拦截器
	interceptor.Clear()

	// 注册测试拦截器
	testInterceptor := &TestMetadataInterceptor{}
	interceptor.Register(testInterceptor)

	defer interceptor.Clear() // 测试后清空

	// 构建查询
	built := Of(&Product{}).
		Eq("name", "test").
		Build()

	// 验证元数据
	if built.Meta == nil {
		t.Fatal("Meta should not be nil")
	}

	if built.Meta.TraceID != "test-trace-123" {
		t.Errorf("Expected TraceID 'test-trace-123', got '%s'", built.Meta.TraceID)
	}

	if built.Meta.UserID != 999 {
		t.Errorf("Expected UserID 999, got %d", built.Meta.UserID)
	}

	endpoint := built.Meta.GetString("endpoint")
	if endpoint != "/api/products" {
		t.Errorf("Expected endpoint '/api/products', got '%s'", endpoint)
	}

	t.Logf("✅ Metadata 设置成功")
	t.Logf("   TraceID: %s", built.Meta.TraceID)
	t.Logf("   UserID: %d", built.Meta.UserID)
	t.Logf("   Custom.endpoint: %s", endpoint)
}

// 测试 AfterBuild 观察 SQL
func TestInterceptor_AfterBuild(t *testing.T) {
	interceptor.Clear()

	sqlLogger := &SQLLoggingInterceptor{}
	interceptor.Register(sqlLogger)

	defer interceptor.Clear()

	// 构建查询
	built := Of(&Product{}).
		Eq("name", "test").
		Gt("price", 100).
		Build()

	sql, args, _ := built.SqlOfSelect()

	// 验证 SQL 被记录
	if !strings.Contains(sqlLogger.LastSQL, "SELECT") {
		t.Errorf("Expected SQL to be logged")
	}

	if len(sqlLogger.LastArgs) != 2 {
		t.Errorf("Expected 2 args, got %d", len(sqlLogger.LastArgs))
	}

	t.Logf("✅ AfterBuild 执行成功")
	t.Logf("   SQL: %s", sql)
	t.Logf("   Args: %v", args)
	t.Logf("   Logged SQL: %s", sqlLogger.LastSQL)
}

// 测试多个拦截器按顺序执行
func TestInterceptor_Order(t *testing.T) {
	interceptor.Clear()

	var executionOrder []string

	// 拦截器 1
	interceptor.Register(&OrderTestInterceptor{
		name:  "first",
		order: &executionOrder,
	})

	// 拦截器 2
	interceptor.Register(&OrderTestInterceptor{
		name:  "second",
		order: &executionOrder,
	})

	defer interceptor.Clear()

	// 构建查询
	Of(&Product{}).Eq("name", "test").Build()

	// 验证执行顺序
	if len(executionOrder) != 4 {
		t.Fatalf("Expected 4 executions, got %d", len(executionOrder))
	}

	expected := []string{"first:before", "second:before", "first:after", "second:after"}
	for i, name := range expected {
		if executionOrder[i] != name {
			t.Errorf("Expected %s at position %d, got %s", name, i, executionOrder[i])
		}
	}

	t.Logf("✅ 拦截器按顺序执行")
	t.Logf("   执行顺序: %v", executionOrder)
}

// 测试拦截器编译时类型限制
func TestInterceptor_TypeSafety(t *testing.T) {
	interceptor.Clear()

	// ⭐ 这个测试验证编译时限制
	// BeforeBuild 只能接收 *Metadata
	// 无法访问 BuilderX 的查询方法

	safeInterceptor := &TypeSafeInterceptor{}
	interceptor.Register(safeInterceptor)

	defer interceptor.Clear()

	built := Of(&Product{}).
		Eq("name", "test").
		Build()

	// 验证只设置了元数据，没有修改查询
	sql, args, _ := built.SqlOfSelect()

	if strings.Contains(sql, "tenant_id") {
		t.Errorf("Interceptor should NOT be able to add tenant_id condition")
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 arg (name), got %d", len(args))
	}

	t.Logf("✅ 类型安全验证通过：Interceptor 无法修改查询")
	t.Logf("   SQL: %s", sql)
	t.Logf("   Args: %v", args)
}

// 测试注册和卸载
func TestInterceptor_RegisterUnregister(t *testing.T) {
	interceptor.Clear()

	i1 := &TestMetadataInterceptor{}
	i2 := &SQLLoggingInterceptor{}

	interceptor.Register(i1)
	interceptor.Register(i2)

	all := interceptor.GetAll()
	if len(all) != 2 {
		t.Fatalf("Expected 2 interceptors, got %d", len(all))
	}

	// 卸载一个
	interceptor.Unregister("test-metadata")

	all = interceptor.GetAll()
	if len(all) != 1 {
		t.Fatalf("Expected 1 interceptor after unregister, got %d", len(all))
	}

	if all[0].Name() != "sql-logging" {
		t.Errorf("Expected remaining interceptor 'sql-logging', got '%s'", all[0].Name())
	}

	// 清空所有
	interceptor.Clear()

	all = interceptor.GetAll()
	if len(all) != 0 {
		t.Fatalf("Expected 0 interceptors after clear, got %d", len(all))
	}

	t.Logf("✅ 注册和卸载测试通过")
}

// ===== 测试辅助类型 =====

// TestMetadataInterceptor 测试元数据设置
type TestMetadataInterceptor struct{}

func (t *TestMetadataInterceptor) Name() string {
	return "test-metadata"
}

func (t *TestMetadataInterceptor) BeforeBuild(meta *interceptor.Metadata) error {
	meta.TraceID = "test-trace-123"
	meta.UserID = 999
	meta.TenantID = 888
	meta.StartTime = time.Now()
	meta.Set("endpoint", "/api/products")
	meta.Set("method", "GET")
	return nil
}

func (t *TestMetadataInterceptor) AfterBuild(built interface{}) error {
	return nil
}

// SQLLoggingInterceptor 测试 SQL 日志
type SQLLoggingInterceptor struct {
	LastSQL  string
	LastArgs []interface{}
}

func (s *SQLLoggingInterceptor) Name() string {
	return "sql-logging"
}

func (s *SQLLoggingInterceptor) BeforeBuild(meta *interceptor.Metadata) error {
	meta.StartTime = time.Now()
	return nil
}

func (s *SQLLoggingInterceptor) AfterBuild(built interface{}) error {
	if b, ok := built.(*Built); ok {
		sql, args, _ := b.SqlOfSelect()
		s.LastSQL = sql
		s.LastArgs = args
	}
	return nil
}

// OrderTestInterceptor 测试执行顺序
type OrderTestInterceptor struct {
	name  string
	order *[]string
}

func (o *OrderTestInterceptor) Name() string {
	return o.name
}

func (o *OrderTestInterceptor) BeforeBuild(meta *interceptor.Metadata) error {
	*o.order = append(*o.order, o.name+":before")
	return nil
}

func (o *OrderTestInterceptor) AfterBuild(built interface{}) error {
	*o.order = append(*o.order, o.name+":after")
	return nil
}

// TypeSafeInterceptor 验证类型安全
type TypeSafeInterceptor struct{}

func (t *TypeSafeInterceptor) Name() string {
	return "type-safe"
}

func (t *TypeSafeInterceptor) BeforeBuild(meta *interceptor.Metadata) error {
	// ✅ 只能设置元数据
	meta.UserID = 123
	meta.Set("test", "value")

	// ❌ 无法调用查询方法（编译时限制）
	// meta.Eq("tenant_id", 123)  // 编译错误
	// meta.VectorSearch(...)      // 编译错误

	return nil
}

func (t *TypeSafeInterceptor) AfterBuild(built interface{}) error {
	return nil
}
