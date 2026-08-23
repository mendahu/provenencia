# Provenance product versioning

## Status

Working policy. The product uses **one SemVer tree** for every native client and the Go core. The `.provenance` on-disk format has its own number and must not be confused with the app version.

---

## Product version (SemVer)

Provenance is a single product. macOS, a future Windows client, and the linked Go core all report the **same** `MAJOR.MINOR.PATCH`.

There is no `mac@1.4` vs `windows@1.1` vs `core@3.2` for users or release notes. A platform that has not shipped a milestone yet is simply behind; when it ships, it uses that milestone’s product version (for example both say `0.2.0`).

**Pre-1.0:** `0.y.z` may still change behavior and format aggressively. Still bump SemVer honestly so git tags and About boxes stay meaningful.

### What bumps

| Bump | When |
| --- | --- |
| **MAJOR** | Incompatible change users or other clients cannot ignore (typically a project-format break that old apps cannot open, or a deliberate reset after 1.0). |
| **MINOR** | Backward-compatible features (new screens, new RPCs old clients can ignore, additive schema that old apps still open). |
| **PATCH** | Bugfixes, packaging, and platform-only hotfixes. A Windows-only crash fix that you release is still **Provenance `x.y.z+1`**, not a Windows-private version. |

Do **not** bump the product version for docs-only or comment-only changes.

A Mac-only UI polish that you cut as a Mac build still takes the next product patch/minor. The version names the **release**, not “how much Go vs Swift changed.”

### Build numbers (not SemVer)

Xcode `CURRENT_PROJECT_VERSION` (and later Windows `FileVersion`) may increase every CI build. They can differ by platform. Users should still see the **marketing** SemVer from `VERSION`.

---

## Source of truth

Canonical string: repo-root [`VERSION`](../VERSION) (`MAJOR.MINOR.PATCH`, no `v` prefix).

Keep these in lockstep on every product bump:

| Place | Field |
| --- | --- |
| [`VERSION`](../VERSION) | file contents |
| [`core/version.go`](../core/version.go) | `core.Version` |
| [`macos/Provenance.xcodeproj`](../macos/Provenance.xcodeproj) | `MARKETING_VERSION` (Debug and Release) |
| git | annotated tag `vMAJOR.MINOR.PATCH` when cutting a release |

Later: Windows assembly/file marketing version; About panels; Sparkle / WinGet metadata.

This repo has no `package.json`. Do not add one for versioning.

---

## Project format version

SQLite `user_version` (and migrations) is **not** the product SemVer. A Mac `0.5.0` and a Windows `0.5.0` must open the same folder if format versions match. An old product version may refuse a newer format.

Document format bumps in the data-model / migration notes when those exist. Product MAJOR is the usual user-facing signal that old apps cannot open new files; additive schema can be MINOR.

---

## User-visible

Ship the product version in **Provenance → About** (and the Windows equivalent). Optional later: project-info UI showing format/`user_version`.

---

## Go toolchain vs product version

`go 1.27` in `go.mod` is the **language toolchain**, not Provenance `1.27.0`. Do not conflate them.
