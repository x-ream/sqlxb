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

// TestInterceptor_Metadata test metadata setting
func TestInterceptor_Metadata(t *testing.T) {
	// Clear interceptor
	interceptor.Clear()

	// Register test interceptor
	testInterceptor := &TestMetadataInterceptor{}
	interceptor.Register(testInterceptor)

	defer interceptor.Clear() // Test after clear

	// Build query
	built := Of(&Product{}).
		Eq("name", "test").
		Build()

	// Verify metadata
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

	t.Logf("✅ Metadata setting successful")
	t.Logf("   TraceID: %s", built.Meta.TraceID)
	t.Logf("   UserID: %d", built.Meta.UserID)
	t.Logf("   Custom.endpoint: %s", endpoint)
}

// TestInterceptor_AfterBuild test AfterBuild observe SQL
func TestInterceptor_AfterBuild(t *testing.T) {
	interceptor.Clear()

	sqlLogger := &SQLLoggingInterceptor{}
	interceptor.Register(sqlLogger)

	defer interceptor.Clear()

	// Build query
	built := Of(&Product{}).
		Eq("name", "test").
		Gt("price", 100).
		Build()

	sql, args, _ := built.SqlOfSelect()

	// Verify SQL is recorded
	if !strings.Contains(sqlLogger.LastSQL, "SELECT") {
		t.Errorf("Expected SQL to be logged")
	}

	if len(sqlLogger.LastArgs) != 2 {
		t.Errorf("Expected 2 args, got %d", len(sqlLogger.LastArgs))
	}

	t.Logf("✅ AfterBuild executed successfully")
	t.Logf("   SQL: %s", sql)
	t.Logf("   Args: %v", args)
	t.Logf("   Logged SQL: %s", sqlLogger.LastSQL)
}

// TestInterceptor_Order test multiple interceptors execute in order
func TestInterceptor_Order(t *testing.T) {
	interceptor.Clear()

	var executionOrder []string

	// Interceptor 1
	interceptor.Register(&OrderTestInterceptor{
		name:  "first",
		order: &executionOrder,
	})

	// Interceptor 2
	interceptor.Register(&OrderTestInterceptor{
		name:  "second",
		order: &executionOrder,
	})

	defer interceptor.Clear()

	// Build query
	Of(&Product{}).Eq("name", "test").Build()

	// Verify execution order
	if len(executionOrder) != 4 {
		t.Fatalf("Expected 4 executions, got %d", len(executionOrder))
	}

	expected := []string{"first:before", "second:before", "first:after", "second:after"}
	for i, name := range expected {
		if executionOrder[i] != name {
			t.Errorf("Expected %s at position %d, got %s", name, i, executionOrder[i])
		}
	}

	t.Logf("✅ Interceptors execute in order")
	t.Logf("   Execution order: %v", executionOrder)
}

// TestInterceptor_TypeSafety test interceptor compile time type restriction
func TestInterceptor_TypeSafety(t *testing.T) {
	interceptor.Clear()

	// ⭐ This test verifies compile time restriction
	// BeforeBuild 只能接收 *Metadata
	// Cannot access BuilderX's query methods

	safeInterceptor := &TypeSafeInterceptor{}
	interceptor.Register(safeInterceptor)

	defer interceptor.Clear()

	built := Of(&Product{}).
		Eq("name", "test").
		Build()

	// Verify only metadata is set, no query modification
	sql, args, _ := built.SqlOfSelect()

	if strings.Contains(sql, "tenant_id") {
		t.Errorf("Interceptor should NOT be able to add tenant_id condition")
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 arg (name), got %d", len(args))
	}

	t.Logf("✅ Type safety verification passed: Interceptor cannot modify query")
	t.Logf("   SQL: %s", sql)
	t.Logf("   Args: %v", args)
}

// TestInterceptor_RegisterUnregister test register and unregister
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

	// Unregister one
	interceptor.Unregister("test-metadata")

	all = interceptor.GetAll()
	if len(all) != 1 {
		t.Fatalf("Expected 1 interceptor after unregister, got %d", len(all))
	}

	if all[0].Name() != "sql-logging" {
		t.Errorf("Expected remaining interceptor 'sql-logging', got '%s'", all[0].Name())
	}

	// Clear all
	interceptor.Clear()

	all = interceptor.GetAll()
	if len(all) != 0 {
		t.Fatalf("Expected 0 interceptors after clear, got %d", len(all))
	}

	t.Logf("✅ Register and unregister test passed")
}

// ===== Test helper types =====

// TestMetadataInterceptor test metadata setting
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

// SQLLoggingInterceptor test SQL logging
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

// OrderTestInterceptor test execution order
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

// TypeSafeInterceptor test type safety
type TypeSafeInterceptor struct{}

func (t *TypeSafeInterceptor) Name() string {
	return "type-safe"
}

func (t *TypeSafeInterceptor) BeforeBuild(meta *interceptor.Metadata) error {
	// ✅ Only metadata can be set (compile time restriction)
	meta.UserID = 123
	meta.Set("test", "value")

	// ❌ Cannot call query methods
	// meta.Eq("tenant_id", 123)  // Compile error
	// meta.VectorSearch(...)      // Compile error

	return nil
}

func (t *TypeSafeInterceptor) AfterBuild(built interface{}) error {
	return nil
}
