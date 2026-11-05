# Some listed lines read the raw term label

- **Date:** 2026-11-05
- **Type:** Friction
- **Where:** Citation composer / saved lines and end titles
- **Annoyance:** The `relationship` and `participation` cases of `listedLine`, and the relationship case of `endTitles`, read the raw `term.label` rather than going through `PropertyTermDisplay`. This is older than the place relationship change.
- **Wanted:** Route them through `PropertyTermDisplay` like the place cases.
