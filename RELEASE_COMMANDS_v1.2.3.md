# xb v1.2.3 发布命令

## 📦 发布信息

- **版本**: v1.2.3
- **提交**: `<commit-sha>`
- **分支**: `main`
- **测试**: ✅ go test ./... （CTE + UNION 场景全部通过）
- **文档**: ✅ README / CHANGELOG / Release Notes 已更新

---

## 🚀 发布步骤

### 1️⃣ 推送代码

```bash
cd d:\MyDev\server\xb
git push origin main
```

### 2️⃣ 创建标签

```bash
git tag v1.2.3
git push origin v1.2.3
```

### 3️⃣ 在 GitHub 创建 Release

- 标题：`xb v1.2.3`
- 内容：使用 `RELEASE_v1.2.3.md`

---

## 📋 发布检查清单

- [x] 所有测试通过（`go test ./...`）
- [x] README 已更新（新增 CTE/UNION 章节）
- [x] CHANGELOG 更新至 1.2.3
- [x] Release Notes & Test Report 已生成
- [ ] 代码已推送到远程
- [ ] 标签已创建并推送
- [ ] GitHub Release 已发布

---

## 📝 参考提交 (v1.2.2 → v1.2.3)

```
<commit-sha>  feat: add CTE + UNION builders and metadata hook
...
```

> 在最终发布前，请将 `<commit-sha>` 与提交列表替换为实际记录。

---

## ✨ v1.2.3 核心特性

- `With()/WithRecursive()`：一行代码构建 CTE/递归查询。
- `UNION(kind, fn)`：UNION / UNION ALL 链式拼接。
- `Meta(func)`：链式注入 TraceID、TenantID、custom 标签。
- 安全性：FROM Alias 自动归一化，常量冲突清理。

---

## 🔗 相关链接

- **Repository**: https://github.com/fndome/xb
- **Documentation**: ./README.md
- **Changelog**: ./commit_message/CHANGELOG.md
- **Release Notes**: ./RELEASE_v1.2.3.md
- **Test Report**: ./TEST_REPORT_v1.2.3.md

---

**准备发布！** 🚀

