# Maintainability lens

The question: **this works — is there a shape that fits the codebase better and
costs the next change less?** Not more correct. Fewer copies, fewer special
cases, fewer places to look, and the next named add already has a slot.

Every patch an agent ships is read by every agent after it. A codebase that
repeats a few regular patterns is cheap to read and cheap to change. A web of
one-off fixes is expensive on both, forever. Implementing models drive straight
at the instruction and do not look around. Look around.

Used by [`review-plan`](../../review-plan/SKILL.md) on a plan, brief, or map,
and by [`review-pr`](../SKILL.md) on a diff or an area of the app.

## Read first

- The plan (`prs/S#-NN-<slug>.md`): **Codebase fit**, **Decisions**, **Design**.
  On a diff, the code must match them; on a plan, they are what you review.
- Every skill in the plan's **Skills to load**, including its **Do not** list.
- [`docs/design-system-layers.md`](../../../../docs/design-system-layers.md) and
  [`docs/macos-client-patterns.md`](../../../../docs/macos-client-patterns.md)
  when `macos/` is touched; [`docs/application-stack.md`](../../../../docs/application-stack.md)
  for where Go and Swift split.
- The code **around** the change: the package, the feature folder, its
  siblings, the registry it should feed. Not only the changed lines.

## How to work

1. Read the plan, then the change (the diff against its base, or the plan's
   Design and Steps).
2. For each thing the change adds — file, type, function, table, FFI method,
   component, cache key, L10n namespace, test helper — run Q1–Q11 and Q14.
   **Search before you claim.** Every finding names the code you found, or the
   map row or idea file that names the future add.
3. Then step back from the parts (Q12) and place each finding (Q13).

## Questions

### Q1. Does it already exist?

Search for an existing type, function, file, table, component, FFI method, L10n
namespace, or test helper doing the same job. Search by behavior, not just by
name: the field it reads, the SQL it runs, the interaction it owns.

- Exists and fits → reuse it.
- Does 80% → extend it (a parameter, a case, a slot) when the extension stays
  honest for the existing callers.
- On a plan, a thing missing from **Codebase fit** is itself a finding: the
  writer did not search. On a diff, a new thing the plan's **Codebase fit** did
  not list is scope the implementer added.

### Q2. Does a registry or declared model already own this?

Provenencia keeps cross-cutting facts in one declared place, and CI holds the
code to it. Special-casing where an entry belongs is a patch.

| Concern | Owner |
| --- | --- |
| Tables, FKs, kinds | `core/database/catalogmodel` ([`add-catalog-model`](../../add-catalog-model/SKILL.md)) |
| Delete, inbound probes, releases | `core/database/deleteimpact` ([`add-catalog-delete`](../../add-catalog-delete/SKILL.md)) |
| Audit effects and scopes | `core/database/audit` ([`add-audit-scope`](../../add-audit-scope/SKILL.md)) |
| Seeded terms, types, properties | `subjectvocab`, `sourcevocab` registries + `Install` ([`add-seeded-vocabulary`](../../add-seeded-vocabulary/SKILL.md)) |
| Omnibar hit kinds | `core/search` registry ([`add-searchable-kind`](../../add-searchable-kind/SKILL.md)) |
| Workspace places, list/detail caches | `PlaceRegistry`, `CatalogQueryKey`, `CatalogQueryRegistry` ([`add-workspace-place`](../../add-workspace-place/SKILL.md)) |
| Navigation history | `WorkspaceLocation`, `WorkspaceNavigation` ([`add-workspace-location`](../../add-workspace-location/SKILL.md)) |
| Go wire vocabulary in Swift | `Platform/ConclusionVocabulary.swift` (`.cursor/rules/macos-client.mdc`) |
| Copy | `L10n` + `Localizable.xcstrings` ([`add-localized-string`](../../add-localized-string/SKILL.md)) |
| UI chrome | `DesignSystem` layers ([`add-ui-component`](../../add-ui-component/SKILL.md)) |
| Error codes | `core/apperr` + `L10n.Errors` |

Also the reverse: if this is the **third** hand-rolled list of the same kind of
thing (three `switch kind` blocks that must stay in step), the better shape is
a registry. Propose it in this PR (`Fit: this PR`) when the PR adds an entry to
the list or touches one of the copies; ship-pr fixes those on the same pass.

### Q3. Is the change at the layer that owns the concern?

- Domain rules live in Go; the Mac client stays thin. Logic in a SwiftUI
  `body` or model that belongs in `core/`, or in an `api/ffi/handlers`
  function that belongs in `core/`, is a finding. Handlers are thin adaptors.
- Queries live in `core/database/<table>/`, not in a feature package.
- A fix in one call site for a defect in a shared component belongs in the
  component, so every caller gains it. Precedent: S10-17 fixed letter cycling
  in `PVSelect` itself, not in the name editor.
- Red flags: `if kind == …` in shared code, a Bool parameter added to a shared
  function for one caller, a feature reaching into another feature's internals,
  an import cycle, a type that knows every screen.

### Q4. One concept, one name, one place

- Names match the domain docs (Source, Artifact, Citation, Observation, Subject,
  Identity Claim, Property, term…). A synonym for an existing concept is a
  finding; so is one word for two concepts.
- New files sit beside their kin and follow the neighbors' naming:

  | Area | Shape |
  | --- | --- |
  | `core/<domain>/` | Semantic package names (`ingest`, `search`, `catalogsession`) |
  | `core/database/<table>/` | One nested query package per table |
  | `api/ffi/handlers/` | Thin RPC adaptors |
  | `macos/App/Features/<Feature>/` | View + model for one product flow |
  | `macos/App/Platform/` | Store, FFI, L10n — not a screen |
  | `macos/App/DesignSystem/<Layer>/<Name>/` | One folder per control; no category nesting |
  | `macos/ProvenenciaTests/` | Unit tests — not under `App/` |

- No `util`, `helpers`, `common`, `shared`, `misc`, `Extensions.swift`
  grab-bags. A future reader looking for this behavior finds it in **one obvious
  place**, named for what it does.

### Q5. Copies and near-copies

- Code that copies an existing function, view, query, or test and tweaks it:
  name both. Second copy → it must have a reason to diverge. Third → extract,
  in this PR (`Fit: this PR`) when the PR adds or touches a copy, across every
  package that holds one. A plan Out line that deferred it does not change the Fit.
- UI: a feature-private control that re-solves an interaction a `PV*` already
  owns (hover, press, selected, disabled, hit area, tooltip, AT traits) is a
  near-copy even when it is a `private func`. Judge by what the user can do,
  not the type name. Method: [`evaluate-ui-component`](../../evaluate-ui-component/SKILL.md) §3–4.
- Copy-paste drift this change makes worse is a finding; fixing old drift it
  does not touch is "log it" (Q13).

### Q6. Not too much, either

Speculative generality is debt too:

- A protocol with one conformer, a config option nobody sets, a generic engine
  for one call site, an extension point no map row or idea names.
- A new `DesignSystem` component or recipe with one caller. The layer rule:
  snowflake until a second real call site.
- A wrapper that only forwards and adds no policy, typing, or boundary.

Generalize for a **second real caller** or a **future add named in the map or
`docs/ideas/`**. Not for "might be useful".

### Q7. Leave it smaller

What does this change make dead? Code paths, strings, tests, `FakeStore` stubs,
FFI methods, docs, earlier workarounds. They go in this PR. The repo does this:
"Drop citation composer strings the rewrite no longer shows." On a diff, also
look for what the change itself left behind: unused imports, commented-out
code, a TODO with no owner, an outdated comment.

### Q8. The next named add

Read the later PRs in the map and the idea files the feature links. For each
one this code will be touched by: does this shape give it a slot (a registry
entry, an enum case, a column already there), or will that PR have to undo or
fork this one? Only futures that are written down count.

Count what the next named add must touch. A per-type field, `case`, or
nil-checked append that every new type repeats (`DateChange`, `NameChange`,
`DepictionChanges`, each read by the same append blocks), where one list or
one entry would serve them all, is a fork the next add pays for. When the map
or an idea file names that next type, it is a **blocker**.

### Q9. State and caches

One owner per concern; a parent machine instead of sibling ping-pong; caches
that stay the single source of truth. Target: view → feature model (`enum Phase`
+ drafts) → `GenealogyStore` / session mutations → `QueryHandle` for display data.

- **Sibling-owned state that must interact.** Two peer views each hold
  `@State` or a private `@Observable` slice and sync via `onChange`,
  `NotificationCenter`, PreferenceKey tunnels, or "did the other one finish?"
  flags. Hoist to the nearest parent model; express transitions as an enum
  (precedent: `CitationComposerModel.Phase`).
- **Split brains.** The same fact in two places (view and model; list and
  detail; feature model and `QueryHandle`). Ask which write wins after a race,
  a Back navigation, or a failed save.
- **Boolean piles** (`isLoading && !isPicking && hasDraft`) that admit
  impossible combinations → one enum with associated values.
- **Shadow caches.** A feature model copying `handle.value` into its own array.
  One `CatalogQueryKey` per list; no payload that embeds another key's rows.
- **Wrong update tool.** `setQueryValue` used as a write-through store (it
  no-ops on a cold key); invalidate skipped after create/delete; two siblings
  patching the same key after one mutation; a mutation the
  `CatalogQueryRegistry` map does not cover.
- **Bypass.** `.task { load() }` into local state while `session.apply(location:)`
  already warms the place; `session.query()` from `body`.
- **Cache vs workflow.** Draft or dirty form state on the session; a
  `QueryHandle.status` used as a wizard phase.
- **Refetch where the write already knows.** When a write reports what it
  changed and the new value (an id set, a time, a field), the cache patches
  the affected rows in place through the cache layer, declared once so the
  next list reuses it. Invalidate-and-reload is for changes the client cannot
  know. A whole list or page rebuilt to move one field, a reload of data that
  is not on screen, or a write result the cache ignores is a finding. A
  one-off patch written in a feature or in one mutation branch, instead of in
  the cache layer, is a second mechanism (Q14).

Allow: ephemeral view-only chrome (hover, focus, scroll, one-shot animation
flags), preview fixtures, deliberate stale-while-revalidate on navigation reads.
Do not demand a state-machine library; a small `enum Phase` is the grain.

### Q10. The pattern lives in a skill

Skills are how the next agent learns the pattern; stale guidance multiplies debt.

- Code that follows a skill follows **all** of it, not the convenient half.
- Code that deviates from a skill: either the skill is wrong (this PR updates
  it) or the code is (finding).
- A new pattern others will copy: this PR adds or updates the skill, rule, or
  domain doc, or the next agent copies the old way.
- A skill, rule, or `docs/*.md` this change makes false (a renamed type, a
  moved path, a retired API) is updated in this PR.

### Q11. Reading cost

Count what a future reader has to learn: new files, types, concepts, FFI
methods, cache keys, L10n namespaces. Is there a version with fewer that does
the same job? Fewer concepts beats fewer lines.

### Q12. Step back

Ignore the parts. Given what the code already has, how would someone who knows
this codebase well build this? If that is materially different, say so as one
finding with the alternative shape, what it saves, and what it costs. On a plan
that is a **Reshape** verdict; on a diff it goes to the orchestrator as a
follow-up unless it is blocking.

### Q13. Where the better shape belongs

A better shape is not automatically this change's job. Pick one and say it:

| Fit | When |
| --- | --- |
| **This PR** | Same files, a few more steps, the PR stays one reviewable behavior. Also any extract-and-dedupe refactor (helper, registry, fixture) whose copies or list this PR adds to or touches, even across other packages |
| **Preparatory PR** | A reshape a researcher would see, or one that needs a product call. The user decides |
| **Log it** | Real debt this change does not touch or worsen. An idea file. Not this PR |

Do not widen a PR silently.

### Q14. One mechanism per cross-cutting job

Some jobs recur across the app: presenting a kind of value (timestamps,
dates, names, counts, refs, file sizes), updating the cache after a write,
resolving what a write touched, mapping an error to copy. Each has one
mechanism, and every caller goes through it. The cheapest time to keep it
that way is the PR that adds the second caller.

- **Look for a parallel path.** A new type, struct, or helper that produces
  the same kind of output as an existing one, shaped for one screen or one
  caller. A value object that bundles the display strings for one view next
  to an existing formatter. A formatter variant with different inputs than
  its siblings (one takes a locale, the others read the process default). It
  does not matter that it is small, well named, or in the right folder: two
  ways to do one job is the debt.
- **Formatting is not styling.** A shared formatter takes the value, the
  display option, and the context (locale, calendar, time zone, clock) and
  returns text. Icons, colors, fonts, placeholders, and layout belong to the
  caller that composes the text into its view. A formatter that returns a
  view, or a view-model that owns formatting rules, mixes the two.
- **Hidden inputs.** A shared function that reads the process locale, the
  wall clock, or the current time zone implicitly cannot be tested for every
  style, so callers re-implement it to make it testable. Context is a
  parameter with a `.current` default.
- **Inside one package too.** A second way to do a job the package already
  does is a second mechanism even when it is private and short: a private
  `parse` that unmarshals the same format beside the exported `Validate`, a
  second scan of the same row shape, a local trim or normalize rule beside the
  one the store applies. The next rule added to one path (a version check, a
  size limit) silently misses the other.
- **The fix shape.** Extend the one mechanism (a style case, a context
  parameter, a declared cache hook, a split of the exported function into its
  steps) and move the existing callers onto it in the same PR. Delete what that
  makes dead (Q7).

## Severity in this lens

| Severity | Use when |
| --- | --- |
| **blocker** | Debt that is expensive to undo once merged: wrong layer for domain logic; a second mechanism beside an existing one (Q14), including a small or private one in the same package and one shaped for one screen; bypassing a registry CI expects; a schema shape, or a per-type field or case, that the next named add must fight or repeat (Q8); a split brain that can show the wrong row; a cache that reloads what the write already reported (Q9); a cost on a common path that grows with the catalog or its history (rate it with the performance lens) |
| **major** | A clear structural win in this PR's files: reuse an existing helper, delete what this made dead, update the skill it made false |
| **minor** | Naming polish, a small extraction, a comment |
| **nit** | Taste |

When unsure between two severities, pick the higher one; the orchestrator can
lower it with evidence, but cannot see a problem rated below its weight.

**Precedent is not a mitigation.** Rate a finding by the debt after merge. A
change that copies, extends, or widens an existing weak pattern adds to it;
"the code already did this" lowers nothing. Say which existing code shares
the problem so the fix plan can include it or log it.

**The plan is not a mitigation either.** A Decision that names the weaker
shape (three per-type fields, a private decoder) does not lower the severity.
Rate the debt, cite the Decision it contradicts, and the orchestrator takes it
to the user (see [`review-pr`](../SKILL.md) "Rate the code as merged").

## Known traps

Mistakes that cost a later PR. Add a line when a **Landed** note or a rework
traces back to a shape a review should have caught. Point at the skill that owns
the rule; do not restate it.

- Seeding vocabulary in a migration instead of `Install` — [`add-catalog-migration`](../../add-catalog-migration/SKILL.md) "Do not".
- A migration test, a schema-hash constant — [`add-catalog-migration`](../../add-catalog-migration/SKILL.md).
- A private `sqlInUse` instead of `deleteimpact` — [`add-catalog-delete`](../../add-catalog-delete/SKILL.md).
- A new kit primitive for one call site — [`add-ui-component`](../../add-ui-component/SKILL.md).
- A feature model copying `handle.value` into its own array — Q9.
- Domain logic in an `api/ffi/handlers` function — Q3.
- New files in legacy `DesignSystem/Components/Core|Forms|Research/` or loose at a layer root — Q4.
- A screen-shaped value type that formats beside an existing shared formatter — Q14.
- Invalidating a list to move one field the write already reported — Q9.
- A private decoder beside the package's exported validator for the same format (S10-32 `locator.parse` beside `Validate`) — Q14.
- One result field per value type, each with its own append, when the map names more types (S10-32 `DateChange` / `NameChange` / `DepictionChanges`) — Q8.

## Good finding shape

The shape to copy. Names in the example are illustrative.

"`SourcesListModel.filtered` re-implements the evidence-status grouping that
`core/database/sourcegraphprogress` already returns; read the counts from that
query and drop the Swift loop." Or: "This is the third `switch subjectKind`
that maps a kind to a mark; add the mark to the subject-kind registry instead
(this PR)."
