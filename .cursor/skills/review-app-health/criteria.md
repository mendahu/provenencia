# App health review — criteria

Detailed rubric for `.cursor/skills/review-app-health`. Read sections as needed while reviewing; do not dump this whole file into the user report.

---

## 1. Code cleanliness

**Look for**

- Unused imports, unused types/funcs, commented-out blocks, unreachable branches
- Unnecessary indirection: pass-through wrappers that add no policy, typing, or boundary
- One-off helpers duplicated nearby that should be one function—or the reverse: premature abstraction
- Outdated comments that contradict code; TODO/FIXME without owner or ticket that are stale
- Redundant nil/empty checks, duplicate switch arms, copy-paste with tiny drift
- Generated / vendored noise: do not “clean” `Platform/Generated/` or demand style nits there

**Provenencia notes**

- Prefer deleting dead FFI methods and FakeStore stubs that nothing calls
- Mac previews/fixtures may look “unused” from production paths—judge carefully

**Good finding shape** — “`FooWrapper` only calls `Bar` with same args; used once at X—inline or delete.”

---

## 2. Security

Whole-app posture (not only the latest diff).

**Look for**

- Path traversal / unsandboxed file access from project dir, ingest, derivatives, identity paths
- SQL: string-concatenated queries; trust of client-supplied identifiers without validation
- FFI boundary: trusting Mac-supplied bytes without size/type checks; leaking internal errors to UI
- Secrets in repo, logs, or protobuf payloads (tokens, absolute home paths in shared logs)
- Entitlements / TCC usage: broader than needed; insecure defaults in debug left on for release
- Ingest: MIME/type enforcement bypass, unbounded reads, zip/path tricks if archives appear
- Auth/identity: install identity mishandling, cross-project data bleed via held catalog session

**Optional delta pass**

For recent branch/uncommitted changes only, the Cursor security-review subagent may add signal. Still complete this whole-app pass.

**Do not** produce exploit PoCs. Describe risk and hardening direction only.

---

## 3. Performance

**Look for**

- Main-thread / UI-path blocking: sync heavy FFI, large image work, JSON/proto parse in `body`
- Catalog session: open-per-call regressions; accidental nested exclusive opens; badge refresh that serializes the world
- DB: N+1 queries, missing indexes for hot lists, full table scans on workspace open, unbounded `List` without pagination where lists grow
- Search/FTS: over-fetch, repeated reproject, query on every keystroke without debounce where required
- Derivatives/thumbnails: recomputing when cache should hit; loading full originals for grids
- SwiftUI: expensive work in `body`; missing stable identity; observation surfaces too wide (whole model vs field)
- Go: repeated Open/Migrate; holding locks across slow I/O; copying large buffers

**Output** — Prefer “opportunity” findings: measured or strongly reasoned bottleneck + a plausible faster shape.

---

## 4. Project structure & naming

**Look for**

- Vague folder/file names: `common`, `shared`, `project`, `util`, `utils`, `helpers`, `misc`, `stuff`, `temp`, `old`
- Packages that mix unrelated domains
- Files whose names hide responsibility (`Helpers.swift`, `Extensions.swift` grab-bags)
- Mac layout drift: screens outside `Features/<Name>/`; store/FFI types leaking into feature folders incorrectly; new flat files at App root
- Go: domain logic in `api/ffi` that belongs in `core/`; database queries outside `core/database/<table>/`

**Target shape**

| Area | Expectation |
| --- | --- |
| `core/<domain>/` | Semantic package names (`ingest`, `search`, `catalogsession`) |
| `core/database/<table>/` | Nested domain query packages |
| `api/ffi/handlers/` | Thin RPC adaptors |
| `macos/App/Features/<Feature>/` | View + model for one product flow |
| `macos/App/Platform/` | Store, FFI, L10n—not a screen |
| `macos/App/DesignSystem/` | Reusable UI primitives |
| `macos/ProvenenciaTests/` | Unit tests—not under `App/` |

Prefer renaming toward a **product concept** (see existing `Features/Catalog/` over a `Shared/` dump).

---

## 5. Separation of concerns

**Look for**

- Cross-domain imports: feature A types depending on feature B internals; `database` importing UI concepts; handlers importing unrelated domains to “just get it done”
- Cycles: A → B → A (Go packages or Swift modules/types)
- Shared bits living too deep—should be **hoisted** to a parent package or `Platform` / small `core/` library so the dependency tree stays a DAG
- God types: one model/store that knows every screen
- Business rules in SwiftUI `View` / `body`
- Duplicated discrete functionality that should be a reusable utility **once** there are ≥2 real call sites and a clear name
- UI chrome duplication (menus, lists, panels) — review under **§12 UI component organization**; still flag here when a feature module depends sideways on another feature’s private control instead of a shared primitive

**Principle** — Modules compose elegantly; boundaries are deliberate; shared code moves **up**, not sideways via spaghetti.

---

## 6. Idiomatic patterns

Fight the grain = debt.

**Go**

- Standard layout, `context` where needed, wrapping errors with `apperr`, table-driven tests
- Follow project skills for migrations, catalog queries, refs, seeded vocab, FFI handlers

**Swift / SwiftUI**

- Views as value types; cheap `body`; logic on models; `GenealogyStore` + `FakeStore`
- AppKit only as wrapped escape hatches
- Navigation via `WorkspaceLocation` / `WorkspaceNavigation.go(to:)` for restoreable places
- Prefer framework/`Observation` patterns the codebase already uses over bespoke publishers

**SQLite / catalog**

- Migrations + `user_version`; session serialization; no ad-hoc parallel open policies

**Flag** — Patchy helpers that reimplement what SwiftUI, the store protocol, or a skill already prescribe.

---

## 7. Test coverage

**Look for**

- Domain changes in `core/` without Go tests
- New store APIs without FakeStore + model tests where behavior is non-trivial
- Handlers without `runRPC`-style coverage for happy path + important errors
- Gaps: error paths, session lifecycle, ingest policy, search mapping, navigation restore
- Brittle tests asserting raw English instead of `String(localized: L10n.…)`
- Tests that hit dylib / real SQLite from Mac unit target (should not)

**Balance** — Do not demand 100% lines. Demand coverage on **boundaries and rules** that AI edits tend to break.

---

## 8. Error handling & messaging UX

**Look for**

- Swallowed errors (`_ = …`, empty `catch`, ignoring `lastRefreshError`)
- Generic copy (“Something went wrong”) where a stable code + specific `L10n.Errors` exists or should
- Wrong placement: alert when inline field validation fits; toast for form-field mistakes; modal for transient refresh failure
- Internal/debug strings shown to users; missing mapping for new `apperr` codes
- Loss of context: error doesn’t say *which* file/project/field failed when the UI knows

**Target** — Specific, context-aware messaging at the point of action (inline, floated, or toast)—consistent with existing workspace/onboarding patterns.

---

## 9. Internationalization

**Look for**

- Hard-coded user-facing strings in `macos/App` views/models (English literals in `Text`/`Button`/`Label`/`navigationTitle`/panels)
- Call sites using raw catalog key strings instead of `L10n.*`
- New FFI error codes without `L10n.Errors` + catalog entries
- Catalog / `L10n.swift` drift (key in one place only)
- Localization by matching English Go error strings

**Allow** — Non-copy data (user names, paths, refs, IDs); brand `Provenencia` in values; accessibility identifiers as dotted stable tokens; Generated protobuf.

**Fix path** — `.cursor/skills/add-localized-string`.

---

## 10. Docs / skills / rules drift

Agents treat project docs and Cursor guidance as truth. Stale guidance **amplifies** debt under AI-assisted edits.

**Look for**

- Skills or rules that point at deleted/renamed paths, obsolete APIs, or retired workflows
- Contradictions: skill says A, rule says B, `docs/*.md` says C, code does D
- “Blessed pattern” docs that no longer match the tree (catalog session, L10n, workspace nav, FFI handlers, migrations)
- Orphan skills with no matching code area—or code areas with repeated AI mistakes and **no** skill/rule when one would help
- Deployment-plan / idea docs marked current that describe shipped-and-changed behavior (archive or update)
- Duplicate guidance copied in three places that has drifted apart—prefer one authoritative doc + thin skill/rule pointers

**How to sample**

1. List `.cursor/skills/*/SKILL.md` and `.cursor/rules/*.mdc`.
2. For each skill touched by recent product work (or a rotating sample), open linked paths and confirm they exist and still describe reality.
3. Spot-check high-churn docs: `application-stack.md`, `macos-client-patterns.md`, active spike notes vs `Features/` and `core/`.

**Good finding shape** — “`use-catalog-session` still mentions `openProjectCatalog`; code only has `withProjectCatalog`—update skill or restore name.”

**Remediation** — Update the authoritative doc/skill/rule, or move obsolete notes to `docs/ideas/archive/` / spike archive. Do not leave knowingly false agent instructions.

---

## 11. Accessibility

Mac client AT and UI-test hygiene. Notes: [`docs/macos-client-patterns.md`](../../../docs/macos-client-patterns.md) §5; `.cursor/rules/macos-client.mdc`.

**Look for**

- Interactive controls missing stable `.accessibilityIdentifier("dotted.name")` when they are (or will be) UI-tested or are primary actions
- Icon-only buttons/toolbars without an accessibility label (or labeled only by invisible English that bypasses `L10n`)
- UI tests or instructions that find controls by localized title instead of identifier
- Images/icons: decorative vs informative—missing labels on meaningful imagery; noisy labels on pure decoration
- Focus / keyboard: custom controls that trap focus, omit actions available only via hidden gesture, or break standard Mac key loops without cause
- Omnibar, sidebar, source page, and onboarding—high-traffic surfaces—sampled for VoiceOver-readable names and traits
- Contrast / hit-target nits only when clearly broken in DesignSystem usage (don’t turn this into a full visual QA pass)

**Allow** — Identifiers as stable dotted tokens (not localized). Decorative assets intentionally hidden from AT.

**Good finding shape** — “Toolbar refresh is icon-only at X with no label/identifier—add `L10n` label + `workspace.toolbar.refresh` identifier.”

---

## 12. UI component organization

Catch LLM- and rush-driven UI debt against Frost layers (**design system / recipes / snowflakes** — [`docs/ideas/design-system-hardening.md`](../../../docs/ideas/design-system-hardening.md)). This dimension **samples and flags**; for a full classify → cousins → compose-down → small-PR plan on one type, hand off to [`evaluate-ui-component`](../evaluate-ui-component/SKILL.md). For adding new chrome, [`add-ui-component`](../add-ui-component/SKILL.md).

**Look for**

- **Orphans** — `PV*` types with zero production call sites under `Features/` / `Platform/` (previews/tests alone don’t count).
- **Near-duplicates** — panels, menus, chips, list shells, empty states, or dialogs that share chrome but diverge only in open gesture or payload.
- **Snowflake reimplementing DS** — feature-private control that redraws an existing `PV*` (dialog footer, menu host, chip) instead of composing it.
- **Recipe trapped as snowflake** — reusable product mapping living under `Features/<OneDomain>/` that a second surface already copied (promote to a recipe).
- **Wrong layer** — catalog/store/feature meaning inside `DesignSystem/`; or a one-off promoted to `PV*` with a single call site; or a preemptive recipe with one caller.
- **Wrapper sprawl** — separate panel stacks for open-policy differences instead of one panel + composable open policies.

**How to sample**

1. Inventory `macos/App/DesignSystem/Components/**` (`PV*` types).
2. Grep `Features` / `Platform` for call sites—flag zero-use orphans.
3. Spot-check high-churn UI for private `*Panel` / `*Sheet` / `*Row`; diff against nearest `PV*`.
4. Ask: design system, recipe, or snowflake? If a third screen needed this tomorrow, compose or promote—don’t copy. For a single hot type, prefer [`.cursor/skills/evaluate-ui-component`](../evaluate-ui-component/SKILL.md) over expanding this whole-app pass.

**Provenencia notes**

- Target: content-agnostic `PV*` in `DesignSystem/`; recipes as thin product maps when ≥2 call sites; snowflakes under `Features/<Name>/`; store/FFI in `Platform/`.
- Prefer **one floating-menu kit** (`PVContextMenu` / `PVSelect` / jump menu—see [`archive/spike-5`](../../../docs/deployment-plan/archive/spike-5/completed.md#s5-10--unify-floating-menus)). Omnibar and `PVComboBox` keep separate hosts by design.
- Dialogs: compose `.pvDialog` / `.pvConfirm` / `.pvConfirmSheet`; don’t redraw scrim/footer chrome in a feature sheet.
- Do not invent new DesignSystem types for a single unproven call site; hoist snowflake → recipe (or into DS if agnostic) at ≥2 real surfaces.

**Good finding shape** — “New overlay menu reimplements card chrome instead of `PVContextMenuPanel`; wire it through the shared kit with a custom row payload.”

**Allow** — Deliberate snowflakes with no second consumer yet; AppKit escape hatches wrapped once; Generated / preview-only fixtures.

---

## Priority guide

Action items land in one of three groups (see skill Output format). When unsure,
prefer **medium** over **high**. An empty high-priority section is a fine outcome.

| Priority | Use when |
| --- | --- |
| **High** | Security risk, data loss/corruption risk, severe UX breakage, clear layering violation actively causing bugs, guidance that would cause unsafe/wrong agent edits on a critical path, primary flows hard or impossible for AT |
| **Medium** | Meaningful debt, perf likely to hurt dogfood, structural smell that will multiply under AI edits, docs/skills that systematically mis-train agents, missing coverage on important boundaries, orphaned or near-duplicate UI primitives that encourage copy-paste |
| **Low** | Nits, naming polish, optional cleanup, speculative perf, missing identifiers on low-traffic controls, minor doc stale phrasing |

### Sizing action items

Each numbered item should be the **smallest meaningful PR**:

- One concern, one done state, reviewable without a design debate
- Split “fix error UX across the app” into per-surface or per-code items
- Merge only when splitting would leave a non-compiling or non-valuable intermediate
- Title should read like a PR subject; Problem / Consequence / Benefit do the rest
