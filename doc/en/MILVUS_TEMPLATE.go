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

//go:build ignore
// +build ignore

// ============================================================================
// Milvus 扩展模板（复制此文件到 to_milvus_json.go 开始实现）
// ============================================================================
//
// ⭐ 注意：此文件仅作为模板参考，不会被编译（build ignore）
//
// 这是一个完整的 Milvus 支持模板，展示如何基于 VectorDBRequest 接口
// 快速实现 Milvus 向量数据库的支持。
//
// 实现步骤：
//  1. 复制此文件为 to_milvus_json.go（移除 build ignore 标签）
//  2. 在 oper.go 添加 Milvus 操作符常量
//  3. 在 cond_builder_milvus.go 添加 Builder 方法
//  4. 运行测试验证
//
// ============================================================================

package xb

import (
	"encoding/json"
	"fmt"

	. "github.com/fndome/xb"
)

// ============================================================================
// Step 1: 在 oper.go 添加 Milvus 专属操作符
// ============================================================================

/*
在 oper.go 文件添加：

const (
	MILVUS_NPROBE     = "MILVUS_NPROBE"      // 搜索参数 nprobe
	MILVUS_ROUND_DEC  = "MILVUS_ROUND_DEC"   // 小数位四舍五入
	MILVUS_EF         = "MILVUS_EF"          // HNSW 搜索参数
	MILVUS_EXPR       = "MILVUS_EXPR"        // 过滤表达式
	MILVUS_XX         = "MILVUS_XX"          // 自定义参数
)
*/

// ============================================================================
// Step 2: 定义 Milvus 专属接口（继承 VectorDBRequest）
// ============================================================================

// MilvusRequest Milvus 专属请求接口
// 继承 VectorDBRequest，自动支持通用参数（ScoreThreshold, WithVector）
type MilvusRequest interface {
	VectorDBRequest // ⭐ 继承通用接口

	// Milvus 专属方法
	GetSearchParams() **MilvusSearchParams
	GetExpr() *string
}

// ============================================================================
// Step 3: 定义 Milvus 请求结构体
// ============================================================================

// MilvusSearchRequest Milvus 搜索请求
type MilvusSearchRequest struct {
	CollectionName string      `json:"collection_name"`
	Vectors        [][]float32 `json:"vectors"`
	TopK           int         `json:"topk"`
	MetricType     string      `json:"metric_type"`

	// ⭐ 通用字段（自动支持）
	ScoreThreshold *float32 `json:"score_threshold,omitempty"`
	OutputFields   []string `json:"output_fields,omitempty"` // WithVector 控制此字段

	// ⭐ Milvus 专属字段
	SearchParams *MilvusSearchParams `json:"search_params,omitempty"`
	Expr         string              `json:"expr,omitempty"`
}

// MilvusSearchParams Milvus 搜索参数
type MilvusSearchParams struct {
	NProbe   int `json:"nprobe,omitempty"`
	RoundDec int `json:"round_decimal,omitempty"`
	Ef       int `json:"ef,omitempty"` // HNSW 参数
}

// ============================================================================
// Step 4: 实现接口方法
// ============================================================================

// ⭐ 实现 VectorDBRequest（通用接口）

func (r *MilvusSearchRequest) GetScoreThreshold() **float32 {
	return &r.ScoreThreshold
}

func (r *MilvusSearchRequest) GetWithVector() *bool {
	// Milvus 通过 OutputFields 控制是否返回向量
	// 这里返回 nil 表示不支持直接设置 bool
	// 实际应用中需要在 applyMilvusParams 中处理
	return nil
}

func (r *MilvusSearchRequest) GetFilter() interface{} {
	return &r.Expr // Milvus 使用 Expr 字符串过滤
}

// ⭐ 实现 MilvusRequest（Milvus 专属接口）

func (r *MilvusSearchRequest) GetSearchParams() **MilvusSearchParams {
	return &r.SearchParams
}

func (r *MilvusSearchRequest) GetExpr() *string {
	return &r.Expr
}

// ============================================================================
// Step 5: 参数应用函数（复用通用逻辑）
// ============================================================================

// applyMilvusParams 应用 Milvus 专属参数
func applyMilvusParams(bbs []Bb, req MilvusRequest) {
	// ⭐ 第一层：复用通用参数应用
	ApplyCommonVectorParams(bbs, req)

	// ⭐ 第二层：应用 Milvus 专属参数
	for _, bb := range bbs {
		switch bb.Op {
		case "MILVUS_NPROBE": // 需要在 oper.go 定义
			ensureMilvusParams(req)
			(*req.GetSearchParams()).NProbe = bb.Value.(int)

		case "MILVUS_ROUND_DEC":
			ensureMilvusParams(req)
			(*req.GetSearchParams()).RoundDec = bb.Value.(int)

		case "MILVUS_EF":
			ensureMilvusParams(req)
			(*req.GetSearchParams()).Ef = bb.Value.(int)

		case "MILVUS_EXPR":
			expr := bb.Value.(string)
			*req.GetExpr() = expr
		}
	}
}

// ensureMilvusParams 确保 SearchParams 已初始化
func ensureMilvusParams(req MilvusRequest) {
	params := req.GetSearchParams()
	if *params == nil {
		*params = &MilvusSearchParams{}
	}
}

// ============================================================================
// Step 6: JSON 转换函数（在 Built 上，与 Qdrant 一致）
// ============================================================================

// JsonOfMilvusSelect 转换为 Milvus 搜索 JSON
// ⭐ 命名与 SQL 一致：JsonOfSelect (Milvus 版本)
// ⭐ 设计与 Qdrant 一致：在 Built 上调用，从 VectorSearch 参数中获取信息
//
// 返回:
//   - JSON 字符串
//   - error
//
// 示例:
//
//	built := C().
//	    VectorScoreThreshold(0.8).      // 通用参数
//	    MilvusNProbe(64).               // Milvus 专属
//	    MilvusExpr("age > 18").         // 过滤表达式
//	    MilvusX("consistency_level", "Strong"). // 自定义参数
//	    VectorSearch("users", "embedding", []float32{0.1, 0.2}, 10, L2Distance).
//	    Build()
//
//	json, err := built.JsonOfMilvusSelect()
func (built *Built) JsonOfMilvusSelect() (string, error) {
	// 1️⃣ 从 Built.Conds 中找到 VECTOR_SEARCH 参数
	vectorBb := findVectorSearchBb(built.Conds)
	if vectorBb == nil {
		return "", fmt.Errorf("no VECTOR_SEARCH found, use VectorSearch() to specify search parameters")
	}

	params := vectorBb.Value.(VectorSearchParams)

	// 2️⃣ 创建 Milvus 请求对象
	req := &MilvusSearchRequest{
		CollectionName: params.TableName,
		Vectors:        [][]float32{params.Vector},
		TopK:           params.Limit,
		MetricType:     milvusDistanceMetric(params.Distance),
	}

	// 3️⃣ 应用参数（自动处理通用 + 专属参数）
	applyMilvusParams(built.Conds, req)

	// 4️⃣ 序列化为 JSON（复用通用逻辑）
	return milvusMergeAndSerialize(req, built.Conds)
}

// findVectorSearchBb 从 Bb 数组中找到 VECTOR_SEARCH
func findVectorSearchBb(bbs []Bb) *Bb {
	for i := range bbs {
		if bbs[i].Op == VECTOR_SEARCH {
			return &bbs[i]
		}
	}
	return nil
}

// milvusDistanceMetric 转换距离度量
func milvusDistanceMetric(metric VectorDistance) string {
	switch metric {
	case CosineDistance:
		return "COSINE"
	case L2Distance:
		return "L2"
	case InnerProduct:
		return "IP"
	default:
		return "L2"
	}
}

// milvusMergeAndSerialize 合并自定义参数并序列化
// ⭐ 这个函数和 Qdrant 的 mergeAndSerialize 逻辑完全一致
func milvusMergeAndSerialize(req interface{}, bbs []Bb) (string, error) {
	// ⭐ 复用通用提取函数
	customParams := ExtractCustomParams(bbs, "MILVUS_XX")

	if len(customParams) == 0 {
		// 无自定义参数，直接序列化
		bytes, err := json.MarshalIndent(req, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal Milvus request: %w", err)
		}
		return string(bytes), nil
	}

	// 有自定义参数，先序列化为 map，再添加自定义字段
	bytes, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Milvus request: %w", err)
	}

	var reqMap map[string]interface{}
	if err := json.Unmarshal(bytes, &reqMap); err != nil {
		return "", fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	// 添加用户自定义参数
	for k, v := range customParams {
		reqMap[k] = v
	}

	// 重新序列化
	finalBytes, err := json.MarshalIndent(reqMap, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal final JSON: %w", err)
	}

	return string(finalBytes), nil
}

// ============================================================================
// Step 7: 在 cond_builder_milvus.go 添加 Builder 方法
// ============================================================================

/*
创建 cond_builder_milvus.go 文件：

package xb

// ⭐ 通用参数（已在 cond_builder.go 实现）
// func (b *CondBuilder) VectorScoreThreshold(threshold float32)
// func (b *CondBuilder) VectorWithVector(withVector bool)

// MilvusNProbe 设置 Milvus nprobe 搜索参数
func (b *CondBuilder) MilvusNProbe(nprobe int) *CondBuilder {
	return b.append(Bb{op: MILVUS_NPROBE, value: nprobe})
}

// MilvusRoundDec 设置 Milvus 小数位四舍五入
func (b *CondBuilder) MilvusRoundDec(dec int) *CondBuilder {
	return b.append(Bb{op: MILVUS_ROUND_DEC, value: dec})
}

// MilvusEf 设置 Milvus HNSW ef 参数
func (b *CondBuilder) MilvusEf(ef int) *CondBuilder {
	return b.append(Bb{op: MILVUS_EF, value: ef})
}

// MilvusExpr 设置 Milvus 过滤表达式
func (b *CondBuilder) MilvusExpr(expr string) *CondBuilder {
	return b.append(Bb{op: MILVUS_EXPR, value: expr})
}

// MilvusX 用户自定义 Milvus 参数（类似 Qdrant 的 QdrantX）
//
// 示例:
//   MilvusX("consistency_level", "Strong")
//   MilvusX("travel_timestamp", 12345)
func (b *CondBuilder) MilvusX(key string, value interface{}) *CondBuilder {
	return b.append(Bb{op: MILVUS_XX, key: key, value: value})
}
*/

// ============================================================================
// Step 8: 测试示例
// ============================================================================

/*
创建 to_milvus_json_test.go 文件：

package xb

import (
	"encoding/json"
	"testing"
)

func TestMilvusSearchRequest_Interface(t *testing.T) {
	req := &MilvusSearchRequest{}

	// ✅ 验证实现了 VectorDBRequest
	var _ VectorDBRequest = req

	// ✅ 验证实现了 MilvusRequest
	var _ MilvusRequest = req
}

func TestJsonOfMilvusSelect(t *testing.T) {
	// ⭐ 与 SQL 命名一致的调用方式
	built := C().
		VectorScoreThreshold(0.8).
		MilvusNProbe(64).
		MilvusExpr("age > 18").
		MilvusX("consistency_level", "Strong").
		VectorSearch("users", "embedding", []float32{0.1, 0.2, 0.3}, 10, L2Distance).
		Build()

	jsonStr, err := built.JsonOfMilvusSelect()
	if err != nil {
		t.Fatalf("JsonOfMilvusSelect failed: %v", err)
	}

	// 验证 JSON 结构
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// 验证基本字段
	if result["collection_name"] != "users" {
		t.Errorf("collection_name = %v, want 'users'", result["collection_name"])
	}

	if result["topk"] != 10 {
		t.Errorf("topk = %v, want 10", result["topk"])
	}

	if result["metric_type"] != "L2" {
		t.Errorf("metric_type = %v, want 'L2'", result["metric_type"])
	}

	// 验证通用参数
	if result["score_threshold"] != 0.8 {
		t.Errorf("score_threshold = %v, want 0.8", result["score_threshold"])
	}

	// 验证 Milvus 专属参数
	searchParams := result["search_params"].(map[string]interface{})
	if searchParams["nprobe"] != 64 {
		t.Errorf("nprobe = %v, want 64", searchParams["nprobe"])
	}

	if result["expr"] != "age > 18" {
		t.Errorf("expr = %v, want 'age > 18'", result["expr"])
	}

	// 验证自定义参数
	if result["consistency_level"] != "Strong" {
		t.Errorf("consistency_level = %v, want 'Strong'", result["consistency_level"])
	}
}
*/

// ============================================================================
// 总结
// ============================================================================

/*
通过这个模板，Milvus 用户只需：

✅ 5 个步骤（定义操作符 → 定义接口 → 实现方法 → 应用参数 → 序列化）
✅ 自动复用通用逻辑（ApplyCommonVectorParams, extractCustomParams）
✅ 代码零重复（通用参数、自定义参数、JSON 序列化全部复用）
✅ 类型安全（编译时检查）
✅ 优雅的 API（像 Qdrant 一样流畅）

估计代码量：
- to_milvus_json.go: ~200 行
- cond_builder_milvus.go: ~50 行
- to_milvus_json_test.go: ~100 行
总计：~350 行（vs Qdrant 的 800 行，代码量减少 56%）

核心原因：复用了通用接口和函数！🎉
*/
