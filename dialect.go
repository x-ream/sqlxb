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

// ============================================================================
// Custom 接口：数据库专属配置（核心抽象）
// ============================================================================

// Custom 数据库专属配置接口
// 每个向量数据库实现自己的 Custom，通过接口多态实现不同的行为
//
// 设计原则：
//  - ✅ 接口即标识：QdrantCustom 实现了 Custom，本身就说明是 Qdrant
//  - ✅ 多态调用：built.JsonOfSelect() 自动调用对应实现的 ToJSON()
//  - ✅ 无需枚举：不需要 Dialect 枚举，Go 的接口自动处理
//  - ✅ 简单直接：一个方法搞定，草根到底
//
// 实现示例：
//
//	// Qdrant 实现
//	type QdrantCustom struct {
//	    DefaultHnswEf int
//	}
//
//	func (c *QdrantCustom) ToJSON(built *Built) (string, error) {
//	    // Qdrant 的 JSON 生成逻辑
//	}
//
//	// Milvus 实现
//	type MilvusCustom struct {
//	    DefaultNProbe int
//	}
//
//	func (c *MilvusCustom) ToJSON(built *Built) (string, error) {
//	    // Milvus 的 JSON 生成逻辑
//	}
//
// 使用示例：
//
//	// Qdrant
//	built := xb.C().
//	    WithCustom(&xb.QdrantCustom{DefaultHnswEf: 256}).
//	    Build()
//
//	json, _ := built.JsonOfSelect()  // ← 自动调用 QdrantCustom.ToJSON()
//
//	// Milvus
//	built := xb.C().
//	    WithCustom(&xb.MilvusCustom{DefaultNProbe: 64}).
//	    Build()
//
//	json, _ := built.JsonOfSelect()  // ← 自动调用 MilvusCustom.ToJSON()
type Custom interface {
	// ToJSON 生成查询 JSON
	// 参数:
	//   - built: Built 对象（包含所有查询条件）
	// 返回:
	//   - JSON 字符串
	//   - error
	//
	// 说明:
	//   - 每个数据库的实现不同（QdrantCustom vs MilvusCustom）
	//   - Go 的接口多态自动调用对应的实现
	//   - 不需要额外的类型判断或枚举
	ToJSON(built *Built) (string, error)
}

// ============================================================================
// 说明
// ============================================================================

// Custom 接口非常简单，只有一个方法：ToJSON()
//
// 为什么不需要更多方法？
//  - GetDialect()？不需要，类型本身就是标识（QdrantCustom vs MilvusCustom）
//  - ApplyParams()？不需要，在 ToJSON() 内部处理
//  - Insert/Update/Delete？如果需要，可以添加新接口继承 Custom
//
// 这就是 Go 的哲学：简单、直接、实用
//
// 示例：添加 Insert 支持
//
//	type CustomWithInsert interface {
//	    Custom
//	    ToInsertJSON(built *Built) (string, error)
//	}
//
//	// Milvus 支持 Insert
//	func (c *MilvusCustom) ToInsertJSON(built *Built) (string, error) {
//	    // Milvus 插入逻辑
//	}
//
//	// 使用时类型断言
//	if ci, ok := built.Custom.(CustomWithInsert); ok {
//	    json, _ := ci.ToInsertJSON(built)
//	}
