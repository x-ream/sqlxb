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

// Interceptor interceptor interface
// Used for infrastructure observation (logging, monitoring)
// Not used for business logic
type Interceptor interface {
	// Name interceptor name (for registration/unregistration)
	Name() string

	// BeforeBuild before Build() is executed
	// ⭐ Only pass Metadata, enforced at compile time to only set metadata
	// ⭐ Cannot modify query logic (type system ensures)
	// Return error can prevent Build()
	BeforeBuild(meta *Metadata) error

	// AfterBuild after Build() is executed
	// Used for observing generated SQL (logging, monitoring, auditing)
	// Return error can prevent subsequent execution
	AfterBuild(built interface{}) error
}
