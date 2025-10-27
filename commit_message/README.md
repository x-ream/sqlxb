# 提交信息管理

本目录用于管理 Git 提交信息，包括历史提交记录和模板。

---

## 📁 目录结构

```
commit_message/
├── README.md                          # 本说明文件
├── COMMIT_MESSAGE_v0.9.1.txt         # v0.9.1 提交信息（中文）✅ 已提交
├── COMMIT_MESSAGE_v0.9.1_EN.txt      # v0.9.1 提交信息（英文）✅ 已提交
└── TEMPLATE.txt                       # 提交信息模板（未来使用）
```

---

## 📝 文件说明

### 已使用的提交信息

格式：`COMMIT_MESSAGE_vX.Y.Z.txt` 和 `COMMIT_MESSAGE_vX.Y.Z_EN.txt`

- **用途**: 记录历史版本的提交信息
- **版本控制**: ✅ 提交到 Git（作为文档）
- **作用**:
  - 规范化提交信息格式
  - 方便未来查阅版本变更
  - 作为 Release Notes 参考

### 提交信息模板

格式：`TEMPLATE.txt`

- **用途**: 作为未来提交的模板
- **版本控制**: ✅ 提交到 Git（共享规范）

### 临时文件

格式：`COMMIT_MESSAGE_DRAFT*.txt`, `temp*.txt`

- **用途**: 本地草稿，测试用
- **版本控制**: ❌ 被 `.gitignore` 忽略

---

## 🚀 使用方法

### 1. 使用已有的提交信息

```bash
# 提交代码
git add .
git commit -F commit_message/COMMIT_MESSAGE_v0.9.1.txt

# 或使用英文版
git commit -F commit_message/COMMIT_MESSAGE_v0.9.1_EN.txt
```

### 2. 创建新版本的提交信息

```bash
# 复制模板
cp commit_message/TEMPLATE.txt commit_message/COMMIT_MESSAGE_v0.9.2.txt

# 编辑提交信息
# ... 修改 COMMIT_MESSAGE_v0.9.2.txt

# 使用
git commit -F commit_message/COMMIT_MESSAGE_v0.9.2.txt
```

### 3. 本地草稿（不提交到 Git）

```bash
# 创建草稿
cp commit_message/TEMPLATE.txt commit_message/COMMIT_MESSAGE_DRAFT.txt

# 编辑草稿
# ... 修改

# 使用（这个文件不会被 Git 追踪）
git commit -F commit_message/COMMIT_MESSAGE_DRAFT.txt
```

---

## 📋 提交信息规范

### 格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type（类型）

- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档变更
- `style`: 代码格式（不影响功能）
- `refactor`: 重构（不是新功能，也不是修复 bug）
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

### Scope（范围）

- `core`: 核心功能
- `vector`: 向量数据库
- `qdrant`: Qdrant 集成
- `test`: 测试
- `doc`: 文档

### 示例

```
fix(core): 修复 float64/int 零值过滤和 OR/AND 子查询处理

**问题 (Issues Fixed)**:

1. **float64/int 零值无法过滤**
   - `interface{} == 0` 对 float64 类型无效
   - 需要类型断言后再比较：`v.(float64) == 0.0`
   
2. **向量查询中 OR/AND 子查询渲染错误**
   - `SqlOfVectorSearch` 使用简化的 `buildConditionSql`
   - 该方法忽略 `subs`，导致 OR_SUB 输出为 `OR OR ?`
   - 修复：改用正确的 `toCondSql()`

**版本 (Version)**: v0.9.1
```

---

## 🔄 CI/CD 集成

如果需要在 CI/CD 中使用这些提交信息：

```yaml
# .github/workflows/release.yml
- name: Create Release
  run: |
    # 使用提交信息创建 Release Notes
    gh release create ${{ github.ref }} \
      --title "Release ${{ github.ref }}" \
      --notes-file commit_message/COMMIT_MESSAGE_${VERSION}.txt
```

---

## 📚 历史版本

- **v0.9.1**: 修复 float64/int 零值过滤和 OR/AND 子查询处理
  - 文件: `COMMIT_MESSAGE_v0.9.1.txt`
  - 日期: 2025-01

---

## ✅ 最佳实践

1. **版本提交信息**: ✅ 提交到 Git
   - 作为文档保存
   - 方便未来查阅
   - 规范团队提交格式

2. **临时草稿**: ❌ 不提交
   - 本地测试使用
   - 避免污染仓库

3. **模板文件**: ✅ 提交到 Git
   - 统一团队规范
   - 新贡献者参考

---

*本目录由 AI (Claude) 协助创建和维护*

