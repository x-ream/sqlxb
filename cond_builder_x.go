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

type CondBuilderX struct {
	CondBuilder
}

// Sub subquery (type-safe subquery construction)
//
// # Used to build nested queries using xb's fluent API instead of handwritten SQL
//
// Parameters:
//   - s: SQL template, use ? as subquery placeholder
//   - f: subquery construction function
//
// Example:
//
//	// ✅ IN subquery
//	xb.Of("orders").
//	    Sub("user_id IN ?", func(sb *BuilderX) {
//	        sb.Of(&VipUser{}).Select("id")
//	    }).
//	    Build()
//	// Generates: SELECT * FROM orders WHERE user_id IN (SELECT id FROM vip_users)
//
//	// ✅ EXISTS subquery
//	xb.Of(&User{}).
//	    Sub("EXISTS ?", func(sb *BuilderX) {
//	        sb.Of(&Order{}).
//	           Select("1").
//	           X("orders.user_id = users.id")
//	    }).
//	    Build()
//	// Generates: SELECT * FROM users WHERE EXISTS (SELECT 1 FROM orders WHERE orders.user_id = users.id)
//
//	// ✅ Complex subquery
//	xb.Of(&Product{}).
//	    Sub("price > ?", func(sb *BuilderX) {
//	        sb.Of(&Product{}).
//	           Select("AVG(price)").
//	           Eq("category", "electronics")
//	    }).
//	    Build()
//	// Generates: SELECT * FROM products WHERE price > (SELECT AVG(price) FROM products WHERE category = ?)
//
// Advantages:
//   - Type-safe: uses xb API instead of string concatenation
//   - High readability: clear nested structure
//   - Maintainable: subqueries can also use Eq/In/X and other methods
func (x *CondBuilderX) Sub(s string, f func(sb *BuilderX)) *CondBuilderX {

	b := new(BuilderX)
	f(b)
	bb := Bb{
		Op:    SUB,
		Key:   s,
		Value: b,
	}
	x.bbs = append(x.bbs, bb)
	return x
}
