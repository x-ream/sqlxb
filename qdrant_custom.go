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
// QdrantCustom: Qdrant 数据库专属配置
// ============================================================================

// QdrantCustom Qdrant 专属配置实现
//
// 实现 Custom 接口，提供 Qdrant 的默认配置和预设模式
type QdrantCustom struct {
	// 默认参数（如果用户没有显式指定，使用这些默认值）
	DefaultHnswEf         int     // 默认 HNSW EF 参数
	DefaultScoreThreshold float32 // 默认相似度阈值
	DefaultWithVector     bool    // 默认是否返回向量
}

// NewQdrantCustom 创建 Qdrant Custom（默认配置）
func NewQdrantCustom() *QdrantCustom {
	return &QdrantCustom{
		DefaultHnswEf:         128,
		DefaultScoreThreshold: 0.0,
		DefaultWithVector:     true,
	}
}

// ToJSON 实现 Custom 接口
// ⭐ 这是唯一需要实现的方法
func (c *QdrantCustom) ToJSON(built *Built) (string, error) {
	// ⭐ 委托给内部实现
	return built.toQdrantJSON()
}

// ============================================================================
// 便捷构造函数
// ============================================================================

// QdrantHighPrecision 高精度模式（慢，但准确）
func QdrantHighPrecision() *QdrantCustom {
	return &QdrantCustom{
		DefaultHnswEf:        512,
		DefaultScoreThreshold: 0.85,
		DefaultWithVector:    true,
	}
}

// QdrantHighSpeed 高速模式（快，但可能不太准）
func QdrantHighSpeed() *QdrantCustom {
	return &QdrantCustom{
		DefaultHnswEf:        32,
		DefaultScoreThreshold: 0.5,
		DefaultWithVector:    false,
	}
}

// QdrantBalanced 平衡模式（默认）
func QdrantBalanced() *QdrantCustom {
	return NewQdrantCustom()
}

