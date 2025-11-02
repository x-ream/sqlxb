# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2025-01-XX

### 🎨 Complete API Unification (Major Improvement over v1.1.0)

**Design Philosophy**: Don't add concepts to solve problems. Less is more.

### Added
- **Unified Insert API**: `Insert(func(ib *InsertBuilder))` works for both SQL and vector databases
- **Smart Format Detection**: QdrantCustom automatically handles Insert(func) format
- **Vector Array Support**: `[]float32` and `[]float64` preserved in InsertBuilder/UpdateBuilder
- **Convenience Methods**:
  - `Built.SqlOfUpsert()` - MySQL UPSERT without Custom
  - `Built.SqlOfInsertIgnore()` - MySQL INSERT IGNORE without Custom

### Removed (Simplification)
- **Preset Constructors** (complexity removed):
  - ❌ `QdrantHighPrecision()`, `QdrantHighSpeed()`, `QdrantBalanced()`
  - ❌ `MySQLWithUpsert()`, `MySQLWithIgnore()`
  - ❌ `InsertPoint()`, `InsertPoints()` methods
  - ❌ `Delete()` method (not needed)

### Changed
- `JsonOfDelete()` now auto-sets `built.Delete = true` internally
- Only basic constructors remain: `NewQdrantCustom()`, `NewMySQLCustom()`
- Users configure via fields or existing closures (`QdrantX()`)

### Why v1.2.0 (Not v1.1.1)?

v1.1.0 had design issues:
- Too many preset functions (5 removed)
- Inconsistent API (InsertPoint vs Insert)
- Over-engineered solutions

v1.2.0 achieves true simplicity:
- ✅ One `Insert(func)` API for all databases
- ✅ One `Update(func)` API for all databases
- ✅ No extra methods needed
- ✅ Complete unification

### Tests
- 140+ tests, all passing
- New tests for unified Insert/Update/Delete API
- Comprehensive Qdrant CRUD validation

### Documentation
- All docs updated to reflect simplified design
- Removed references to preset constructors
- Added unified API examples

---

## [1.1.0] - 2025-01-XX (Deprecated - Use v1.2.0)

### Added
- **Custom Interface**: Unified abstraction for database-specific features
  - `Custom` interface with single `Generate(built *Built) (interface{}, error)` method
  - Supports both SQL databases (returns `*SQLResult`) and vector databases (returns `string` JSON)
- **Full CRUD for Vector Databases**:
  - `JsonOfInsert()` - Generate INSERT JSON for vector databases
  - `JsonOfUpdate()` - Generate UPDATE JSON for vector databases
  - `JsonOfDelete()` - Generate DELETE JSON for vector databases
  - `JsonOfSelect()` - Existing SELECT JSON (now unified through Custom interface)
- **Official Custom Implementations**:
  - `QdrantCustom` - Full CRUD support for Qdrant (Insert/Update/Delete/Search)
  - `MySQLCustom` - MySQL-specific features (UPSERT, INSERT IGNORE)
- **Public API Additions**:
  - `SqlData()` - Generate data query SQL (public, for Custom implementations)
  - `SqlCount()` - Generate COUNT SQL (public, for Custom implementations)
  - `SqlInsert()` - Generate INSERT SQL (public, for Custom implementations)
  - `SQLResult` struct - Encapsulates SQL, Args, CountSQL, and Meta map
- **Meta Map Enhancement**:
  - `Meta map[string]string` in `SQLResult` for ORM field mapping
  - Supports alias mapping (`AS xxx`) and table prefixes (`table.field`)

### Changed
- **Breaking**: `PageCondition` fields are now public:
  - `page` → `Page`
  - `rows` → `Rows`
  - `last` → `Last`
  - `isTotalRowsIgnored` → `IsTotalRowsIgnored`
- `Custom()` method renamed from `WithCustom()` for more fluent API
- `SqlOfPage()`, `SqlOfSelect()`, `SqlOfInsert()`, `SqlOfUpdate()`, `SqlOfDelete()` now check `built.Custom` and delegate to Custom implementation if available

### Deprecated
- None (all existing APIs remain compatible)

### Documentation
- Added comprehensive Custom interface guides:
  - `CUSTOM_INTERFACE_README.md` - Overview and navigation
  - `CUSTOM_INTERFACE_PHILOSOPHY.md` - Design philosophy
  - `CUSTOM_QUICKSTART.md` - Quick start guide
  - `CUSTOM_VECTOR_DB_GUIDE.md` - Vector database implementation guide
  - `XB_ORM_PHILOSOPHY.md` - xb's role as ORM complement
  - `QDRANT_FULL_CRUD_SUMMARY.md` - Qdrant CRUD implementation summary

### Tests
- Added 8 new tests for Qdrant CRUD operations (all passing)
- Total: 130+ tests, all passing
- Validated Custom interface architecture through complete CRUD implementation

---

## [1.0.0] - 2024-XX-XX

### Added
- Production-ready release
- Stable API for SQL and vector database queries
- Full support for PostgreSQL, MySQL, Qdrant
- Comprehensive documentation and examples

---

## [0.10.0] - 2024-XX-XX

### Added
- **Qdrant Advanced API**:
  - Recommend API for personalized recommendations
  - Discover API for theme exploration
  - Scroll API for large dataset traversal
  - Functional parameters for unified builder style
- 100% backward compatibility with existing features

---

## [0.9.0] - 2024-XX-XX

### Added
- Vector database support (PostgreSQL pgvector, Qdrant)
- `VectorSearch()` method for vector similarity search
- Diversity strategies (MMR, min-distance, hash-based)
- Qdrant-specific parameters (`QdrantX()`, `QdrantBuilderX`)

---

## Earlier Versions

See git history for earlier version changes.

[1.1.0]: https://github.com/fndome/xb/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/fndome/xb/compare/v0.10.0...v1.0.0
[0.10.0]: https://github.com/fndome/xb/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/fndome/xb/releases/tag/v0.9.0

