# 从 sqlxb 迁移到 xb

## 🔄 项目重命名通知

**版本**: v0.10.5  
**日期**: 2025-10-28

从 `v0.10.5` 开始，项目从 `sqlxb` 重命名为 `xb`。

---

## 📋 为什么重命名？

1. **更简洁的名称** - `xb` 比 `sqlxb` 更短、更易记
2. **品牌统一** - 与 x-ream 组织命名风格保持一致
3. **更好的可发现性** - 短名称在搜索和推荐时更具优势

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
    github.com/x-ream/xb v0.10.5
)
```

---

### 2️⃣ 更新 import 语句

**修改前**:
```go
import (
    "github.com/x-ream/sqlxb"
)
```

**修改后**:
```go
import (
    "github.com/x-ream/xb"
)
```

---

### 3️⃣ 包名保持不变（向后兼容）

✅ **无需修改代码** - 包名仍然是 `sqlxb`：

```go
// ✅ 这些代码无需修改
builder := sqlxb.Of(&User{})
qx := sqlxb.QdrantX{}
built := builder.Build()
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
go list -m all | grep "x-ream"
```

**期望输出**:
```
github.com/x-ream/xb v0.10.5
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

**A**: 可以。旧的 `github.com/x-ream/sqlxb` 仓库会保留到 `v0.10.4`，但不再维护。建议尽快迁移到 `xb`。

---

### Q2: 我需要修改代码中的 `sqlxb` 包名吗？

**A**: **不需要**。包名仍然是 `sqlxb`，只需要修改 `import` 路径即可：

```go
import (
    "github.com/x-ream/xb"  // ✅ 只改这里
)

// ✅ 代码无需修改
builder := sqlxb.Of(&User{})
```

---

### Q3: 我的项目使用了 v0.10.4 之前的版本，怎么办？

**A**: 分两步迁移：

1. **先升级到 v0.10.4**（仍使用 `sqlxb`）
2. **再升级到 v0.10.5**（切换到 `xb`）

```bash
# Step 1
go get github.com/x-ream/sqlxb@v0.10.4
go mod tidy

# 测试确认无误后
# Step 2
go get github.com/x-ream/xb@v0.10.5
# 然后按照上面的步骤修改 import 路径
```

---

### Q4: 我使用了 `replace` 指令怎么办？

**A**: 更新 `go.mod` 中的 `replace` 指令：

**修改前**:
```go
replace github.com/x-ream/sqlxb => /path/to/local/sqlxb
```

**修改后**:
```go
replace github.com/x-ream/xb => /path/to/local/xb
```

---

## 🔗 相关资源

- **GitHub 仓库**: https://github.com/x-ream/xb
- **文档**: https://github.com/x-ream/xb/blob/main/README.md
- **Roadmap**: https://github.com/x-ream/xb/blob/main/doc/ROADMAP_v1.0.0.md
- **Issues**: https://github.com/x-ream/xb/issues

---

## 💬 需要帮助？

如果您在迁移过程中遇到问题：

1. **查阅文档**: [doc/README.md](./doc/README.md)
2. **提交 Issue**: https://github.com/x-ream/xb/issues
3. **查看示例**: [examples/](./examples/README.md)

---

**感谢您使用 xb（原 sqlxb）！** 🚀

