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

import "sync"

var (
	globalInterceptors []Interceptor
	mu                 sync.RWMutex
)

// Register register global interceptor
// Interceptors are executed in registration order
//
// Example:
//
//	interceptor.Register(&LoggingInterceptor{})
//	interceptor.Register(&PrometheusInterceptor{})
func Register(i Interceptor) {
	mu.Lock()
	defer mu.Unlock()
	globalInterceptors = append(globalInterceptors, i)
}

// Unregister uninstall interceptor (by name)
//
// Example:
//
//	interceptor.Unregister("logging")
func Unregister(name string) {
	mu.Lock()
	defer mu.Unlock()

	for i, interceptor := range globalInterceptors {
		if interceptor.Name() == name {
			globalInterceptors = append(globalInterceptors[:i], globalInterceptors[i+1:]...)
			break
		}
	}
}

// Clear clear all interceptors
// Mainly used for test environment
func Clear() {
	mu.Lock()
	defer mu.Unlock()
	globalInterceptors = []Interceptor{}
}

// GetAll get all interceptors (read only)
// Mainly used for internal use
func GetAll() []Interceptor {
	mu.RLock()
	defer mu.RUnlock()

	// Return copy to avoid external modification
	result := make([]Interceptor, len(globalInterceptors))
	copy(result, globalInterceptors)
	return result
}
