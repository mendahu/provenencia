# Persons list sorts "Mr." and "Mrs." under M

- **Date:** 2026-10-15
- **Type:** Bug (sorting)
- **Where:** Persons list sort
- **Annoyance:** The alphabetical Persons list sorts by the full name including honorific prefixes, so people recorded as "Mr. …" or "Mrs. …" land under **M** instead of beside their actual surname or name.
- **Wanted:** Sorting ignores name prefixes / honorifics (Mr., Mrs., Ms., Miss, Dr., Rev., …) and orders by the real name. Other prefixes in the title registry most likely need the same treatment.

## Code notes (observed, not fixed)

- `core/database/conclusionheaders/order.go` — `personSortKey` is the Persons list order: a name with parts sorts by `nameformats.Format(name, profile.Sorted)`; a name with **no parts** falls back to `strings.TrimSpace(h.Entity.Label)`, the raw working label. For a person whose name was never split into parts, "Mr. John Smith" is the sort key verbatim. `sortByTitle` then collates case- and diacritic-insensitively but does not strip anything.
- `core/database/nameformats/registry.go` — the seeded Western profile's **sorted** style is `surname_prefix`, `surname, `, `prefix`, `given`, `nick`, `suffix`. Surname leads, so a fully typed name is fine, but when the surname part is empty the key starts with the prefix ("Mrs. Smith" → sorts under M). The prefix sits before `given` in the sorted order as well, so within one surname rows order by honorific rather than given name.
- `macos/App/Features/Names/NameValueDraft.swift` — `classify(_:allowSurname:)` only peels a leading title while more than one word remains, and a single leftover word becomes `given`, not `surname`. "Mrs. Smith" therefore yields `prefix` + `given` with no surname, which feeds the empty-surname case above. "Mr. John Smith" splits correctly.
- `macos/App/Features/Names/NameTitleRegistry.swift` — `prefixes` is the honorific list (`mr`, `mrs`, `ms`, `miss`, `mx`, `dr`, `rev`, `hon`, `sir`, `dame`, `prof`, `fr`, `capt`, `col`, `gen`, `sr`); Swift-side only today. A Go-side fix that strips titles from the label fallback would need the same list, or the sort key could simply omit `prefix` parts (`namevalues.PartTypePrefix`) and skip leading registry titles in the label fallback.
- `macos/App/Features/Conclusions/ConclusionListPage.swift` — the list renders rows in the order Go returns ("Go owns list order (conclusionheaders.sortByTitle)"), so the fix is in `core`, not the view. `PersonHeaderDisplay.title(…, style: .sorted)` is the displayed title and already shares the same `Sorted` order.
- `core/database/conclusionheaders/conclusionheaders_test.go` / `core/database/nameformats/nameformats_test.go` — existing tests to extend with a "Mrs. Smith" and a parts-less "Mr. John Smith" case.
