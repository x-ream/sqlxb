# 从 sqlxb 迁移到 xb

## 🔄 项目重命名通知

**最新版本**: v0.11.0  
**日期**: 2025-10-28

### 变更历史

#### v0.11.0 (2025-10-28)
- ⚠️ **GitHub 组织迁移**: `x-ream/xb` → `fndome/xb`
- 模块路径: `github.com/fndome/xb`
- 仓库地址: `https://github.com/fndome/xb`

#### v0.10.5 (2025-10-28)
- 包名变更: `sqlxb` → `xb`
- 模块路径: `github.com/x-ream/xb`

---

## 📋 为什么重命名？

1. **更简洁的名称** - `xb` 比 `sqlxb` 更短、更易记
2. **品牌统一** - 统一到 fndome 组织
3. **更好的可发现性** - 短名称在搜索和推荐时更具优势
4. **包名统一** - 模块名 `xb` 与包名 `xb` 保持一致（v0.10.5）

---

## 🚀 迁移步骤

### 1️⃣ 更新 `go.mod`

**修改前**:
```go
module your-project

require (
    github.com/x-ream/sqlxb v0.10.4
)
```

**修改后**:
```go
module your-project

require (
    github.com/fndome/xb v0.11.0  // ⚠️ 组织已迁移
)
```

---

### 2️⃣ 更新 import 语句

**修改前**:
```go
import (
    "github.com/x-ream/sqlxb"  // 旧组织 + 旧包名
)
```

**修改后（v0.11.0）**:
```go
import (
    "github.com/fndome/xb"  // ✅ 新组织 + 新模块名
)

// 代码中使用 xb 包名
builder := xb.Of(&User{})  // ✅ 包名已改为 xb
```

---

### 3️⃣ 包名已更改（⚠️ 破坏性变更）

⚠️ **需要修改代码** - 包名从 `sqlxb` 改为 `xb`：

```go
// ❌ 旧代码（不再有效）
builder := sqlxb.Of(&User{})

// ✅ 新代码
builder := xb.Of(&User{})
```

---

## 🔧 批量替换命令

### Linux / macOS / Git Bash
```bash
# 更新 go.mod
find . -name "go.mod" -type f -exec sed -i 's|github.com/x-ream/sqlxb|github.com/x-ream/xb|g' {} +

# 更新所有 Go 文件
find . -name "*.go" -type f -exec sed -i 's|github.com/x-ream/sqlxb|github.com/x-ream/xb|g' {} +

# 运行 go mod tidy
go mod tidy
```

### Windows PowerShell
```powershell
# 更新 go.mod
Get-ChildItem -Recurse -Filter "go.mod" | ForEach-Object {
    (Get-Content $_.FullName) -replace 'github.com/x-ream/sqlxb', 'github.com/x-ream/xb' | Set-Content $_.FullName
}

# 更新所有 Go 文件
Get-ChildItem -Recurse -Filter "*.go" | ForEach-Object {
    (Get-Content $_.FullName) -replace 'github.com/x-ream/sqlxb', 'github.com/x-ream/xb' | Set-Content $_.FullName
}

# 运行 go mod tidy
go mod tidy
```

---

## ✅ 验证迁移

### 1. 确认依赖更新
```bash
go list -m all | grep "fndome"
```

**期望输出**:
```
github.com/fndome/xb v0.11.0
```

### 2. 运行测试
```bash
go test ./...
```

### 3. 构建项目
```bash
go build ./...
```

---

## 📌 常见问题

### Q1: 旧版本的 `sqlxb` 还能用吗？

**A**: 可以。旧的仓库会保留，但不再维护：
- `github.com/x-ream/sqlxb` - 保留到 v0.10.4
- `github.com/x-ream/xb` - 保留到 v0.10.5

建议迁移到新组织：`github.com/fndome/xb v0.11.0`

---

### Q2: 我需要修改代码中的 `sqlxb` 包名吗？

**A**: **需要！** 从 v0.10.5 开始，包名已改为 `xb`：

```go
import (
    "github.com/fndome/xb"  // ✅ v0.11.0 新组织路径
)

// ⚠️ 需要修改所有代码
// ❌ 旧代码
builder := sqlxb.Of(&User{})

// ✅ 新代码
builder := xb.Of(&User{})
```

**批量替换**:
```bash
# 1. 替换 import 路径
find . -name "*.go" -type f -exec sed -i 's|github.com/x-ream/xb|github.com/fndome/xb|g' {} +

# 2. 替换包名（如果还在用 sqlxb）
find . -name "*.go" -type f -exec sed -i 's/sqlxb\./xb\./g' {} +
```

---

### Q3: 我的项目使用了 v0.10.4 之前的版本，怎么办？

**A**: 分三步迁移：

1. **先升级到 v0.10.4**（仍使用 `sqlxb`，旧组织）
2. **再升级到 v0.10.5**（切换到 `xb`，旧组织）
3. **最后升级到 v0.11.0**（新组织）

```bash
# Step 1: 升级到 v0.10.4（包名 sqlxb）
go get github.com/x-ream/sqlxb@v0.10.4
go mod tidy

# Step 2: 升级到 v0.10.5（包名 xb，组织 x-ream）
go get github.com/x-ream/xb@v0.10.5
# 批量替换: sqlxb. → xb.

# Step 3: 升级到 v0.11.0（组织 fndome）
go get github.com/fndome/xb@v0.11.0
# 批量替换: github.com/x-ream/xb → github.com/fndome/xb
```

---

### Q4: 我使用了 `replace` 指令怎么办？

**A**: 更新 `go.mod` 中的 `replace` 指令：

**修改前**:
```go
replace github.com/x-ream/xb => /path/to/local/xb
```

**修改后（v0.11.0）**:
```go
replace github.com/fndome/xb => /path/to/local/xb
```

---

## 🔗 相关资源

- **GitHub 仓库**: https://github.com/fndome/xb
- **文档**: https://github.com/fndome/xb/blob/main/README.md
- **Roadmap**: https://github.com/fndome/xb/blob/main/doc/ROADMAP_v1.0.0.md
- **Issues**: https://github.com/fndome/xb/issues

### 旧仓库（只读）
- **x-ream/sqlxb**: https://github.com/x-ream/sqlxb (保留到 v0.10.4)
- **x-ream/xb**: https://github.com/x-ream/xb (保留到 v0.10.5)

---

## 💬 需要帮助？

如果您在迁移过程中遇到问题：

1. **查阅文档**: [doc/README.md](./doc/README.md)
2. **提交 Issue**: https://github.com/fndome/xb/issues
3. **查看示例**: [examples/](./examples/README.md)

---

**感谢您使用 xb（原 sqlxb）！** 🚀

