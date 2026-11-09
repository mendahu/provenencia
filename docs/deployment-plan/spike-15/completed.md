# Spike 15 — Completed

One short paragraph per landed PR, newest first. PR map: [`deployment-plan.md`](deployment-plan.md).

**S15-02 — Sources carry a date and a not-public flag** ([#762](https://github.com/mendahu/provenencia/pull/762)). A Source has an optional date of its own (a shared DateValue, set, edited in place and cleared by `SetSourceDate`, released when the Source is deleted) and a not-public flag that defaults to public (`SetSourcePublic`, stored and read back only). Every Source on the wire carries the date, a `date_sort_key` and `is_public`. The date order moved to `valuecompare.DateKeyOf`, which the Timeline, birth order, the Sources key and the reconciler's stored date `sort_key` now share (`CacheVersion` 24). `datevalues.PutTx`, `namevalues.PutTx` and `deleteimpact.ReleaseReplaced` are the one way an owner stores and releases a value row. No Swift view changes; product 0.5.0.

**S15-01 — The credibility argument is a text area** ([#758](https://github.com/mendahu/provenencia/pull/758)). The Source page's credibility argument is a multi-line text area that wraps and grows to 12 lines; Return adds a line and Save saves it with its line breaks. `PVTextArea` takes its own accessibility label. Two text areas in one column no longer crash: each finds its own text view and a text view keeps one Tab router.
