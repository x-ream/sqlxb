# xb  (Extensible Builder)
[![OSCS Status](https://www.oscs1024.com/platform/badge/fndome/xb.svg?size=small)](https://www.oscs1024.com/project/fndome/xb?ref=badge_small)
![workflow build](https://github.com/fndome/xb/actions/workflows/go.yml/badge.svg)
[![GitHub tag](https://img.shields.io/github/tag/fndome/xb.svg?style=flat)](https://github.com/fndome/xb/tags)
[![Go Report Card](https://goreportcard.com/badge/github.com/fndome/xb)](https://goreportcard.com/report/github.com/fndome/xb)



**AI-First SQL Or JSON Builder** for Relational and Vector Databases

A tool of sql or json query builder, build sql for sql.DB, [sqlx](https://github.com/jmoiron/sqlx), [gorp](https://github.com/go-gorp/gorp),
or build condition sql for some orm framework, like [xorm](https://github.com/go-xorm/xorm), [gorm](https://github.com/go-gorm/gorm)....
also can build json for some json parameter db, like [Qdrant](https://github.com/qdrant/qdrant) ....


> 🎉 **v1.1.0 Released**: Custom Interface for Database-Specific Features! Full CRUD support for vector databases.

---

## 🚀 NEW: Custom Interface (v1.1.0)

**Unified abstraction for SQL and Vector Databases with database-specific features!**

**✨ New in v1.1.0**:
- 🎯 **Custom Interface** - Unified abstraction for all database types
- 📝 **Full CRUD** - Insert/Update/Delete support for vector databases
- 🔧 **Official Implementations** - QdrantCustom (full CRUD), MySQLCustom (UPSERT)
- 🏗️ **Extensible Architecture** - One `Generate()` method handles all operations
- 📚 **Complete Documentation** - Templates and guides for custom implementations

```go
// MySQL UPSERT (v1.1.0) - 无需 Custom
built := xb.Of(user).
    Insert(func(ib *xb.InsertBuilder) {
        ib.Set("id", user.ID).
           Set("name", user.Name).
           Set("email", user.Email)
    }).
    Build()
sql, args := built.SqlOfUpsert()
// INSERT INTO users ... ON DUPLICATE KEY UPDATE ...

// Qdrant Vector Search (v1.1.0)
// 方式1: 使用默认 Custom
built := xb.Of(&CodeVector{}).
    Custom(xb.NewQdrantCustom()).
    Eq("language", "golang").
    VectorSearch("embedding", queryVector, 10).
    Build()
json, _ := built.JsonOfSelect()

// 方式2: 使用 QdrantX 闭包配置（推荐）
built := xb.Of(&CodeVector{}).
    Custom(xb.NewQdrantCustom()).
    Eq("language", "golang").
    VectorSearch("embedding", queryVector, 10).
    QdrantX(func(qx *xb.QdrantBuilderX) {
        qx.HnswEf(512).ScoreThreshold(0.85)
    }).
    Build()
json, _ := built.JsonOfSelect()

// Standard SQL (no Custom needed)
built := xb.Of(&User{}).
    Eq("status", 1).
    Gt("age", 18).
    Build()
sql, args, _ := built.SqlOfSelect()
// SELECT * FROM users WHERE status = ? AND age > ?
```

📖 **[Read the Custom Interface Guide →](./doc/CUSTOM_INTERFACE_README.md)**

**Architecture Highlights**:
- ✅ One interface method for all operations (Select/Insert/Update/Delete)
- ✅ Supports both SQL databases (MySQL, Oracle) and vector databases (Qdrant, Milvus)
- ✅ Type-safe: `SQLResult` for SQL, `string` for JSON
- ✅ Easy to extend: Implement your own database in minutes

---

## 🔍 Qdrant Advanced API (since v0.10.0)

**The first unified ORM for both Relational and Vector Databases!**

**✨ New in v0.10.0**:
- 🎯 **Recommend API** - Personalized recommendations with positive/negative samples
- 🔍 **Discover API** - Explore common themes from user context
- 🔄 **Scroll API** - Efficient traversal for large datasets
- 🎨 **Functional Parameters** - Unified builder style
- 🔧 **100% Backward Compatible** - All existing features preserved

```go
// MySQL (existing)
xb.Of(&Order{}).Eq("status", 1).Build().SqlOfSelect()

// VectorDB (v0.10.0) - Same API!
xb.Of(&CodeVector{}).
    Eq("language", "golang").
    VectorSearch("embedding", queryVector, 10).
    QdrantX(func(qx *QdrantBuilderX) {
        qx.Recommend(func(rb *RecommendBuilder) {
            rb.Positive(123, 456).Limit(20)
        })
    }).
    Build()
```

📖 **[Read the Vector Database Design Docs →](./doc/VECTOR_README.md)**

**Features**:
- ✅ Unified API for MySQL + VectorDB
- ✅ Type-safe ORM for vectors
- ✅ Auto-optimized hybrid queries
- ✅ 100% backward compatible

**Development**: AI-First approach (Claude AI + Human review)

---

## 🤖 AI-First Development

xb v0.8.0+ is developed using an innovative **AI-First** approach:

- 🤖 **AI Assistant (Claude via Cursor)**: Architecture design, code implementation, testing, documentation
- 👨‍💻 **Human Maintainer**: Code review, strategic decisions, critical algorithm oversight

### Maintenance Model (80/15/5)

- **80%** of code: AI independently maintains (simple, clear patterns)
- **15%** of code: AI assists, human reviews (medium complexity)
- **5%** of code: Human leads, AI assists (critical algorithms like `from_builder_optimization.go`)

### v0.8.1 Vector Database Support

**Achieved entirely through AI-First development**:
- Architecture & Design: AI Assistant (Claude)
- Code Implementation: AI Assistant (763 lines)
- Testing: AI Assistant (13 test cases, 100% passing)
- Documentation: AI Assistant (120+ pages)
- Review & Approval: Human Maintainer

This makes xb **one of the first major Go ORM projects successfully maintained by AI**.

---

## Program feature:
* ignore building nil or empty string

## Available field of struct:
    
* base: string, *bool, *int64, *float64, time.Time....
* json: struct, map, array, slice
* bytes: []byte

## Example

    SELECT * FROM t_cat WHERE id > ? AND (price >= ? OR is_sold = ?)

    var Db *sqlx.DB
    ....

	var c Cat
	builder := xb.Of(&c).Gt("id", 10000).And(func(cb *CondBuilder) {
		cb.Gte("price", catRo.Price).OR().Eq("is_sold", catRo.IsSold)
    })

    countSql, dataSql, vs, _ := builder.Build().SqlOfPage()
    var catList []Cat
	err = Db.Select(&catList, dataSql, vs...)


## 📚 Documentation

**[Complete Documentation Index →](./doc/README.md)**

Quick links:
- [Vector Database Quick Start](./doc/VECTOR_QUICKSTART.md)
- [Vector Diversity + Qdrant Guide](./doc/VECTOR_DIVERSITY_QDRANT.md)
- [All Filtering Mechanisms](./doc/ALL_FILTERING_MECHANISMS.md)
- [Custom Vector DB Guide](./doc/CUSTOM_VECTOR_DB_GUIDE.md)
- [Custom JOINs Guide](./doc/CUSTOM_JOINS_GUIDE.md)
- [Contributors](./doc/CONTRIBUTORS.md)

**AI Application Ecosystem**:
- **[AI Application Docs →](./doc/ai_application/README.md)** - Complete AI/RAG/Agent integration guide
- [AI Agent Toolkit](./doc/ai_application/AGENT_TOOLKIT.md) - JSON Schema, OpenAPI
- [RAG Best Practices](./doc/ai_application/RAG_BEST_PRACTICES.md) - Document retrieval guide
- [LangChain Integration](./doc/ai_application/LANGCHAIN_INTEGRATION.md) - Python LangChain
- [Performance Optimization](./doc/ai_application/PERFORMANCE.md) - AI app tuning

**Complete Application Examples**:
- **[Examples →](./examples/README.md)** - Full working applications
- [PostgreSQL + pgvector App](./examples/pgvector-app/) - Code search
- [Qdrant Integration App](./examples/qdrant-app/) - Document retrieval
- [RAG Application](./examples/rag-app/) - Full RAG system
- [PageIndex App](./examples/pageindex-app/) - Structured document retrieval

## Contributing

We warmly welcome all forms of contributions! 🎉

- 🐛 **Report bugs**: [GitHub Issues](https://github.com/fndome/xb/issues)
- 💡 **Request features**: [GitHub Issues](https://github.com/fndome/xb/issues)
- 💬 **Discuss ideas**: [GitHub Discussions](https://github.com/fndome/xb/discussions)
- 💻 **Submit code**: See [CONTRIBUTING](./doc/CONTRIBUTING.md)

> In the era of rapid tech iteration, we embrace change and listen to the community. See [VISION.md](./VISION.md) for our development philosophy.

## Quickstart

* [Single Example](#single-example)
* [Join Example](#join-example)


### Single Example

```Go

import (
    . "github.com/fndome/xb"
)

type Cat struct {
	Id       uint64    `db:"id"`
	Name     string    `db:"name"`
	Age      uint      `db:"age"`
	Color    string    `db:"color"`
	Weight   float64   `db:"weight"`
	IsSold   *bool     `db:"is_sold"`
	Price    *float64  `db:"price"`
	CreateAt time.Time `db:"create_at"`
}

func (*Cat) TableName() string {
	return "t_cat"
}

// IsSold, Price, fields can be zero, must be pointer, like Java Boolean....
// xb has func: Bool(true), Int(v) ....
// xb no relect, not support omitempty, should rewrite ro, dto
type CatRo struct {
	Name   string   `json:"name, string"`
	IsSold *bool    `json:"isSold, *bool"`
	Price  *float64 `json:"price, *float64"`
	Age    uint     `json:"age", unit`
}

func main() {
	cat := Cat{
		Id:       100002,
		Name:     "Tuanzi",
		Age:      1,
		Color:    "B",
		Weight:   8.5,
		IsSold:   Bool(true),
		Price:    Float64(10000.00),
		CreateAt: time.Now(),
	}
    // INSERT .....

    // PREPARE TO QUERY
	catRo := CatRo{
		Name:	"Tu",
		IsSold: nil,
		Price:  Float64(5000.00),
		Age:    1,
	}

	preCondition := func() bool {
		if cat.Color == "W" {
			return true
		} else if cat.Weight <= 3 {
			return false
		} else {
			return true
		}
	}

	var c Cat
	var builder = Of(&c)
	builder.LikeLeft("name",catRo.Name)
	builder.X("weight <> ?", 0) //X(k, v...), hardcode func, value 0 and nil will NOT ignore
    //Eq,Ne,Gt.... value 0 and nil will ignore, like as follow: OR().Eq("is_sold", catRo.IsSold)
	builder.And(func(cb *CondBuilder) {
            cb.Gte("price", catRo.Price).OR().Gte("age", catRo.Age).OR().Eq("is_sold", catRo.IsSold))
	    })
    //func Bool NOT designed for value nil or 0; designed to convert complex logic to bool
    //Decorator pattern suggest to use func Bool preCondition, like:
    //myBoolDecorator := NewMyBoolDecorator(para)
    //builder.Bool(myBoolDecorator.fooCondition, func(cb *CondBuilder) {
	builder.Bool(preCondition, func(cb *CondBuilder) {
            cb.Or(func(cb *CondBuilder) {
                cb.Lt("price", 5000)
            })
	})
	builder.Sort("id", ASC)
        builder.Paged(func(pb *PageBuilder) {
                pb.Page(1).Rows(10).IgnoreTotalRows()
            })
    countSql, dataSql, vs, _ := builder.Build().SqlOfPage()
    // ....

    //dataSql: SELECT * FROM t_cat WHERE id > ? AND name LIKE ? AND weight <> 0 AND (price >= ? OR age >= ?) OR (price < ?)
    //ORDER BY id ASC LIMIT 10

	//.IgnoreTotalRows(), will not output countSql
    //countSql: SELECT COUNT(*) FROM t_cat WHERE name LIKE ? AND weight <> 0 AND (price >= ? OR age >= ?) OR (price < ?)
    
    //sqlx: 	err = Db.Select(&catList, dataSql,vs...)
	joinSql, condSql, cvs := builder.Build().SqlOfCond()
    
    //conditionSql: id > ? AND name LIKE ? AND weight <> 0 AND (price >= ? OR age >= ?) OR (price < ?)

}
```


### Join Example

```Go
import (
        . "github.com/fndome/xb"
    )
    
func main() {
	
	sub := func(sb *BuilderX) {
                sb.Select("id","type").From("t_pet").Gt("id", 10000) //....
            }
	
        builder := X().
		Select("p.id","p.weight").
		FromX(func(fb *FromBuilder) {
                    fb.
                        Sub(sub).As("p").
                        JOIN(INNER).Of("t_dog").As("d").On("d.pet_id = p.id").
                        JOIN(LEFT).Of("t_cat").As("c").On("c.pet_id = p.id").
                            Cond(func(on *ON) {
                                on.Gt("c.id", ro.MinCatId)
                            })
		    }).
	        Ne("p.type","PIG").
                Having(func(cb *CondBuilderX) {
                    cb.Sub("p.weight > ?", func(sb *BuilderX) {
                        sb.Select("AVG(weight)").From("t_dog")
                    })
                })
    
}


```

---

## 🎯 Use Case Decision Guide

**Get direct answers without learning — Let AI decide for you**

> 📖 **[中文版 (Chinese Version) →](./doc/USE_CASE_GUIDE_ZH.md)**

### Scenario 1️⃣: Semantic Search & Personalization

**Use Vector Database (pgvector / Qdrant)**

```
Applicable Use Cases:
  ✅ Product recommendations ("Users who bought A also liked...")
  ✅ Code search ("Find similar function implementations")
  ✅ Customer service ("Find similar historical tickets")
  ✅ Content recommendations ("Similar articles, videos")
  ✅ Image search ("Find similar images")

Characteristics:
  - Fragmented data (each record independent)
  - Requires similarity matching
  - No clear structure

Example:
  xb.Of(&Product{}).
      VectorSearch("embedding", userVector, 20).
      Eq("category", "electronics")
```

---

### Scenario 2️⃣: Structured Long Document Analysis

**Use PageIndex**

```
Applicable Use Cases:
  ✅ Financial report analysis ("How is financial stability in 2024?")
  ✅ Legal contract retrieval ("Chapter 3 breach of contract terms")
  ✅ Technical manual queries ("Which page contains installation steps?")
  ✅ Academic paper reading ("Methodology section content")
  ✅ Policy document analysis ("Specific provisions in Section 2.3")

Characteristics:
  - Long documents (50+ pages)
  - Clear chapter structure
  - Context preservation required

Example:
  xb.Of(&PageIndexNode{}).
      Eq("doc_id", docID).
      Like("title", "Financial Stability").
      Eq("level", 1)
```

---

### Scenario 3️⃣: Hybrid Retrieval (Structure + Semantics)

**Use PageIndex + Vector Database**

```
Applicable Use Cases:
  ✅ Research report Q&A ("Investment advice for tech sector")
  ✅ Knowledge base retrieval (need both structure and semantics)
  ✅ Medical literature analysis ("Treatment plan related chapters")
  ✅ Patent search ("Patents with similar technical solutions")

Characteristics:
  - Both structured and semantic needs
  - Long documents + precise matching requirements

Example:
  // Step 1: PageIndex locates chapter
  xb.Of(&PageIndexNode{}).
      Like("title", "Investment Advice").
      Eq("level", 2)
  
  // Step 2: Vector search within chapter
  xb.Of(&DocumentChunk{}).
      VectorSearch("embedding", queryVector, 10).
      Gte("page", chapterStartPage).
      Lte("page", chapterEndPage)
```

---

### Scenario 4️⃣: Traditional Business Data

**Use Standard SQL (No Vector/PageIndex needed)**

```
Applicable Use Cases:
  ✅ User management ("Find users over 18")
  ✅ Order queries ("Orders in January 2024")
  ✅ Inventory management ("Products with low stock")
  ✅ Statistical reports ("Sales by region")

Characteristics:
  - Structured data
  - Exact condition matching
  - No semantic understanding needed

Example:
  xb.Of(&User{}).
      Gte("age", 18).
      Eq("status", "active").
      Paged(...)
```

---

## 🤔 Quick Decision Tree

```
Your data is...

├─ Fragmented (products, users, code snippets)
│  └─ Need "similarity" matching?
│     ├─ Yes → Vector Database ✅
│     └─ No  → Standard SQL ✅
│
└─ Long documents (reports, manuals, contracts)
   └─ Has clear chapter structure?
      ├─ Yes → PageIndex ✅
      │  └─ Also need semantic matching?
      │     └─ Yes → PageIndex + Vector ✅
      └─ No → Traditional RAG (chunking + vector) ✅
```

---

## 💡 Core Principles

```
Don't debate technology choices — Look at data characteristics:

1️⃣ Fragmented data + need similarity
   → Vector Database

2️⃣ Long documents + structured + need chapter location
   → PageIndex

3️⃣ Long documents + unstructured + need semantics
   → Traditional RAG (chunking + vector)

4️⃣ Structured data + exact matching
   → Standard SQL

5️⃣ Complex scenarios
   → Hybrid approach
```

**xb supports all scenarios — One API for everything!** ✅


