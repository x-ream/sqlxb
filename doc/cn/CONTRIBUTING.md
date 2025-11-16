# 贡献指南

感谢您考虑为 `xb` 做出贡献！🎉

我们热烈欢迎各种形式的贡献。本指南将帮助您入门。

---

## 📋 目录

* [报告安全问题](#报告安全问题)
* [报告一般问题](#报告一般问题)
* [提出功能建议](#提出功能建议)
* [代码贡献](#代码贡献)
* [测试贡献](#测试贡献)
* [文档](#文档)
* [社区参与](#社区参与)

---

## 🔒 报告安全问题

安全问题始终受到严肃对待。我们不鼓励任何人在公开场合传播安全问题。

如果您在 xb 中发现安全漏洞：
- ❌ **不要** 在公开场合讨论
- ❌ **不要** 打开公开 issue
- ✅ **请** 发送私密邮件至 [8966188@gmail.com](mailto:8966188@gmail.com)

---

## 🐛 报告一般问题

我们将 xb 的每个用户视为有价值的贡献者。使用 xb 后，如果您有反馈，请随时通过 [新建 ISSUE](https://github.com/fndome/xb/issues/new/choose) 打开问题。

### Issue 指南

我们欣赏**写得好**、**详细**、**明确**的问题报告。在打开新问题之前：
1. 搜索现有问题以避免重复
2. 向现有问题添加详细信息，而不是创建新问题
3. 遵循问题模板
4. 删除敏感数据（密码、密钥、私人数据等）

### Issue 模板

```markdown
**问题描述**：
问题的清晰描述

**重现步骤**：
1. ...
2. ...
3. ...

**预期行为**：
应该发生什么

**实际行为**：
实际发生了什么

**环境**：
- Go 版本：
- xb 版本：
- 数据库：PostgreSQL / MySQL / Qdrant
- 操作系统：
```

### 问题类型

* 🐛 Bug 报告
* ✨ 功能请求
* ⚡ 性能问题
* 💡 功能提案
* 📐 功能设计
* 🆘 需要帮助
* 📖 文档不完整
* 🧪 测试改进
* ❓ 关于项目的问题

---

## 💡 提出功能建议

对于功能请求，请在 [Issues](https://github.com/fndome/xb/issues) 上使用 `[Feature Request]` 标签：

### 功能请求模板

```markdown
**业务场景**：
为什么需要此功能？

**预期 API**：
```go
// 您希望如何使用它
xb.NewFeature()...
```

**替代方案**：
今天存在哪些替代方案？

**参考资料**：
相关文档或项目的链接
```

### 决策过程

我们根据以下标准评估功能：
1. ✅ 真实用户需求
2. ✅ 与 xb 愿景一致
3. ✅ API 向后兼容性
4. ✅ 社区维护者可用性

有关技术演进方法，请参阅 [VISION.md](../VISION.md)。

---

## 💻 代码贡献

xb 的每一项改进都受到鼓励！在 GitHub 上，贡献通过 Pull Requests (PRs) 进行。

### 贡献内容

* 修复拼写错误
* 修复 bug
* 删除冗余代码
* 添加缺失的测试用例
* 增强功能
* 添加澄清注释
* 重构丑陋的代码
* 改进文档
* **还有更多！**

> **我们期待您的任何 PR。**

### 工作区设置

```bash
# Fork 并克隆
git clone https://github.com/YOUR_USERNAME/xb.git
cd xb

# 创建功能分支
git checkout -b feature/your-feature-name
```

### 开发工作流

1. **编写代码**
   - 遵循现有代码风格
   - 添加必要的注释
   - 确保类型安全

2. **添加测试**
   ```bash
   # 运行测试
   go test ./...
   
   # 检查覆盖率
   go test -cover
   ```

3. **更新文档**
   - 为新功能更新 `README.md`
   - 向 `examples/` 添加示例
   - 更新相关的 `.md` 文件

4. **提交更改**
   ```bash
   git add .
   git commit -m "feat: add XXX feature"
   ```

### 提交消息约定

```
type: 简短描述

详细描述（可选）

- 更改 1
- 更改 2
```

**类型**：
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档
- `test`: 测试
- `refactor`: 代码重构
- `perf`: 性能优化
- `chore`: 构建/工具配置

**示例**：
```
feat: add Milvus vector database support

- Implement MilvusX Builder
- Add unit tests
- Update docs and examples
```

### 代码风格

```go
// ✅ 好例子
func (b *Builder) Eq(key string, value interface{}) *Builder {
    if value == nil {
        return b  // 自动过滤 nil
    }
    b.conds = append(b.conds, Condition{
        Key:   key,
        Op:    "=",
        Value: value,
    })
    return b
}

// ❌ 避免
func (b *Builder) eq(k string, v interface{}) *Builder {  // 应该导出
    b.conds = append(b.conds, Condition{k, "=", v})  // 使用字段名
    return b
}
```

### 提交 Pull Requests

1. **推送到 Fork**
   ```bash
   git push origin feature/your-feature-name
   ```

2. **创建 PR**
   - 在 GitHub 上创建 Pull Request
   - 填写 PR 模板
   - 等待审查

3. **处理反馈**
   - 回复审查评论
   - 进行请求的更改
   - 推送更新（PR 自动更新）

---

## 🧪 测试贡献

欢迎任何测试用例！目前，xb 功能测试用例是高度优先的。

### 测试要求

新功能必须包括：

1. **单元测试**
   - 覆盖核心逻辑
   - 包括边缘情况
   - 测试自动过滤

2. **示例**（对于重要功能）
   - 向 `examples/` 添加完整示例
   - 包含 README.md
   - 必须可运行

3. **文档**
   - API 文档
   - 使用示例
   - 重要说明

### 测试风格

```go
// ✅ 好测试
func TestEqAutoFiltering(t *testing.T) {
    // Arrange
    builder := Of(&User{})
    
    // Act
    builder.Eq("status", nil).  // 应该被忽略
            Eq("name", "Alice")  // 应该工作
    
    // Assert
    sql, args, _ := builder.Build().SqlOfSelect()
    if !strings.Contains(sql, "name = ?") {
        t.Errorf("Expected name condition")
    }
    if len(args) != 1 {
        t.Errorf("Expected 1 arg, got %d", len(args))
    }
}
```

### 外部测试项目

欢迎提交测试 PR 到：
- https://github.com/sim-wangyan/xb-test-on-sqlx
- 或您自己的项目：`https://github.com/YOUR_USERNAME/xb-test-YOUR_PROJECT`

---

## 📖 文档

文档改进非常受重视！

### 改进内容

- 修复拼写错误和错误
- 添加缺失的文档
- 创建更多示例
- 提高清晰度
- 翻译成其他语言

### 文档结构

```
xb/
├── README.md              # 主要文档
├── VISION.md             # 项目愿景
├── MIGRATION.md          # 迁移指南
├── doc/
│   ├── CONTRIBUTING.md   # 本文件
│   ├── ai_application/  # AI 应用指南
│   └── ...
└── examples/             # 示例应用
    ├── pgvector-app/
    ├── qdrant-app/
    ├── rag-app/
    └── pageindex-app/
```

---

## 🤝 社区参与

GitHub 是我们的主要协作平台。除了 PR，您可以通过多种方式提供帮助：

### 贡献方式

- 💬 回复其他人的问题
- 🆘 帮助解决用户问题
- 👀 审查 PR 设计
- 🔍 审查 PR 中的代码
- 💭 讨论以澄清想法
- 📢 在 GitHub 之外推广 xb
- ✍️ 撰写关于 xb 的博客
- 🎓 在 [Discussions](https://github.com/fndome/xb/discussions) 中分享最佳实践

### 沟通渠道

- 💬 **技术讨论**：[GitHub Discussions](https://github.com/fndome/xb/discussions)
- 🐛 **Bug 报告**：[GitHub Issues](https://github.com/fndome/xb/issues)
- 📖 **文档**：[README.md](../README.md)

---

## 🎯 优先贡献

### 高价值领域

1. **Bug 修复** 🐛
   - 解决报告的问题
   - 修复边缘情况
   - 改进错误消息

2. **文档** 📖
   - 填补文档空白
   - 添加更多示例
   - 改进解释

3. **性能** ⚡
   - 减少内存分配
   - 优化 SQL 生成
   - 提高查询性能

4. **数据库支持** 🗄️
   - Milvus / Weaviate / Pinecone
   - 保持 API 一致性
   - 提供测试和文档

5. **AI 用例** 🤖
   - RAG 最佳实践
   - Agent 工具集成
   - Prompt 工程辅助方法

---

## 📏 行为准则

- ✅ 尊重所有贡献者
- ✅ 提供建设性反馈
- ✅ 欢迎新人的问题
- ❌ 禁止人身攻击或骚扰

---

## 🏗️ 项目结构

```
xb/
├── builder_x.go          # 核心 Builder
├── cond_builder_x.go     # 条件构建器
├── to_sql.go            # SQL 生成
├── qdrant_x.go          # Qdrant 客户端
├── to_qdrant_json.go    # Qdrant JSON 生成
├── vector_types.go      # 向量类型
├── doc/                 # 文档
│   ├── ai_application/  # AI 应用文档
│   └── ...
├── examples/            # 示例应用
│   ├── pgvector-app/
│   ├── qdrant-app/
│   ├── rag-app/
│   └── pageindex-app/
└── *_test.go           # 测试文件
```

---

## 🌟 最后的话

> **在技术快速迭代的时代，灵活性比完美规划更重要。**

有关拥抱变化和社区驱动开发的方法，请参阅 [VISION.md](../VISION.md)。

---

**总之：任何帮助都是贡献。** 🚀

感谢您让 xb 变得更好！

