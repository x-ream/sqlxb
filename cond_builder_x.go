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

// Sub 子查询（类型安全的子查询构建）
//
// 用于构建嵌套查询，使用 xb 的链式 API 而不是手写 SQL
//
// 参数：
//   - s: SQL 模板，使用 ? 作为子查询占位符
//   - f: 子查询构建函数
//
// 示例：
//
//	// ✅ IN 子查询
//	xb.Of("orders").
//	    Sub("user_id IN ?", func(sb *BuilderX) {
//	        sb.Of(&VipUser{}).Select("id")
//	    }).
//	    Build()
//	// 生成: SELECT * FROM orders WHERE user_id IN (SELECT id FROM vip_users)
//
//	// ✅ EXISTS 子查询
//	xb.Of(&User{}).
//	    Sub("EXISTS ?", func(sb *BuilderX) {
//	        sb.Of(&Order{}).
//	           Select("1").
//	           X("orders.user_id = users.id")
//	    }).
//	    Build()
//	// 生成: SELECT * FROM users WHERE EXISTS (SELECT 1 FROM orders WHERE orders.user_id = users.id)
//
//	// ✅ 复杂子查询
//	xb.Of(&Product{}).
//	    Sub("price > ?", func(sb *BuilderX) {
//	        sb.Of(&Product{}).
//	           Select("AVG(price)").
//	           Eq("category", "electronics")
//	    }).
//	    Build()
//	// 生成: SELECT * FROM products WHERE price > (SELECT AVG(price) FROM products WHERE category = ?)
//
// 优势：
//   - 类型安全：使用 xb API 而不是字符串拼接
//   - 可读性高：嵌套结构清晰
//   - 可维护：子查询也能用 Eq/In/X 等方法
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
