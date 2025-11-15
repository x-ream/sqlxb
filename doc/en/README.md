# xb Documentation (English)

The `doc/en/` directory hosts the canonical English documentation set for xb and its multi-language builders (`xb`, `vxb`, `zxb`). Every document listed here will eventually have a Chinese counterpart under `doc/cn/`.

---

## Current Index

| Category | File | Status | Notes |
|----------|------|--------|-------|
| Overview | `README.md` | ✅ | Entry point + migration status |
| Quickstart | `QUICKSTART.md` | ✅ | SQL + Qdrant basics |
| Vector Guide | `VECTOR_GUIDE.md` | ✅ | Embedding hygiene + hybrid patterns |
| Qdrant Guide | `QDRANT_GUIDE.md` | ✅ | Recommend/Discover/Scroll cookbook |
| Custom Interface | `CUSTOM_INTERFACE.md` | ✅ | Implementing `Custom.Generate()` |
| Filtering & Validation | `FILTERING.md` | ✅ | Auto-skip, guards, validation hooks |
| AI Application Starter | `AI_APPLICATION.md` | ✅ | Consolidated RAG + agent guidance |

> Legacy documents that are still relevant but not yet ported remain under `xb/doc/`. Each section below calls out its temporary source for reference.

---

## Migration Strategy

1. **Authoritative English first** – new content lands in `doc/en/`. Once stable, it is translated to `doc/cn/` using the same filename.
2. **One topic = one file** – avoids fragmentation (`VECTOR_GUIDE.md` now replaces five separate vector docs).
3. **Deprecate legacy paths gradually** – old links continue to work until every README snippet references the new location.
4. **Release alignment** – whenever a feature makes it into release notes, the corresponding doc/en page must mention it.

---

## How to Contribute

1. Pick an English doc that needs updates and edit it in-place (Markdown, ASCII preferred).
2. If a Chinese mirror exists, create/update `doc/cn/<same filename>` in a separate PR or commit.
3. Cross-link from the repository root `README.md` once the document is production ready.

Questions or suggestions? Open a discussion under [GitHub Discussions](https://github.com/fndome/xb/discussions) with the `documentation` tag. PRs that migrate or translate docs are extremely welcome!


