# Organizations as a Subject type

**Status:** idea, 2026-10-25. Not scheduled. Descoped while planning S11-48, which records religion and languages spoken as term Properties on a person instead.

## The problem

Many records name a body a person belonged to, served or passed through: a parish or congregation, an employer, a school, a regiment or ship's company, a lodge (Orange Lodge, Masons, Oddfellows), a union, a benevolent society. Today that body can only be a `remark`, a researcher text Property, or a term (S11-48's religion). None of those gives the body a record of its own:

- No place for its history: when it was founded, what it was part of, what it merged into.
- No way to say two records mean the same body ("St. Patrick's, Ottawa" in a baptism and in a marriage).
- No query from the body's side: everyone baptized at St. Patrick's, everyone in the 48th Highlanders.

## The idea

A new Subject type, `organization`, with its own ref prefixes, beside `person`, `event` and `place`.

- **Properties.** A name (text, multiple, like `toponym`), a period (`start_date` / `end_date`, as on a Place), and for a religious body a `religion` term from S11-48, so "all Catholics" still counts people recorded only through a congregation.
- **Structure like places.** Organizations have the same shape as places: a period, "part of" (parish → diocese → church; company → battalion → regiment) and "succeeded by" (Methodist Church of Canada → United Church of Canada, 1925). An `organization_relationship` bridge copies `place_relationship` and its `part_of` / `succeeded_by` terms.
- **People link with a graded role.** A person–organization bridge carries a `role` term that says how strong the link is: `affiliated` (a census says so), `member`, `baptized`, `communicant`, `employee`, `student`, `soldier`, `clergy`, `officer`. The Source decides the role, so "R.C." on a census never claims membership. Whether the person practised is a Conclusion the researcher argues, not a shape in the data.
- **Events link too.** A baptism happened *at* a congregation; an enlistment was *into* a regiment. Either an event–organization bridge, or a person–organization link with the event behind it.

## Open questions

1. **Building or body.** "Baptized at St. Patrick's" can mean the church building (a Place, through the event's location) or the congregation (an organization). Probably both, with a link from the body to the places it used.
2. **Denominations: terms or organizations.** S11-48 makes a denomination a term. If organizations land, is "Roman Catholic Church" also an organization at the top of a part-of chain, and does the `religion` term point at it, or stay a plain classification? Keeping both means two paths to "all Catholics".
3. **Term hierarchy.** Wesleyan Methodist, Primitive Methodist and Episcopal Methodist are all Methodist. A part-of chain of organizations answers that; so would a broader-than link between terms ([`term-management.md`](term-management.md)). Pick one.
4. **Matching.** Whether a shared organization (same congregation, same employer) counts as evidence in Promote, and how much ([`matching-model.md`](../matching-model.md)).
5. **Gazetteer-style service.** A lookup for well-known bodies (dioceses, regiments) the way [`place-gazetteer-service.md`](place-gazetteer-service.md) proposes for places.

## What it would touch

- **Seed:** `subjectvocab` types, Properties, bindings and terms; `connectrules` for the bridges; new ref prefixes ([`seeded-vocabulary.md`](../seeded-vocabulary.md) §3.1).
- **Swift:** a `subject_organization` mark and asset, Organizations list and detail as a workspace place, composer support (`add-workspace-place`, `add-ui-component`), each behind a design brief.
- **Conclusions:** an Organization canonical entity and its reconciliation, like Place.
