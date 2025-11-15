# Qdrant Custom Design Analysis (English)

Translation of `xb/doc/QDRANT_CUSTOM_DESIGN_ANALYSIS.md`. It dissects how the Qdrant adapter is structured and why certain trade-offs were made.

---

## Architectural layers

1. **Builder** – collects DSL state (filters, vectors, sorts, payload selectors).
2. **Custom config** – exposes helpful knobs (`Recommend`, `Discover`, `Scroll`, diversity, payload selector).
3. **Generator** – maps builder state to JSON payloads.
4. **Transport** – left to the caller (HTTP client, SDK, etc.).

---

## Key decisions

- **Single entry point** – `JsonOfSelect()` inspects the config to choose Search/Recommend/Discover/Scroll.
- **Composable builders** – advanced APIs re-use the same `CondBuilder` DSL for filters.
- **Preset functions** – `NewQdrantCustom()` exposes default settings; advanced presets (high recall, high diversity) can be layered on.

---

## Extensibility hooks

- `WithPayloadSelector`
- `WithHashDiversity`
- `WithMinDistance`
- `WithNamespace`, `WithTenant`

Each helper updates the adapter state and can be combined with advanced APIs.

---

## Testing strategy

- JSON snapshots for every branch (search + 3 advanced APIs).
- Regression tests ensure metadata propagation and auto-filtering behave consistently.
- Add targeted tests whenever Qdrant ships new optional fields.

---

## Related docs

- `doc/en/QDRANT_GUIDE.md`
- `doc/en/QDRANT_ADVANCED_API.md`
- `doc/en/TO_JSON_DESIGN_CLARIFICATION.md`

