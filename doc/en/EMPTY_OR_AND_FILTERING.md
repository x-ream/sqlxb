# Empty OR/AND Filtering (English)

This note mirrors `xb/doc/EMPTY_OR_AND_FILTERING.md`. It explains how xb prevents `WHERE ()` or `OR ()` clauses by dropping empty blocks.

---

## Behavior

- `Or(func(*CondBuilder))` and `And(func(*CondBuilder))` evaluate their nested builder first.
- If the nested builder yields zero valid conditions, the block is skipped entirely.
- This avoids SQL like `WHERE status = ? AND ()`.

---

## Practical tips

- Use `Bool(condition, func(*CondBuilder))` to wrap optional blocks explicitly.
- For debugging, log `built.Conds` with `%+v` to see which blocks survived.
- If you need to force insert an empty block (very rare), use `X("...")` manually.

---

## Related docs

- `doc/en/ALL_FILTERING_MECHANISMS.md`
- `doc/en/FILTERING.md`

