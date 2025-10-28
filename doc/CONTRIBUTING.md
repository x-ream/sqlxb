# Contributing to xb

Thank you for considering contributing to `xb`! 🎉

We warmly welcome all forms of contributions. This guide will help you get started.

---

## 📋 Table of Contents

* [Reporting Security Issues](#reporting-security-issues)
* [Reporting General Issues](#reporting-general-issues)
* [Proposing Features](#proposing-features)
* [Code Contribution](#code-contribution)
* [Test Contribution](#test-contribution)
* [Documentation](#documentation)
* [Community Engagement](#community-engagement)

---

## 🔒 Reporting Security Issues

Security issues are always treated seriously. We discourage anyone from spreading security issues publicly. 

If you find a security vulnerability in xb:
- ❌ **DO NOT** discuss it in public
- ❌ **DO NOT** open a public issue
- ✅ **DO** send a private email to [8966188@gmail.com](mailto:8966188@gmail.com)

---

## 🐛 Reporting General Issues

We regard every user of xb as a valued contributor. After using xb, if you have feedback, please feel free to open an issue via [NEW ISSUE](https://github.com/fndome/xb/issues/new/choose).

### Issue Guidelines

We appreciate **WELL-WRITTEN**, **DETAILED**, **EXPLICIT** issue reports. Before opening a new issue:
1. Search existing issues to avoid duplicates
2. Add details to existing issues rather than creating new ones
3. Follow the issue template
4. Remove sensitive data (passwords, keys, private data, etc.)

### Issue Template

```markdown
**Describe the Issue**:
Clear description of the problem

**Steps to Reproduce**:
1. ...
2. ...
3. ...

**Expected Behavior**:
What should happen

**Actual Behavior**:
What actually happened

**Environment**:
- Go version:
- xb version:
- Database: PostgreSQL / MySQL / Qdrant
- OS:
```

### Types of Issues

* 🐛 Bug report
* ✨ Feature request
* ⚡ Performance issues
* 💡 Feature proposal
* 📐 Feature design
* 🆘 Help wanted
* 📖 Documentation incomplete
* 🧪 Test improvement
* ❓ Questions about the project

---

## 💡 Proposing Features

For feature requests, use the `[Feature Request]` label on [Issues](https://github.com/fndome/xb/issues):

### Feature Request Template

```markdown
**Business Scenario**:
Why is this feature needed?

**Expected API**:
```go
// How you'd like to use it
xb.NewFeature()...
```

**Alternatives**:
What alternatives exist today?

**References**:
Links to relevant docs or projects
```

### Decision Process

We evaluate features based on:
1. ✅ Real user needs
2. ✅ Alignment with xb's vision
3. ✅ API backward compatibility
4. ✅ Community maintainer availability

See [VISION.md](../VISION.md) for our approach to technical evolution.

---

## 💻 Code Contribution

Every improvement to xb is encouraged! On GitHub, contributions come via Pull Requests (PRs).

### What to Contribute

* Fix typos
* Fix bugs
* Remove redundant code
* Add missing test cases
* Enhance features
* Add clarifying comments
* Refactor ugly code
* Improve documentation
* **And more!**

> **WE ARE LOOKING FORWARD TO ANY PR FROM YOU.**

### Workspace Setup

```bash
# Fork and clone
git clone https://github.com/YOUR_USERNAME/xb.git
cd xb

# Create feature branch
git checkout -b feature/your-feature-name
```

### Development Workflow

1. **Write Code**
   - Follow existing code style
   - Add necessary comments
   - Ensure type safety

2. **Add Tests**
   ```bash
   # Run tests
   go test ./...
   
   # Check coverage
   go test -cover
   ```

3. **Update Docs**
   - Update `README.md` for new features
   - Add examples to `examples/`
   - Update relevant `.md` files

4. **Commit Changes**
   ```bash
   git add .
   git commit -m "feat: add XXX feature"
   ```

### Commit Message Convention

```
type: short description

Detailed description (optional)

- Change 1
- Change 2
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `test`: Testing
- `refactor`: Code refactoring
- `perf`: Performance optimization
- `chore`: Build/tools config

**Example**:
```
feat: add Milvus vector database support

- Implement MilvusX Builder
- Add unit tests
- Update docs and examples
```

### Code Style

```go
// ✅ Good example
func (b *Builder) Eq(key string, value interface{}) *Builder {
    if value == nil {
        return b  // Auto-filter nil
    }
    b.conds = append(b.conds, Condition{
        Key:   key,
        Op:    "=",
        Value: value,
    })
    return b
}

// ❌ Avoid
func (b *Builder) eq(k string, v interface{}) *Builder {  // Should be exported
    b.conds = append(b.conds, Condition{k, "=", v})  // Use field names
    return b
}
```

### Submitting Pull Requests

1. **Push to Fork**
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create PR**
   - Create Pull Request on GitHub
   - Fill out the PR template
   - Wait for review

3. **Address Feedback**
   - Respond to review comments
   - Make requested changes
   - Push updates (PR auto-updates)

---

## 🧪 Test Contribution

Any test case is welcomed! Currently, xb function test cases are high priority.

### Test Requirements

New features must include:

1. **Unit Tests**
   - Cover core logic
   - Include edge cases
   - Test auto-filtering

2. **Examples** (for important features)
   - Add complete example to `examples/`
   - Include README.md
   - Must be runnable

3. **Documentation**
   - API docs
   - Usage examples
   - Important notes

### Test Style

```go
// ✅ Good test
func TestEqAutoFiltering(t *testing.T) {
    // Arrange
    builder := Of(&User{})
    
    // Act
    builder.Eq("status", nil).  // Should be ignored
            Eq("name", "Alice")  // Should work
    
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

### External Test Projects

Welcome to submit test PRs to:
- https://github.com/sim-wangyan/xb-test-on-sqlx
- Or your own project: `https://github.com/YOUR_USERNAME/xb-test-YOUR_PROJECT`

---

## 📖 Documentation

Documentation improvements are highly valued!

### What to Improve

- Fix typos and errors
- Add missing documentation
- Create more examples
- Improve clarity
- Translate to other languages

### Documentation Structure

```
xb/
├── README.md              # Main docs
├── VISION.md             # Project vision
├── MIGRATION.md          # Migration guide
├── doc/
│   ├── CONTRIBUTING.md   # This file
│   ├── ai_application/   # AI app guides
│   └── ...
└── examples/             # Example apps
    ├── pgvector-app/
    ├── qdrant-app/
    ├── rag-app/
    └── pageindex-app/
```

---

## 🤝 Community Engagement

GitHub is our primary collaboration platform. Besides PRs, you can help in many ways:

### Ways to Contribute

- 💬 Reply to others' issues
- 🆘 Help solve user problems
- 👀 Review PR designs
- 🔍 Review code in PRs
- 💭 Discuss to clarify ideas
- 📢 Advocate xb beyond GitHub
- ✍️ Write blogs about xb
- 🎓 Share best practices in [Discussions](https://github.com/fndome/xb/discussions)

### Communication Channels

- 💬 **Technical Discussion**: [GitHub Discussions](https://github.com/fndome/xb/discussions)
- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/fndome/xb/issues)
- 📖 **Documentation**: [README.md](../README.md)

---

## 🎯 Priority Contributions

### High-Value Areas

1. **Bug Fixes** 🐛
   - Resolve reported issues
   - Fix edge cases
   - Improve error messages

2. **Documentation** 📖
   - Fill documentation gaps
   - Add more examples
   - Improve explanations

3. **Performance** ⚡
   - Reduce memory allocations
   - Optimize SQL generation
   - Improve query performance

4. **Database Support** 🗄️
   - Milvus / Weaviate / Pinecone
   - Maintain API consistency
   - Provide tests and docs

5. **AI Use Cases** 🤖
   - RAG best practices
   - Agent tool integration
   - Prompt engineering helpers

---

## 📏 Code of Conduct

- ✅ Respect all contributors
- ✅ Provide constructive feedback
- ✅ Welcome newcomers' questions
- ❌ No personal attacks or harassment

---

## 🏗️ Project Structure

```
xb/
├── builder_x.go          # Core Builder
├── cond_builder_x.go     # Condition builder
├── to_sql.go            # SQL generation
├── qdrant_x.go          # Qdrant client
├── to_qdrant_json.go    # Qdrant JSON generation
├── vector_types.go      # Vector types
├── doc/                 # Documentation
│   ├── ai_application/  # AI app docs
│   └── ...
├── examples/            # Example apps
│   ├── pgvector-app/
│   ├── qdrant-app/
│   ├── rag-app/
│   └── pageindex-app/
└── *_test.go           # Test files
```

---

## 🌟 Final Words

> **In the era of rapid technological iteration, flexibility matters more than perfect planning.**

See [VISION.md](../VISION.md) for our approach to embracing change and community-driven development.

---

**In a word: ANY HELP IS CONTRIBUTION.** 🚀

Thank you for making xb better!
