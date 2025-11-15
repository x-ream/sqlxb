# AI Application FAQ (English)

English adaptation of `xb/doc/ai_application/FAQ.md`.

---

## Common questions

**Q: Can I mix SQL and vector queries in one builder?**  
A: Yes—vector features simply add extra metadata to `Built` and do not break SQL generation.

**Q: Do I need Qdrant to use xb?**  
A: No, xb works with plain SQL. Qdrant is the first official vector adapter.

**Q: How do I debug skipped filters?**  
A: Inspect `built.Conds` or `built.SqlOfSelect()` to see which clauses survived auto-filtering.

---

## Support

- Open GitHub issues with reproduction steps.
- Join discussions to propose new AI integrations.

