# xb v1.3.0 发布命令

## 📦 发布信息

- **版本**: v1.3.0
- **提交**: `<commit-sha>`
- **分支**: `main`
- **测试**: ✅ go test ./... （JsonOfSelect + Qdrant 高级 API 全部通过）
- **文档**: ✅ README / MIGRATION / Release Notes / Test Report 已更新

---

## 🚀 发布步骤

### 1️⃣ 推送代码

```bash
cd d:\MyDev\server\xb
git push origin main
```

### 2️⃣ 创建标签

```bash
git tag v1.3.0
git push origin v1.3.0
```

### 3️⃣ GitHub Release

- 标题：`xb v1.3.0`
- 内容：复制 `RELEASE_v1.3.0.md`
- 附件：无（Go module 自动拉取）

---

## 📋 发布检查清单

- [x] `go test ./...`
- [x] README / MIGRATION 更新
- [x] `RELEASE_v1.3.0.md`
- [x] `TEST_REPORT_v1.3.0.md`
- [ ] 代码已推送
- [ ] `v1.3.0` 标签已创建并推送
- [ ] GitHub Release 已发布

---

## 📝 参考命令

```bash
git status
git log -5 --oneline
```

确认无残留修改后再执行发布步骤。

---

## ✨ v1.3.0 核心特性

- `JsonOfSelect()` 统一所有 Qdrant JSON 生成。
- `QdrantCustom.Recommend/Discover/ScrollID` 自动注入高级参数。
- 全量文档/示例切换至新入口，附迁移指南。
- 新增 Recommend/Discover/Scroll 回归测试。

---

## 🔗 相关链接

- Repository: https://github.com/fndome/xb
- Release Notes: ./RELEASE_v1.3.0.md
- Test Report: ./TEST_REPORT_v1.3.0.md
- Migration Guide: ./MIGRATION.md

---

**一切就绪，准备发布 v1.3.0！** 🚀


