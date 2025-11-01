# 向量数据库扩展指南

**让 Milvus/Weaviate/Pinecone 用户轻松实现自定义 Builder** 🎯

## 📋 目录

1. [设计理念](#设计理念)
2. [核心接口](#核心接口)
3. [快速开始：5步添加 Milvus 支持](#快速开始5步添加-milvus-支持)
4. [完整示例：Milvus Builder](#完整示例milvus-builder)
5. [复用通用逻辑](#复用通用逻辑)
6. [测试模板](#测试模板)

---

## 设计理念

### 三层架构

```
┌─────────────────────────────────────────────────────┐
│  xb 通用层（所有数据库共享）                            │
│  - Bb 结构体                                         │
│  - VectorDBRequest 接口                              │
│  - ApplyCommonVectorParams 函数                      │
└─────────────────────────────────────────────────────┘
                        ↓ 继承
┌─────────────────────────────────────────────────────┐
│  数据库专属层（Qdrant/Milvus/Weaviate...）           │
│  - QdrantRequest 接口（继承 VectorDBRequest）        │
│  - MilvusRequest 接口（继承 VectorDBRequest）        │
└─────────────────────────────────────────────────────┘
                        ↓ 实现
┌─────────────────────────────────────────────────────┐
│  请求结构体                                          │
│  - QdrantSearchRequest（实现 QdrantRequest）         │
│  - MilvusSearchRequest（实现 MilvusRequest）         │
└─────────────────────────────────────────────────────┘
```

### 设计优势

✅ **通用参数自动复用**：`ScoreThreshold`, `WithVector` 等通用参数，所有数据库自动支持  
✅ **专属参数灵活扩展**：每个数据库可以添加自己的专属参数（如 Qdrant 的 `HnswEf`）  
✅ **代码零重复**：通用逻辑写一次，所有数据库共享  
✅ **类型安全**：Go 接口保证编译时类型检查

---

## 核心接口

### 1️⃣ VectorDBRequest（所有数据库通用）

```go
// 定义在 vector_db_request.go
type VectorDBRequest interface {
    GetScoreThreshold() **float32  // 相似度阈值（通用）
    GetWithVector() *bool          // 是否返回向量（通用）
    GetFilter() interface{}        // 过滤器（类型各异）
}
```

### 2️⃣ ApplyCommonVectorParams（通用参数应用）

```go
// 所有数据库自动复用
func ApplyCommonVectorParams(bbs []Bb, req VectorDBRequest) {
    // 自动处理 QDRANT_SCORE_THRESHOLD, QDRANT_WITH_VECTOR
}
```

---

## 快速开始：5步添加 Milvus 支持

### Step 1: 定义 Milvus 专属操作符

```go
// 在 oper.go 添加
const (
    MILVUS_NPROBE     = "MILVUS_NPROBE"      // 搜索参数 nprobe
    MILVUS_ROUND_DEC  = "MILVUS_ROUND_DEC"   // 小数位四舍五入
    MILVUS_EXPR       = "MILVUS_EXPR"        // 过滤表达式
    MILVUS_XX         = "MILVUS_XX"          // 自定义参数
)
```

### Step 2: 定义 Milvus 专属接口

```go
// 在 to_milvus_json.go 创建
type MilvusRequest interface {
    VectorDBRequest  // ⭐ 继承通用接口

    // Milvus 专属方法
    GetSearchParams() **MilvusSearchParams
    GetExpr() *string
}
```

### Step 3: 定义请求结构体

```go
type MilvusSearchRequest struct {
    CollectionName string          `json:"collection_name"`
    Vectors        [][]float32     `json:"vectors"`
    TopK           int              `json:"topk"`
    MetricType     string           `json:"metric_type"`
    
    // ⭐ 通用字段
    ScoreThreshold *float32        `json:"score_threshold,omitempty"`
    WithVector     bool             `json:"output_fields"`
    
    // ⭐ Milvus 专属字段
    SearchParams   *MilvusSearchParams `json:"search_params,omitempty"`
    Expr           string           `json:"expr,omitempty"`
}

type MilvusSearchParams struct {
    NProbe   int  `json:"nprobe,omitempty"`
    RoundDec int  `json:"round_decimal,omitempty"`
}
```

### Step 4: 实现接口方法

```go
// 实现 VectorDBRequest（通用接口）
func (r *MilvusSearchRequest) GetScoreThreshold() **float32 {
    return &r.ScoreThreshold
}

func (r *MilvusSearchRequest) GetWithVector() *bool {
    return &r.WithVector
}

func (r *MilvusSearchRequest) GetFilter() interface{} {
    return &r.Expr  // Milvus 使用 Expr 表达式过滤
}

// 实现 MilvusRequest（Milvus 专属接口）
func (r *MilvusSearchRequest) GetSearchParams() **MilvusSearchParams {
    return &r.SearchParams
}

func (r *MilvusSearchRequest) GetExpr() *string {
    return &r.Expr
}
```

### Step 5: 实现参数应用函数

```go
// 应用 Milvus 专属参数
func applyMilvusParams(bbs []Bb, req MilvusRequest) {
    // ⭐ 第一层：复用通用参数应用
    ApplyCommonVectorParams(bbs, req)

    // ⭐ 第二层：应用 Milvus 专属参数
    for _, bb := range bbs {
        switch bb.op {
        case MILVUS_NPROBE:
            ensureMilvusParams(req)
            (*req.GetSearchParams()).NProbe = bb.value.(int)

        case MILVUS_ROUND_DEC:
            ensureMilvusParams(req)
            (*req.GetSearchParams()).RoundDec = bb.value.(int)

        case MILVUS_EXPR:
            expr := bb.value.(string)
            *req.GetExpr() = expr
        }
    }
}

func ensureMilvusParams(req MilvusRequest) {
    params := req.GetSearchParams()
    if *params == nil {
        *params = &MilvusSearchParams{}
    }
}
```

---

## 完整示例：Milvus Builder

### 用户 API（与 Qdrant 完全一致）

```go
import "github.com/fndo-io/xb"

// ⭐ 与 Qdrant 完全一致的调用方式
built := xb.C().
    // ⭐ 通用参数（自动支持）
    VectorScoreThreshold(0.8).      // 相似度阈值
    VectorWithVector(true).         // 返回向量
    
    // ⭐ Milvus 专属参数
    MilvusNProbe(64).               // 搜索参数
    MilvusRoundDec(2).              // 小数位
    MilvusExpr("age > 18").         // 过滤表达式
    
    // ⭐ 自定义参数（像 Qdrant 的 QdrantX）
    MilvusX("consistency_level", "Strong").
    MilvusX("travel_timestamp", 12345).
    
    // ⭐ 向量搜索参数（与 Qdrant 一致）
    VectorSearch("my_collection", "embedding", []float32{0.1, 0.2, 0.3}, 10, xb.L2Distance).
    Build()

// ⭐ 转换为 JSON（与 SQL 命名一致）
json, err := built.JsonOfMilvusSelect()
```

**对比 SQL 和 Qdrant**:

```go
// SQL
built := xb.C().
    Eq("language", "golang").
    Gt("score", 0.8).
    Build()

sql, args, _ := built.SqlOfSelect()  // ← SQL 查询

// Qdrant
built := xb.C().
    VectorScoreThreshold(0.8).
    QdrantHnswEf(128).
    VectorSearch("code_vectors", "embedding", vec, 10, xb.CosineDistance).
    Build()

json, err := built.JsonOfQdrantSelect()  // ← Qdrant JSON（统一命名）

// Milvus（完全一致的命名）
built := xb.C().
    VectorScoreThreshold(0.8).
    MilvusNProbe(64).
    VectorSearch("code_vectors", "embedding", vec, 10, xb.L2Distance).
    Build()

json, err := built.JsonOfMilvusSelect()  // ← Milvus JSON（统一命名）
```

### Builder 函数（简单封装）

```go
// ⭐ 通用参数（已在 cond_builder.go 实现）
// func (b *CondBuilder) VectorScoreThreshold(threshold float32)
// func (b *CondBuilder) VectorWithVector(withVector bool)

// ⭐ Milvus 专属参数（新增）
func (b *CondBuilder) MilvusNProbe(nprobe int) *CondBuilder {
    return b.append(Bb{op: MILVUS_NPROBE, value: nprobe})
}

func (b *CondBuilder) MilvusRoundDec(dec int) *CondBuilder {
    return b.append(Bb{op: MILVUS_ROUND_DEC, value: dec})
}

func (b *CondBuilder) MilvusExpr(expr string) *CondBuilder {
    return b.append(Bb{op: MILVUS_EXPR, value: expr})
}

func (b *CondBuilder) MilvusX(key string, value interface{}) *CondBuilder {
    return b.append(Bb{op: MILVUS_XX, value: map[string]interface{}{key: value}})
}
```

### 转换为 JSON（与 SQL 命名一致）

```go
// ⭐ 与 SQL 命名一致：JsonOfMilvusSelect (类似 SqlOfSelect)
func (built *Built) JsonOfMilvusSelect() (string, error) {
    // 1️⃣ 从 Built.Conds 中找到 VECTOR_SEARCH 参数
    vectorBb := findVectorSearchBb(built.Conds)
    if vectorBb == nil {
        return "", fmt.Errorf("no VECTOR_SEARCH found")
    }

    params := vectorBb.Value.(VectorSearchParams)

    // 2️⃣ 创建请求对象
    req := &MilvusSearchRequest{
        CollectionName: params.TableName,
        Vectors:        [][]float32{params.Vector},
        TopK:           params.Limit,
        MetricType:     milvusDistanceMetric(params.Distance),
    }

    // 3️⃣ 应用参数（自动处理通用 + 专属参数）
    applyMilvusParams(built.Conds, req)

    // 4️⃣ 序列化为 JSON
    return milvusMergeAndSerialize(req, built.Conds)
}

// 辅助函数
func findVectorSearchBb(bbs []Bb) *Bb {
    for i := range bbs {
        if bbs[i].Op == VECTOR_SEARCH {
            return &bbs[i]
        }
    }
    return nil
}

func milvusDistanceMetric(metric VectorDistance) string {
    switch metric {
    case CosineDistance: return "COSINE"
    case L2Distance: return "L2"
    case InnerProduct: return "IP"
    default: return "L2"
    }
}
```

---

## 复用通用逻辑

### 1️⃣ 参数应用复用

```go
// ⭐ Qdrant 复用
func applyQdrantParams(bbs []Bb, req QdrantRequest) {
    ApplyCommonVectorParams(bbs, req)  // 复用通用逻辑
    // ... Qdrant 专属逻辑
}

// ⭐ Milvus 复用
func applyMilvusParams(bbs []Bb, req MilvusRequest) {
    ApplyCommonVectorParams(bbs, req)  // 复用通用逻辑
    // ... Milvus 专属逻辑
}
```

### 2️⃣ JSON 序列化复用

```go
// 在 to_qdrant_json.go 已实现
func mergeAndSerialize(req interface{}, bbs []Bb, customOp string) (string, error) {
    // 1. 提取自定义参数（QDRANT_XX / MILVUS_XX）
    customParams := extractCustomParams(bbs, customOp)
    
    // 2. 序列化请求对象
    bytes, _ := json.Marshal(req)
    
    // 3. 合并自定义参数
    // 4. 返回最终 JSON
}

// ⭐ Milvus 直接复用
return mergeAndSerialize(req, built.Conds, MILVUS_XX)
```

### 3️⃣ 自定义参数提取复用

```go
// 通用提取函数（改造 extractQdrantCustomParams）
func extractCustomParams(bbs []Bb, customOp string) map[string]interface{} {
    result := make(map[string]interface{})
    for _, bb := range bbs {
        if bb.op == customOp {  // QDRANT_XX / MILVUS_XX
            if m, ok := bb.value.(map[string]interface{}); ok {
                for k, v := range m {
                    result[k] = v
                }
            }
        }
    }
    return result
}
```

---

## 测试模板

### 单元测试

```go
func TestMilvusSearchRequest_Interface(t *testing.T) {
    req := &MilvusSearchRequest{}

    // ✅ 验证实现了 VectorDBRequest
    var _ VectorDBRequest = req

    // ✅ 验证实现了 MilvusRequest
    var _ MilvusRequest = req
}

func TestToMilvusSearchJSON(t *testing.T) {
    json, err := C().
        VectorScoreThreshold(0.8).
        MilvusNProbe(64).
        MilvusExpr("age > 18").
        MilvusX("consistency_level", "Strong").
        ToMilvusSearchJSON("users", [][]float32{{0.1, 0.2}}, 10, "L2")

    assert.NoError(t, err)
    
    // 验证 JSON 结构
    var result map[string]interface{}
    json.Unmarshal([]byte(json), &result)
    
    assert.Equal(t, 0.8, result["score_threshold"])
    assert.Equal(t, 64, result["search_params"].(map[string]interface{})["nprobe"])
    assert.Equal(t, "age > 18", result["expr"])
    assert.Equal(t, "Strong", result["consistency_level"])  // 自定义参数
}
```

---

## 文件组织建议

```
xb/
├── vector_db_request.go         # ⭐ 通用接口（所有数据库共享）
├── to_qdrant_json.go            # Qdrant 实现
├── to_milvus_json.go            # Milvus 实现（新增）
├── to_weaviate_json.go          # Weaviate 实现（新增）
├── cond_builder.go              # Builder 基础
├── cond_builder_vector.go       # 通用向量参数 Builder
├── cond_builder_milvus.go       # Milvus 专属 Builder（新增）
└── oper.go                      # 所有操作符常量
```

---

## 总结

### 对于 Milvus 用户

✅ **5个步骤**即可添加完整 Milvus 支持  
✅ **通用参数自动继承**（ScoreThreshold, WithVector）  
✅ **专属参数灵活扩展**（NProbe, RoundDec）  
✅ **自定义参数优雅支持**（MilvusX）  
✅ **零重复代码**（复用 ApplyCommonVectorParams, mergeAndSerialize）

### 对于框架维护者

✅ **接口驱动设计**：新数据库只需实现接口  
✅ **通用逻辑复用**：参数应用、JSON 序列化全部复用  
✅ **易于维护**：每个数据库独立文件，互不影响  
✅ **类型安全**：编译时检查，运行时零错误

---

**下一步建议**：

1. 将 `mergeAndSerialize` 改为通用函数（接受 `customOp` 参数）
2. 将 `extractQdrantCustomParams` 改为 `extractCustomParams`
3. 创建 `to_milvus_json.go` 模板文件
4. 添加通用的 Builder 方法（`VectorScoreThreshold`, `VectorWithVector`）

需要我实现这些改进吗？ 🚀

