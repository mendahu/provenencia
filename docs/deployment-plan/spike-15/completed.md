# Spike 15 — Completed

One short paragraph per landed PR, newest first. PR map: [`deployment-plan.md`](deployment-plan.md).

**S15-01 — The credibility argument is a text area** ([#758](https://github.com/mendahu/provenencia/pull/758)). The Source page's credibility argument is a multi-line text area that wraps and grows to 12 lines; Return adds a line and Save saves it with its line breaks. `PVTextArea` takes its own accessibility label. Two text areas in one column no longer crash: each finds its own text view and a text view keeps one Tab router.
