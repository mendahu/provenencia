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
   component, cache key, L10n namespace, test helper — run Q1–Q11.
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
a registry. Propose it, usually as a preparatory PR.

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
  in this PR if small, else a preparatory PR.
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
| **This PR** | Same files, a few more steps, the PR stays one reviewable behavior |
| **Preparatory PR** | A refactor that makes this one simple. A map change; the user decides |
| **Log it** | Real debt this change does not touch or worsen. An idea file. Not this PR |

Do not widen a PR silently.

## Severity in this lens

| Severity | Use when |
| --- | --- |
| **blocker** | Debt that is expensive to undo once merged: wrong layer for domain logic, a second mechanism beside an existing one, bypassing a registry CI expects, a schema shape the next named add must fight, a split brain that can show the wrong row |
| **major** | A clear structural win in this PR's files: reuse an existing helper, delete what this made dead, update the skill it made false |
| **minor** | Naming polish, a small extraction, a comment |
| **nit** | Taste |

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

## Good finding shape

The shape to copy. Names in the example are illustrative.

"`SourcesListModel.filtered` re-implements the evidence-status grouping that
`core/database/sourcegraphprogress` already returns; read the counts from that
query and drop the Swift loop." Or: "This is the third `switch subjectKind`
that maps a kind to a mark; add the mark to the subject-kind registry instead
(preparatory PR)."
