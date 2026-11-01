# S14-15 — Xcode synchronized folders

**Status:** Planned
**Spike / slice:** Spike 14 · Slice 5 — Audit fixes
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: only the Xcode project file and contributor docs change; the app is built from the same files)
**Board:** —
**Depends on:** —
**Blocks:** — (lands best before the other Slice 5 PRs that add Swift files: S14-14, S14-16, S14-17, S14-20, S14-21, S14-22, S14-27)
**Branch:** `s14-15-xcode-synchronized-folders`
**Version bump:** none — a build-system change with no researcher-visible change, as S14-24 took none ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md)
**Rules that bind:** `.cursor/rules/swift-tests.mdc`
**Context:** audit [A-34](../app-health-audit.md#a-34--projectpbxproj-is-maintained-by-hand); working rule "Audit items keep their IDs"; [`macos-client-patterns.md`](../../../macos-client-patterns.md); `macos/App/DesignSystem/README.md` § "Xcode project registration". **Needs a Mac with Xcode 16 or later** (map scope entry).
**Planned:** 2026-11-01 against `main` @ `067be77`
**Review:** [`S14-15-xcode-synchronized-folders.review.md`](S14-15-xcode-synchronized-folders.review.md) — Fits / Ready, 2 rounds

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

A Swift file saved under `macos/App` or `macos/ProvenenciaTests` is in the build with no edit to `project.pbxproj`. The four app source folders and the test folder become Xcode synchronized folders, the 425 hand-written Swift entries (each a `PBXFileReference`, a `PBXBuildFile`, a group child and a Sources phase line) leave the project file, and the skills and the Cursor rule stop telling agents to hand-write project objects. The app compiles exactly the same Swift files and bundles exactly the same resources as before.

**Why now.** Audit A-34: every new Swift file is four hand-written objects edited by agents with no Xcode to check them, and `project.pbxproj` is a conflict hotspot across parallel spikes. The rest of Slice 5 adds Swift files; this PR goes first so those PRs touch no project file.

**What stays.** Every Swift source, resource, build setting, target, scheme, package reference, the Go-core build phase and `MARKETING_VERSION`. `macos/App/Resources` stays a classic group (D-2).

## 2. User stories

**US-1.** As a contributor (human or agent), when I save a new `.swift` file under `macos/App/<App|Platform|Features|DesignSystem>/…` or `macos/ProvenenciaTests/`, it compiles into its target with no edit to `project.pbxproj`, so I cannot break the project file by hand.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A file `macos/App/Features/Sources/SyncProbe.swift` containing `enum SyncProbe { static let value = 1 }` *(new)* (temporary) is added; nothing else changes | `xcodebuild build-for-testing` compiles it: the path is in the app target's `Provenencia.SwiftFileList` |
| AC-1.2 | A file `macos/ProvenenciaTests/SyncProbeTests.swift` *(new)* (temporary), shaped like `PVFontTests.swift`: `import Testing`, `@testable import Provenencia`, `@Suite struct SyncProbeTests { @Test func probeIsBuilt() { #expect(SyncProbe.value == 1) } }`, is added | `xcodebuild test … -only-testing:ProvenenciaTests/SyncProbeTests` runs it and passes |
| AC-1.3 | After the conversion | `grep -c 'sourcecode.swift' macos/Provenencia.xcodeproj/project.pbxproj` prints `0` |

**US-2.** As the researcher, the app I run after this PR is the app I ran before it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `build-for-testing` on `main` @ `067be77` and on the branch, each into its own `-derivedDataPath` | The sorted contents of each target's `.SwiftFileList` are identical. Lines not under `macos/App` or `macos/ProvenenciaTests` (generated files such as `GeneratedAssetSymbols.swift`) are filtered out before the count: 308 app files, 117 test files |
| AC-2.2 | The same two builds | The sorted listing of `Build/Products/Debug/Provenencia.app/Contents/Resources` (relative paths) is identical: no `README.md` or `MARKS.md` in it, `Fonts/` still a subfolder with 15 `.ttf` |
| AC-2.3 | `scripts/check-pr.sh` on a Mac | `check-pr: PASS`, including `xcode-test` and `swift-warnings` |

**US-3.** As a contributor reading the guidance, the skills, rule and docs describe the synchronized folders and no longer describe hand-written project objects.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | The sweep command below, run before STEP-7 and after STEP-8 | Before: 12 hits (the lines D-8 rewrites). After: no hit |
| AC-3.2 | `python3 scripts/check-guidance-names.py` | Passes |
| AC-3.3 | On the converted tree, three plants, each made and undone on its own: (1) `macos/App/Features/Sources/NOTES.md` *(new)* (temporary); (2) `macos/App/Extra/Probe.swift` *(new)* (temporary); (3) `git mv macos/App/Features/GraphCanvas/README.md macos/App/Features/GraphCanvas/README.txt` | While each is present, `python3 scripts/check-guidance-names.py` exits 1 and prints exactly: (1) `macos/App/Features/Sources/NOTES.md: not Swift and not in a membership exception of the macos/App/Features synchronized folder`; (2) `macos/App/Extra: not a synchronized folder or group in the Xcode project`; (3) `macos/App/Features/GraphCanvas/README.txt: not Swift and not in a membership exception of the macos/App/Features synchronized folder` and `macos/App/Features/GraphCanvas/README.md: listed in a membership exception but not on disk`. Undone, it prints `check-guidance-names: OK` |

Sweep command (AC-3.1; plain `|` alternation):

```bash
rg -n 'PBXFileReference|PBXBuildFile|four objects|registered in the Xcode project|Register new files|Resources phase|manually-authored|Wire new files|wiring a .swift file|wiring new Swift test files' .agents .cursor macos/App/DesignSystem/README.md docs/macos-client-patterns.md
```

## 3. Scope

| In | Covered by |
| --- | --- |
| `project.pbxproj` on `objectVersion = 77` with five synchronized root groups (D-2, D-4, D-5) | STEP-2, AC-1.3, AC-2.1 |
| Membership exceptions keep the three `.md` files out of the app bundle (D-3) | STEP-2, AC-2.2 |
| Same compiled files and bundled resources, proved by equality (D-6) | STEP-1, STEP-3, AC-2.1, AC-2.2 |
| A new file builds with no project edit (A-34) | STEP-4, AC-1.1, AC-1.2 |
| `check-guidance-names.py` checks the synchronized folders against the project file on every host (D-7) | STEP-6, AC-3.3 |
| `add-swift-test`, `add-ui-component`, `add-workspace-place`, `.cursor/rules/swift-tests.mdc` drop their pbxproj steps; `docs/design-system-layers.md` names the README exception (D-8) | STEP-7, AC-3.1 |
| `README.md` names Xcode 16 as the minimum; `macos-client-patterns.md` and the DesignSystem README describe the folders (D-4, D-8) | STEP-8, AC-3.1 |

| Out | Home |
| --- | --- |
| Any Swift, Go or resource change, including loading fonts from the bundle root so `Resources/` could be synchronized too | not planned (D-2) |
| Converting `macos/App/Resources` to a synchronized folder | not planned (D-2) |
| A check that fails on a `sourcecode.swift` entry in the project file | not planned (D-7) |
| One synchronized root for all of `macos/App` with folder exceptions (review M-3) | not planned (D-2; D-7's check closes the new-folder trap) |
| Moving the three `.md` files out of `macos/App` | not planned (D-3) |
| Raising `MACOSX_DEPLOYMENT_TARGET`, `SWIFT_VERSION` or any build setting | not planned |
| Updating historical plans, ship logs and audits that mention hand-written entries | not planned: they record what was true then |

## 4. Current state

Verified at `067be77`.

| Area | Where | Today |
| --- | --- | --- |
| Project format | `macos/Provenencia.xcodeproj/project.pbxproj` line 6 `objectVersion = 60;`; PBXProject `AA0000000000000000000000` `compatibilityVersion = "Xcode 15.0";`, `LastUpgradeCheck = 2660` | 2,930 lines; no `PBXFileSystemSynchronizedRootGroup` (`grep -c fileSystemSynchronized` is 0) |
| Swift entries | same file, `PBXBuildFile` (lines 9–443), `PBXFileReference` (469–906), `PBXGroup` (927–2081), Sources phases | 425 `.swift` files, each with a build file, a file reference, a group child and a phase line: app Sources phase `AA0000000000000000000060` lists 308, test Sources phase `AA00000000000000000000B7` lists 117. Measured with a parser at the base: the set of project Swift paths equals the set of `.swift` files on disk under `macos/App` and `macos/ProvenenciaTests`; none is in both targets; no app file is outside `App/`, no test file outside `ProvenenciaTests/`; no two Swift files share a basename |
| Groups | main group `AA0000000000000000000040` → `App` `AA0000000000000000000041` (`path = App`) → `App` `…45`, `Platform` `…46`, `Features` `…47`, `Resources` `…49`, `DesignSystem` `550D63589090B3E5D65DBED2`; main group → `ProvenenciaTests` `AA00000000000000000000B5`, `Core` `AA0000000000000000000043`, `Products` `AA0000000000000000000042` | Every group under these has `path` equal to its folder name (no virtual groups except `Products`); `macos/App` holds exactly the folders `App`, `DesignSystem`, `Features`, `Platform`, `Resources` and no loose files |
| Resources | group `AA0000000000000000000049`; Resources phase `AA0000000000000000000061` | Children: `Assets.xcassets`, `Localizable.xcstrings`, `InfoPlist.xcstrings` (all three in the phase), `Provenencia-Bridging-Header.h` (`SWIFT_OBJC_BRIDGING_HEADER`), `Provenencia.entitlements` (`CODE_SIGN_ENTITLEMENTS`), and `Fonts` `1B183C69AA74C2C22006FC05`, a **folder reference** (`lastKnownFileType = folder`) in the phase, copied as `Contents/Resources/Fonts/` |
| Font loading | `macos/App/DesignSystem/Tokens/PVTypography.swift` line 183 | `Bundle.main.urls(forResourcesWithExtension: "ttf", subdirectory: "Fonts")`: fonts must stay in a `Fonts` subfolder of the bundle |
| Non-Swift files in source folders | `macos/App/DesignSystem/README.md` (file reference `6AAC04817D6F4CCC9CDC9B17` in the DesignSystem group, in no phase), `macos/App/DesignSystem/Recipes/Marks/MARKS.md` and `macos/App/Features/GraphCanvas/README.md` (not in the project) | None is bundled today. `macos/ProvenenciaTests` holds only `.swift` |
| Core | group `AA0000000000000000000043` (`path = Core`), `libprovenencia.dylib` `AA0000000000000000000018` in Frameworks and Embed Libraries; shell phase `AA00000000000000000000C0` "Build Go core" | Unchanged by this PR |
| Targets | `PBXNativeTarget` `AA0000000000000000000050` (Provenencia), `AA00000000000000000000B6` (ProvenenciaTests); scheme `macos/Provenencia.xcodeproj/xcshareddata/xcschemes/Provenencia.xcscheme` refers to both by these IDs | No `fileSystemSynchronizedGroups` key |
| CI | `.github/workflows/macos-test.yml` job `test`, `runs-on: macos-15`, `xcodebuild test … -only-testing:ProvenenciaTests CODE_SIGNING_ALLOWED=NO` | The `macos-15` image's default Xcode is 16.x, which reads `objectVersion = 77`; the job prints `xcodebuild -version` |
| Gate | `scripts/check-pr.sh` `xcode_test`, `swift_warnings` | On a non-Darwin host `xcode-test` is reported failed ("needs macOS") |
| Guidance | `.agents/skills/add-swift-test/SKILL.md` § "Xcode project" (lines 50–58: copy four objects, IDs `AA…xx`, group `B5`, phase `B7`); `.agents/skills/add-ui-component/SKILL.md` line 72 (checklist "New .swift registered in the Xcode project (pbxproj)") and line 132 (step 4 "Register new files in `macos/Provenencia.xcodeproj`."); `.cursor/rules/swift-tests.mdc` line 13 ("Wire new files into the ProvenenciaTests target in `project.pbxproj` only."); `macos/App/DesignSystem/README.md` lines 563–568 (fonts: "add it to the Xcode project's Resources phase") and § "Xcode project registration" (lines 605–612); `docs/macos-client-patterns.md` line 37; `README.md` line 13 ("any current Xcode that can target macOS 14 is fine") | All describe the hand-written project |
| Disk-vs-list check | `scripts/check-guidance-names.py` `kit_rows_missing` (from audit A-33), called from `main` | Walks `macos/App/DesignSystem/{Components,Recipes}` and fails on a folder the DesignSystem README has no row for; runs on every host through check-pr's `guidance-names` |
| More guidance | `.agents/skills/add-swift-test/SKILL.md` frontmatter `description` ("…or wiring a .swift file into the ProvenenciaTests Xcode target."); `.agents/skills/add-workspace-place/SKILL.md` line 187 ("for wiring new Swift test files"); `.agents/skills/add-ui-component/SKILL.md` line 130 ("plus optional README/helpers in the same folder"); `docs/design-system-layers.md` line 43 ("# optional: README.md, private helpers, previews-only fixtures") | Describe wiring that goes away, or invite a non-Swift file with no word about its exception |
| Tests | every suite in `macos/ProvenenciaTests` (117 files), including `PVFontTests.swift` | Run by `xcode-test`; no test reads the project file |

**Precedent.** None in this repo. The target shape is the one Xcode 16 writes for a new project or for **Convert to Folder** on a group: a `PBXFileSystemSynchronizedRootGroup` per folder, listed in the target's `fileSystemSynchronizedGroups`, with a `PBXFileSystemSynchronizedBuildFileExceptionSet` for files kept out of a target. §6 **macOS** gives the exact objects.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The conversion is a one-off edit of `project.pbxproj` made with a throwaway script in the implementer's scratch directory (not committed), then checked by Xcode (STEP-2, STEP-3). | The diff is deterministic and reviewable, keeps every surviving object ID, and touches nothing D-5 does not name. | Xcode's **Convert to Folder** on each group — it rewrites the file in Xcode's own order and can touch unrelated objects, so the diff cannot be checked against this plan; it is also unavailable to an agent driving a terminal. |
| D-2 | Five synchronized root groups, each **reusing its group's ID**: `AA0000000000000000000045` (`App`), `AA0000000000000000000046` (`Platform`), `AA0000000000000000000047` (`Features`), `550D63589090B3E5D65DBED2` (`DesignSystem`) for target `AA0000000000000000000050`; `AA00000000000000000000B5` (`ProvenenciaTests`) for target `AA00000000000000000000B6`. The container group `AA0000000000000000000041` (`App`) and `Resources` `AA0000000000000000000049` stay classic `PBXGroup`s, unchanged. | Reused IDs keep the parent groups' children lists as they are. `Resources` holds the `Fonts` folder reference, which `PVTypography` loads as the bundle subfolder `Fonts`; a synchronized folder copies the resources inside it into the bundle's `Resources/` root, which would flatten `Fonts/` and lose every bundled font. `Resources` has six items and changes rarely. | One synchronized root for all of `macos/App` — flattens `Fonts`; making it work needs a font-loader change (Out). Synchronizing `Resources` with `Fonts` excluded — the fonts would no longer be bundled. |
| D-3 | Two `PBXFileSystemSynchronizedBuildFileExceptionSet`s for target `AA0000000000000000000050` keep the docs out of the app: on `DesignSystem`, `membershipExceptions = (README.md, Recipes/Marks/MARKS.md, )`; on `Features`, `membershipExceptions = (GraphCanvas/README.md, )`. New IDs `AA15150000000000000000E1` (DesignSystem) and `AA15150000000000000000E2` (Features); `rg AA1515 macos/Provenencia.xcodeproj` has no hit at the base. | A synchronized folder makes every non-Swift file a bundle resource: the two `README.md` files would collide in `Contents/Resources` and fail the build, and `MARKS.md` would ship in the app. The files stay where skills and docs link them. | Moving the docs out of `macos/App` — skills, `check-guidance-names.py` and the DesignSystem README rows cite these paths. |
| D-4 | `objectVersion = 60` → `77`. In the PBXProject object, delete `compatibilityVersion = "Xcode 15.0";` and add `preferredProjectObjectVersion = 77;` after `packageReferences`. `README.md` names Xcode 16 as the minimum. | Xcode reads `PBXFileSystemSynchronizedRootGroup` only in format 77 (Xcode 16); this is the pair Xcode 16 writes. CI's `macos-15` runner and the README's "last built with Xcode 26" are both 16 or later. | Keeping format 60 — the new objects are not readable in it. |
| D-5 | Delete: every `PBXBuildFile` whose `fileRef` is a `.swift` file reference (425); every `PBXFileReference` with `lastKnownFileType = sourcecode.swift` (425); file reference `6AAC04817D6F4CCC9CDC9B17` (`DesignSystem/README.md`); every `PBXGroup` that is a descendant of one of the five D-2 groups. The two Sources phases keep their objects with an empty `files` list in the file's multi-line form (`files = (` on one line, `);` on the next, as phase `AA00000000000000000000B9` has). The five D-2 objects leave the `PBXGroup` section. Everything else stays byte-for-byte: build settings, `MARKETING_VERSION`, the Go-core phase, Frameworks, Embed Libraries, Resources, packages, `Core`, `Products`, the scheme. | The synchronized folders supply exactly these files; a leftover explicit entry next to a synchronized folder lists the file twice. | Keeping the old group objects as navigator decoration — they would be dead objects agents keep editing. |
| D-6 | Proof of sameness is equality, not "it builds": the sorted `.SwiftFileList` of each target and the sorted listing of the built app's `Contents/Resources` are identical before and after (STEP-1, STEP-3). | A dropped file that nothing references, or a stray bundled resource, still builds; equality catches both. | Build and test passing alone. |
| D-7 | `scripts/check-guidance-names.py` gains `synced_folders_unlisted()` *(new)*, called from `main` after `kit_rows_missing()`. It keeps **no list of its own**: it reads `macos/Provenencia.xcodeproj/project.pbxproj` with regular expressions over whole objects (`re.S`; an object spans lines) and derives (1) the roots: each `PBXFileSystemSynchronizedRootGroup`, whose folder is `macos/` + the `path` of the `PBXGroup` whose `children` hold its ID (when that group has a `path`) + its own `path` (after STEP-2: `macos/App/App`, `macos/App/Platform`, `macos/App/Features`, `macos/App/DesignSystem`, `macos/ProvenenciaTests`); (2) each root's exceptions: the `membershipExceptions` entries of the sets its `exceptions = (…)` names, as paths relative to the root, quotes stripped; (3) the names the project gives `macos/App`: the `path` of each child (group or root) of the main group's child whose `path` is `App`. It walks each root recursively with `os.walk`, skipping any file or directory whose name starts with `.`, and reports one line each, sorted, in these exact forms: (a) `<file>: not Swift and not in a membership exception of the <root> synchronized folder` for a file not ending in `.swift` and not listed; (b) `<root>/<entry>: listed in a membership exception but not on disk`; (c) `macos/App/<name>: not a synchronized folder or group in the Xcode project` for a file or directory directly under `macos/App` whose name is not in (3); (d) `<root>: synchronized folder missing on disk`. Paths are repo-relative. `main`'s hint line gains "; list a non-Swift file in its synchronized folder's membership exceptions in macos/Provenencia.xcodeproj/project.pbxproj". The module docstring gains a paragraph saying what the check reads and reports, and the `guidance-names` comment in `scripts/check-pr.sh` gains ", and every non-Swift file in a synchronized Xcode folder is a membership exception". No check for `sourcecode.swift` lines. | The exception lists are the one hand-kept list left in the project file (review M-1): a missed exception ships a file in the app with no error, and a second `README.md` fails only the Mac build; this script runs on every host. Deriving the roots from the project keeps one declared list, so a root added in Xcode needs no script edit (review M-5), as `kit_rows_missing` reads its declared list (the README) rather than a copy. (c) catches a new top-level folder or loose file under `macos/App` that nothing builds. | No check — an agent without Xcode gets no signal. A `SYNCED_ROOTS` constant (round 1 draft) — a second copy of the project's list that drifts silently. A new script — this one already owns disk-vs-list checks and runs in check-pr. A `sourcecode.swift` check — the skills remove the habit and a reappearing entry shows in the diff. |
| D-8 | Guidance text, exactly: in `add-swift-test`, lines 52–58 (the paragraph under `## Xcode project` through "Do not add a new native target or scheme entry; those already exist.") are replaced entirely; the heading and the "Bundle id stays" line stay. The new paragraph: "`macos/ProvenenciaTests` is a synchronized folder: a `.swift` file saved there is in the **ProvenenciaTests** target. Do not edit `project.pbxproj`, and never add a test file under `macos/App`." `add-ui-component` checklist line 72 becomes "- [ ] File saved under `macos/App/DesignSystem/` or `macos/App/Features/` (synchronized folders; no `project.pbxproj` edit)"; step 4 becomes "4. Save the file in its folder; the synchronized folder adds it to the build. A non-Swift file (a README) in the folder needs a membership exception in `project.pbxproj` (see the DesignSystem README § Xcode project registration)." `swift-tests.mdc` line 13's last sentence becomes "`macos/ProvenenciaTests` is a synchronized folder: new files need no `project.pbxproj` edit." The DesignSystem README section keeps its heading `## Xcode project registration`, and the fonts paragraph's "(see \"Xcode project registration\" below)" stays; the paragraph's "add it to the Xcode project's Resources phase" becomes "it is bundled through the `Fonts` folder reference". The section body (lines 607–612) is replaced by these two paragraphs: "`App/App`, `App/Platform`, `App/Features`, `App/DesignSystem` and `ProvenenciaTests` are Xcode synchronized folders: a Swift file saved in one is built into its target with no edit to `project.pbxproj`. Any other file in them is copied into the app bundle, so a README or other note is listed in its folder's membership exceptions (`PBXFileSystemSynchronizedBuildFileExceptionSet`); `scripts/check-guidance-names.py` fails on one that is not." and "`Resources/` is a classic group. A new `.ttf` dropped into `Resources/Fonts/` is bundled through the `Fonts` folder reference with no project edit." `add-swift-test`'s frontmatter `description` phrase "or wiring a .swift file into the ProvenenciaTests Xcode target" becomes "or adding a test file under macos/ProvenenciaTests". In `add-workspace-place` line 187, "for wiring new Swift test files" becomes "for new Swift test files". `add-ui-component` line 130's "plus optional README/helpers in the same folder" becomes "plus optional helpers in the same folder; a README there is listed in the folder's membership exceptions, which `check-guidance-names.py` enforces". `docs/design-system-layers.md` line 43 becomes "# optional: private helpers, previews-only fixtures; a README.md is listed in DesignSystem's membership exceptions". `docs/macos-client-patterns.md` after line 37 gains: "`App/App`, `App/Platform`, `App/Features`, `App/DesignSystem` and `ProvenenciaTests` are synchronized folders: a file on disk there is in the build. `App/Resources` is a classic group. A new top-level folder under `macos/App` is not built until it is added as a synchronized folder; `scripts/check-guidance-names.py` fails until it is." | The skills and rule are what agents follow; leaving any hand-entry step makes the next agent write entries the project no longer has. | Pointing at Xcode docs only — agents work without Xcode open. |

**Rules this plan bends.** None. `.cursor/rules/swift-tests.mdc` changes because its pbxproj sentence becomes false, which is this PR's scope (map scope entry), not a bend.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| Five `PBXFileSystemSynchronizedRootGroup` objects | project objects | The five `PBXGroup`s they replace (same IDs) | They are the replacement; D-2 |
| Two `PBXFileSystemSynchronizedBuildFileExceptionSet` objects `AA15150000000000000000E1`, `AA15150000000000000000E2` *(new)* | project objects | None (no exception mechanism in format 60) | D-3 |
| `fileSystemSynchronizedGroups` on both targets *(new key)* | project key | `packageProductDependencies` on the app target | Different list |
| `synced_folders_unlisted()` *(new)* | func | `scripts/check-guidance-names.py` `kit_rows_missing` | Extends the same script with a sibling check (D-7); its declared list is the project file, not the README |

Searched: `rg -n fileSystemSynchronized macos` (none), `rg -n 'pbxproj|PBXFileReference|PBXBuildFile' .agents .cursor docs/*.md macos/App scripts` (the guidance rows in §4 and `check-guidance-names.py`'s extension list, which needs no change).

**Builds on.** Xcode 16's synchronized folder format; the existing classic `Resources` group and `Fonts` folder reference.
**Deletes.** 425 Swift `PBXBuildFile`s, 425 Swift `PBXFileReference`s, file reference `6AAC04817D6F4CCC9CDC9B17`, every descendant `PBXGroup` of the five roots, the Sources phases' file lines, `compatibilityVersion`; the hand-entry instructions in `add-swift-test`, `add-ui-component`, `swift-tests.mdc` and the DesignSystem README.
**Next adds.** A README in a new kit folder is one `membershipExceptions` line, which the check names. Every later PR that adds a Swift file (S14-14, S14-16, S14-17, S14-20, S14-21, S14-22, S14-27, and Spike 13's open PRs) saves the file and edits no project file.

### Reach

None — no new path into app code and no new runtime state; the compiled file set is unchanged (AC-2.1).

### Scale

None — no write, list or loop in the app changes.

### Edges

- **Assumptions.** Synchronized folders build what is on disk, so the conversion is safe only if the project's file set equals the disk's. Measured at the base (§4 Groups and Swift entries rows): equal, no file in both targets, no shared basename, every group path equals its folder. AC-2.1 holds it.
- **Neighbors.** `Platform/Generated/engine.pb.swift` (written by `scripts/generate-proto.sh`) is in the `Platform` folder and stays built. The bridging header and entitlements stay in classic `Resources` and stay unbuilt and referenced by build settings. The three `.md` files are the only non-Swift files inside the five roots (D-3); AC-2.2 holds them out of the bundle.
- **Removed inputs.** None — no file leaves the build.
- **Real data.** The full file set, measured as above: 308 app and 117 test Swift files.
- **Untrusted size.** None — no data the app did not write is read.

### Schema

None — no catalog change.

### Go core

None — no Go change.

### FFI

None — no FFI change.

### macOS

Project objects after STEP-2 (IDs from D-2 and D-3; sections in Xcode's alphabetical order, between `PBXFileReference` and `PBXFrameworksBuildPhase`):

```text
/* Begin PBXFileSystemSynchronizedBuildFileExceptionSet section */
		AA15150000000000000000E1 /* Exceptions for "DesignSystem" folder in "Provenencia" target */ = {
			isa = PBXFileSystemSynchronizedBuildFileExceptionSet;
			membershipExceptions = (
				README.md,
				Recipes/Marks/MARKS.md,
			);
			target = AA0000000000000000000050 /* Provenencia */;
		};
		AA15150000000000000000E2 /* Exceptions for "Features" folder in "Provenencia" target */ = {
			isa = PBXFileSystemSynchronizedBuildFileExceptionSet;
			membershipExceptions = (
				GraphCanvas/README.md,
			);
			target = AA0000000000000000000050 /* Provenencia */;
		};
/* End PBXFileSystemSynchronizedBuildFileExceptionSet section */

/* Begin PBXFileSystemSynchronizedRootGroup section */
		AA0000000000000000000045 /* App */ = {isa = PBXFileSystemSynchronizedRootGroup; path = App; sourceTree = "<group>"; };
		AA0000000000000000000046 /* Platform */ = {isa = PBXFileSystemSynchronizedRootGroup; path = Platform; sourceTree = "<group>"; };
		AA0000000000000000000047 /* Features */ = {isa = PBXFileSystemSynchronizedRootGroup; exceptions = (AA15150000000000000000E2 /* Exceptions for "Features" folder in "Provenencia" target */, ); path = Features; sourceTree = "<group>"; };
		550D63589090B3E5D65DBED2 /* DesignSystem */ = {isa = PBXFileSystemSynchronizedRootGroup; exceptions = (AA15150000000000000000E1 /* Exceptions for "DesignSystem" folder in "Provenencia" target */, ); path = DesignSystem; sourceTree = "<group>"; };
		AA00000000000000000000B5 /* ProvenenciaTests */ = {isa = PBXFileSystemSynchronizedRootGroup; path = ProvenenciaTests; sourceTree = "<group>"; };
/* End PBXFileSystemSynchronizedRootGroup section */
```

In each `PBXNativeTarget`, after `dependencies = ( … );`:

```text
			fileSystemSynchronizedGroups = (
				AA0000000000000000000045 /* App */,
				AA0000000000000000000046 /* Platform */,
				AA0000000000000000000047 /* Features */,
				550D63589090B3E5D65DBED2 /* DesignSystem */,
			);
```

and on `AA00000000000000000000B6`: `fileSystemSynchronizedGroups = ( AA00000000000000000000B5 /* ProvenenciaTests */, );`.

No view, component or state changes.

### Strings

None — no user-facing copy.

## 7. Steps

All steps run on a Mac with Xcode 16 or later. Let `S` be a scratch directory outside the checkout (`mktemp -d`).

**STEP-1 — Baseline on the planning base**
Files: none changed.
Do: on a clean checkout of the branch before any edit (equal to `main`), run `xcodebuild build-for-testing -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -derivedDataPath "$S/dd-before" CODE_SIGNING_ALLOWED=NO`. Then write the three lists (the build is for the host architecture only, so each `find` matches one file; more than one match is a Stop condition):

```bash
R="$(git rev-parse --show-toplevel)/macos"
sort "$(find "$S/dd-before" -name 'Provenencia.SwiftFileList')" | grep -F "$R/App/" > "$S/app-before.txt"
sort "$(find "$S/dd-before" -name 'ProvenenciaTests.SwiftFileList')" | grep -F "$R/ProvenenciaTests/" > "$S/tests-before.txt"
(cd "$S/dd-before/Build/Products/Debug/Provenencia.app/Contents/Resources" && find . | sort) > "$S/res-before.txt"
```
Check: `wc -l` gives 308 and 117 lines for the two file lists; `res-before.txt` contains `./Fonts/Newsreader-400.ttf` and no `README.md`.

**STEP-2 — Convert the project file**
Files: `macos/Provenencia.xcodeproj/project.pbxproj`.
Do: write a throwaway script in `$S` (not committed) that applies D-4 and D-5 and inserts the objects in §6 **macOS** (D-2, D-3). Run it on the project file. Group deletion is by walking `children` from the five D-2 roots: those 5 roots and their 72 descendant groups (77 in all) leave the `PBXGroup` section; every other `PBXGroup` is kept unchanged. Indentation is tabs, as in the rest of the file.
Check: `plutil -lint macos/Provenencia.xcodeproj/project.pbxproj` prints OK; `xcodebuild -list -project macos/Provenencia.xcodeproj` lists targets `Provenencia` and `ProvenenciaTests` and scheme `Provenencia`; `grep -c 'sourcecode.swift'` prints 0 (AC-1.3); `grep -c 'isa = PBXFileSystemSynchronizedRootGroup'` prints 5; `grep -c 'isa = PBXFileSystemSynchronizedBuildFileExceptionSet'` prints 2; `grep -c 'MARKETING_VERSION = 0.2.5'` prints 4 (both targets, Debug and Release); `grep -c 'isa = PBXGroup;'` prints 5 (82 at the base; the survivors are main `AA0000000000000000000040`, `App` `…41`, `Products` `…42`, `Core` `…43`, `Resources` `…49`); `git diff --stat` shows only the project file.

**STEP-3 — Prove the same build**
Files: none changed.
Do: repeat STEP-1's build and the three commands with `dd-after` and `-after.txt` in place of `dd-before` and `-before.txt`. The SwiftFileList paths are absolute; both builds run from the same checkout path.
Check: `diff "$S/app-before.txt" "$S/app-after.txt"`, the tests pair and the resources pair each print nothing (AC-2.1, AC-2.2).

**STEP-4 — Probe a new file**
Files: `macos/App/Features/Sources/SyncProbe.swift` *(new)* (temporary), `macos/ProvenenciaTests/SyncProbeTests.swift` *(new)* (temporary), contents per AC-1.1 and AC-1.2.
Do: add both, with no project edit; run `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -derivedDataPath "$S/dd-probe" -only-testing:ProvenenciaTests/SyncProbeTests CODE_SIGNING_ALLOWED=NO`. Then delete both files.
Check: `grep -c 'SyncProbe.swift' "$(find "$S/dd-probe" -name 'Provenencia.SwiftFileList')"` prints 1 before the files are deleted (AC-1.1); the test passes (AC-1.2); `git status --porcelain` afterwards lists only `project.pbxproj`.

**STEP-5 — Full gate**
Files: none changed.
Do: `scripts/check-pr.sh`.
Check: `check-pr: PASS` (AC-2.3). Commit the project file: "Build every Swift file in the app and test folders without listing it in the Xcode project."

**STEP-6 — Check the synchronized folders**
Files: `scripts/check-guidance-names.py`, `scripts/check-pr.sh` (the `guidance-names` comment only).
Do: add `synced_folders_unlisted()` per D-7, after `kit_rows_missing`, with the standard library only; update the docstring, hint line and check-pr comment per D-7.
Check: `python3 scripts/check-guidance-names.py` prints `check-guidance-names: OK`; AC-3.3's two plants each make it exit 1 naming the file or folder; after each is undone it prints OK, and after the last `git status --porcelain` is empty except for this step's two files. This step runs on any host. Commit: "Fail the guidance check when a synchronized folder holds a file the app would bundle by accident."

**STEP-7 — Skills and rule**
Before editing, run the AC-3.1 sweep command and confirm 12 hits.
Files: `.agents/skills/add-swift-test/SKILL.md` (§ Xcode project and frontmatter), `.agents/skills/add-ui-component/SKILL.md` (lines 72, 130, 132), `.agents/skills/add-workspace-place/SKILL.md` (line 187), `.cursor/rules/swift-tests.mdc`. Skill: `add-swift-test`, `add-ui-component`.
Do: replace the text named in §4 **Guidance** and **More guidance** (skills and rule only) with D-8's text.
Check: AC-3.1's `rg` has no hit in `.agents` and `.cursor`; `python3 scripts/check-guidance-names.py` passes.

**STEP-8 — Docs**
Files: `docs/design-system-layers.md` (line 43, per D-8), `README.md` (line 13, keeping its `1. ` marker: "Install Xcode 16 or later (this repo was last built with Xcode 26) and Go 1.27+."), `docs/macos-client-patterns.md`, `macos/App/DesignSystem/README.md` (fonts paragraph and § "Xcode project registration" per D-8), `docs/deployment-plan/spike-14/completed.md`, `docs/deployment-plan/spike-14/deployment-plan.md` (map row), this plan (status and Landed note).
Do: per D-8 and the Definition of done.
Check: AC-3.1's full `rg` has no hit; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-15-xcode-synchronized-folders.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS`. Commit: "Tell contributors that new Swift files need no Xcode project edit."

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

No Swift or Go test is added: no source changes, and the existing suite is the regression check.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Build | STEP-1 / STEP-3 SwiftFileList diff | before vs after → no difference (308 / 117) | AC-2.1, D-5, D-6 |
| T-2 | Build | STEP-1 / STEP-3 `Contents/Resources` diff | before vs after → no difference; no `.md` | AC-2.2, D-2, D-3 |
| T-3 | Build | STEP-4 probe | new app and test file, no project edit → compiled, test passes | AC-1.1, AC-1.2 |
| T-4 | Lint | STEP-2 checks | `plutil -lint` OK; counts `sourcecode.swift` 0, roots 5, exception sets 2, `MARKETING_VERSION` 4, `PBXGroup` 5 | AC-1.3, D-4 |
| T-5 | Swift | all of `macos/ProvenenciaTests` via `xcode-test` (including `PVFontTests`) | → pass, no Swift warning | AC-2.3 |
| T-6 | Guidance | `scripts/check-guidance-names.py` and AC-3.1's `rg` | → pass, no hit | AC-3.1, AC-3.2, D-8 |
| T-7 | Script | STEP-6, AC-3.3's three plants *(new)* (temporary) | each → exit 1 with the exact lines; undone → OK | AC-3.3, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs, and CI's `macOS tests / test` job on `macos-15` (its first run on the new format).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none added
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI change
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] CI `macOS tests / test` is green on the PR, with its `xcodebuild -version` line quoted in the PR description (the job runs only on a non-draft PR: this box is checked after the PR is marked ready, per ship-pr)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped (audit A-34)
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: docs/macos-client-patterns.md (D-8), README.md (Xcode 16)
- [ ] PR description: what shipped, AC evidence (the three empty diffs, the probe), deviations, manual check
```

## 10. Manual check

1. Open `macos/Provenencia.xcodeproj` in Xcode 16 or later. The navigator shows `App`, `Platform`, `Features`, `DesignSystem` and `ProvenenciaTests` as folders (blue folder icon), and `Resources` as a group with its six items. Xcode offers no project upgrade.
2. ⌘R. The onboarding or workspace window draws in Newsreader and Spectral, not the system font (the fonts loaded from `Fonts/`).
3. In Finder, add an empty `.swift` file under `macos/App/Features/Sources/`; Xcode's navigator shows it without a project edit and `git status` lists no project change. Delete it.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A branch opened before this lands adds hand-written Swift entries and rebases on it | The rebase conflicts in `project.pbxproj`; the branch drops its entries. The skills changed here say so |
| CI's Xcode is older than 16 | The workflow prints `xcodebuild -version`; Stop condition below |
| Opening the project in a newer Xcode rewrites it (key order, `minimizedProjectReferenceProxies`) | Not committed in this PR; a later rewrite is harmless |

**Stop and report if:**

- No Mac with Xcode 16 or later is available: stop before STEP-1.
- A STEP-1 `find` matches more than one SwiftFileList, or STEP-1's lists do not have 308 and 117 lines, or a group under the five roots has a `path` that is not its folder name (the base moved; §4 is stale).
- `xcodebuild -list` or the build rejects the converted file, or Xcode reports the project needs an upgrade beyond format 77.
- Any STEP-3 diff is not empty. Report the differing lines; do not edit the project to hide them.
- A non-Swift file other than the three in D-3 is found under the five roots at your base.
- `synced_folders_unlisted()` reports anything on the converted tree before the plants (the exceptions in §6 and the disk disagree), or finds a number of roots other than 5.
- CI's `test` job (once the PR is out of draft) fails reading the project, or prints an Xcode older than 16.

## 12. Open questions

None.

## 13. PR

**Title:** S14-15 — Xcode synchronized folders
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence (STEP-3's three empty diffs, the STEP-4 probe, check-pr PASS, CI's Xcode version); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-15-xcode-synchronized-folders.md`.
