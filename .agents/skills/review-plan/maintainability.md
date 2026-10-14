# Maintainability lens

The question: **this plan will probably work — is there a shape that fits the
codebase better and costs the next change less?** Not more correct. Fewer
copies, fewer special cases, fewer places to look, and the next named add
already has a slot.

Every patch an agent ships is read by every agent after it. A codebase that
repeats a few regular patterns is cheap to read and cheap to change. A web of
one-off fixes is expensive on both, forever. This lens exists because an
implementing model drives straight at the instruction and does not look
around. Look around.

Whole-app criteria this lens draws on: [`review-app-health/criteria.md`](../review-app-health/criteria.md)
§1 cleanliness, §4 structure and naming, §5 separation, §6 idiom, §12 UI
organization, §13 state. UI cousins: the method in [`evaluate-ui-component`](../evaluate-ui-component/SKILL.md) §3–4.

## How to work

1. Read the plan. Then read **Current state**, **Codebase fit**, **Decisions**,
   **Design**, and **Steps** a second time. Those are where the shape is decided.
2. Read the code the plan touches, and the code **around** it: the package, the
   feature folder, its siblings, the registry it should feed.
3. Read every skill in **Skills to load**, including its **Do not** list. A
   plan that crosses one is a finding.
4. For each thing the plan adds, run the questions below. **Search before you
   claim.** Every finding names the code you found, or the map row or idea file
   that names the future add.
5. Then step back from the rows: is there a different overall shape? (Q12.)

## Questions

### Q1. Does it already exist?

For every row in **Codebase fit**, and anything else **Design** or **Steps**
adds that is not in that table: search for an existing type, function, file,
table, component, FFI method, L10n namespace, or test helper doing the same
job. Search by behavior, not just by name: the field it reads, the SQL it runs,
the interaction it owns.

- Exists and fits → reuse it.
- Does 80% → extend it (a parameter, a case, a slot) when the extension stays
  honest for the existing callers.
- A missing row in **Codebase fit** is itself a finding: the writer did not search.

### Q2. Does a registry or declared model already own this?

Provenencia keeps cross-cutting facts in one declared place, and CI holds the
code to it. A plan that special-cases where an entry belongs is a patch.

| Concern | Owner |
| --- | --- |
| Tables, FKs, kinds | `core/database/catalogmodel` ([`add-catalog-model`](../add-catalog-model/SKILL.md)) |
| Delete, inbound probes, releases | `core/database/deleteimpact` ([`add-catalog-delete`](../add-catalog-delete/SKILL.md)) |
| Audit effects and scopes | `core/database/audit` ([`add-audit-scope`](../add-audit-scope/SKILL.md)) |
| Seeded terms, types, properties | `subjectvocab`, `sourcevocab` registries + `Install` ([`add-seeded-vocabulary`](../add-seeded-vocabulary/SKILL.md)) |
| Omnibar hit kinds | `core/search` registry ([`add-searchable-kind`](../add-searchable-kind/SKILL.md)) |
| Workspace places, list/detail caches | `PlaceRegistry`, `CatalogQueryKey`, `CatalogQueryRegistry` ([`add-workspace-place`](../add-workspace-place/SKILL.md)) |
| Navigation history | `WorkspaceLocation`, `WorkspaceNavigation` ([`add-workspace-location`](../add-workspace-location/SKILL.md)) |
| Go wire vocabulary in Swift | `Platform/ConclusionVocabulary.swift` (`.cursor/rules/macos-client.mdc`) |
| Copy | `L10n` + `Localizable.xcstrings` ([`add-localized-string`](../add-localized-string/SKILL.md)) |
| UI chrome | `DesignSystem` layers ([`add-ui-component`](../add-ui-component/SKILL.md)) |
| Error codes | `apperr` + `L10n.Errors` |

Also the reverse: if this is the **third** hand-rolled list of the same kind of
thing (three `switch kind` blocks that must stay in step), the better shape is
a registry. Propose it, usually as a preparatory PR.

### Q3. Is the change at the layer that owns the concern?

- Domain rules live in Go; the Mac client stays thin. Logic in a SwiftUI view,
  or in an FFI handler, that belongs in `core/` is a finding.
- A fix in one call site for a defect in a shared component should be in the
  component, so every caller gains it. Precedent: S10-17 fixed letter cycling
  in `PVSelect` itself, not in the name editor.
- Red flags: `if kind == …` in shared code, a Bool parameter added to a shared
  function for one caller, a feature reaching into another feature's internals,
  a second owner for the same state (criteria §13).

### Q4. One concept, one name, one place

- Names match the domain docs (Source, Artifact, Citation, Observation, Subject,
  Identity Claim, Property, term…). A synonym for an existing concept is a
  finding; so is one word for two concepts.
- New files sit beside their kin and follow the neighbors' naming
  (`core/database/<table>/`, `Features/<Feature>/`, `DesignSystem/<Layer>/<Name>/`).
  No `util`, `helpers`, `common`, `shared`, `misc`, `Extensions.swift` grab-bags.
- A future reader looking for this behavior should find it in **one obvious
  place**, named for what it does.

### Q5. Copies and near-copies

- A step that copies an existing function, view, query, or test and tweaks it:
  name both. Second copy → the plan must say why they will diverge. Third →
  extract, in this PR if small, else a preparatory PR.
- Copy-paste drift already in the area that this PR would make worse is in scope
  for a finding; fixing it may not be (Q13).

### Q6. Not too much, either

Speculative generality is debt too. Findings for:

- A protocol with one conformer, a config option nobody sets, a generic engine
  for one call site, an extension point no map row or idea names.
- A new `DesignSystem` primitive or recipe with one caller (the layer rule:
  snowflake until a second real call site).
- Indirection that adds no policy: a wrapper that only forwards.

Generalize for a **second real caller** or a **future add named in the map or
`docs/ideas/`**. Not for "might be useful".

### Q7. Leave it smaller

What does this change make dead? Code paths, strings, tests, `FakeStore` stubs,
FFI methods, docs, earlier workarounds this change makes unnecessary. The plan
should delete them (the repo does this: "Drop citation composer strings the
rewrite no longer shows."). An empty **Deletes** line on a plan that replaces
behavior is a finding.

### Q8. The next named add

Read the later PRs in the map and the idea files the feature links. For each
one this code will be touched by: does the plan's shape give it a slot (a
registry entry, an enum case, a column already there), or will that PR have to
undo or fork this one? Only futures that are written down count.

### Q9. State and caches

New state has one owner. No shadow copy of a `QueryHandle` value, no second
cache key for the same list, no sibling views syncing through `onChange`.
Criteria §13 has the full list.

### Q10. The pattern lives in a skill

Skills are how the next agent learns the pattern.

- A plan that follows a skill follows **all** of it, not the convenient half.
- A plan that deviates from a skill: either the skill is wrong (the plan updates
  it in the DoD) or the plan is (finding).
- A plan that introduces a new pattern others will copy: it adds or updates the
  skill, rule, or domain doc in the same PR, or the next agent copies the old way.

### Q11. Reading cost

Count what a future reader has to learn: new files, types, concepts, FFI
methods, cache keys, L10n namespaces. Is there a version with fewer that does
the same job? Fewer concepts beats fewer lines.

### Q12. Step back

Ignore the rows. Given what the code already has, how would someone who knows
this codebase well build this PR? If that is materially different from the plan,
say so as one finding with the alternative shape, what it saves, and what it
costs. That is a **Reshape** verdict; it goes to the user.

### Q13. Where the better shape belongs

A better shape is not automatically this PR's job. Pick one and say it:

| Fit | When |
| --- | --- |
| **This PR** | Same files, a few more steps, the plan stays under its size limit |
| **Preparatory PR** | A refactor that makes this PR simple ("make the change easy, then make the easy change"). Goes in the map before this PR (`plan-spike`) |
| **Log it** | Real debt this PR does not touch or worsen. An idea file or the app-health backlog. Not this PR |

Do not widen the PR silently. A preparatory PR is a map change, and the user decides.

## Known traps

Mistakes that cost a later PR. Add a line when a **Landed** note or a rework
traces back to a plan. Point at the skill that owns the rule; do not restate it.

- Seeding vocabulary in a migration instead of `Install` — [`add-catalog-migration`](../add-catalog-migration/SKILL.md) "Do not".
- A migration test, a schema-hash constant — [`add-catalog-migration`](../add-catalog-migration/SKILL.md).
- A private `sqlInUse` instead of `deleteimpact` — [`add-catalog-delete`](../add-catalog-delete/SKILL.md).
- A new kit primitive for one call site — [`add-ui-component`](../add-ui-component/SKILL.md).
- A feature model copying `handle.value` into its own array — criteria §13.
- Localizing by matching English Go error strings — criteria §9.
- Domain logic in an `api/ffi/handlers` function — criteria §4.
