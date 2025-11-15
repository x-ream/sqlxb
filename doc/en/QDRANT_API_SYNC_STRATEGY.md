# Qdrant API Sync Strategy (English)

English summary of `xb/doc/QDRANT_API_SYNC_STRATEGY.md`. It records how the project keeps xb’s Qdrant implementation aligned with upstream changes.

---

## Process

1. **Monitor releases** – subscribe to Qdrant release notes and API change logs.
2. **Diff schemas** – compare official JSON payloads with xb snapshots whenever a release drops.
3. **Add feature flags** – gate risky additions (e.g., payload projection) behind `QdrantCustom` helpers.
4. **Verify via tests** – expand `xb/qdrant_test.go` and `xb/qdrant_custom_test.go` with new fixtures.

---

## Version mapping

| Qdrant version | xb requirement | Notes |
|----------------|----------------|-------|
| 1.6+ | `JsonOfSelect` unified | Recommend/Discover/Scroll supported |
| 1.4–1.5 | Some APIs require compatibility mode | Use `CustomLegacy()` if needed |
| <1.4 | Not officially tested | Rely on snapshots or custom fork |

---

## Tooling

- `./scripts/qdrant-schema-sync.sh` (internal) collects JSON samples.
- Snapshot tests (`testdata/qdrant/*.json`) ensure no regressions slip in.
- Release checklists require updating this document whenever Qdrant announces breaking changes.

---

## Related docs

- `doc/en/QDRANT_GUIDE.md`
- `doc/en/QDRANT_ADVANCED_API.md`
- `doc/en/VECTOR_DIVERSITY_QDRANT.md`

