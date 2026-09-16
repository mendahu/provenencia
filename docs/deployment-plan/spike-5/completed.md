# Spike 5 — Completed steps

Finished Spike 5 work kept for history. Spike overview: [`README.md`](README.md). Plan and sequence: [`deployment-plan.md`](deployment-plan.md).

IDs stay stable (`S5-NN`, `S5-DN`). Do not renumber when moving steps here.

## Index

| Step | Kind | One-liner |
| --- | --- | --- |
| [S5-01](#s5-01--pr-candidate-ref-minting) | PR | `MintCandidate` / candidate validation / reserved-prefix guard in `core/ref` |

---

## Steps

### S5-01 — PR: Candidate ref minting

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | — |
| **Deliverables** | Done. [`core/ref/ref.go`](../../../core/ref/ref.go): `MintCandidate(prefix)` emitting `{PREFIX}-C-{TOKEN}`, `ValidCandidate` / `ValidateCandidate`, `ValidatePrefix(prefix)` rejecting reserved catalog prefixes, plus `ErrReservedPrefix` and the `reservedPrefixes` set. New wire code `ref.reserved_prefix` in [`core/apperr/apperr.go`](../../../core/apperr/apperr.go). Both mint paths share the existing `normalizePrefix` / `randomToken` helpers. |
| **Tests** | Done. [`core/ref/ref_test.go`](../../../core/ref/ref_test.go): `TestMintCandidate` (valid / lowercase / short / long / digit / blank prefixes, each asserting the minted ref is a candidate and **not** a catalog ref), `TestValidCandidate` (10 rows including the `PER-C4N2P` near-miss and a wrong marker), `TestMintCandidateUnique`, `TestValidatePrefix` (11 rows covering all five reserved prefixes, lowercase normalization, and `SRN` being allowed). Two rows added to the existing `TestValid` for the reverse direction. |
| **Dogfood** | App unchanged. No schema, no FFI, no Swift — nothing is user-visible. |
| **Out** | `node_types` / `nodes` tables and their Go packages (S5-02…S5-04). Ref uniqueness retry, which belongs to the insert path in S5-04. Omnibar handling of candidate refs — [`core/search/refpath.go`](../../../core/search/refpath.go) still calls `ref.Valid`, which is correct while Nodes are not searchable. **L10n mapping for `ref.reserved_prefix`**, deferred because the code is unreachable from Swift until researchers can define Node Types in the vocabulary browser (Spike 7); map it then. |

**Landed:** `core/ref` can mint both identity forms the layer needs. The decision worth recording is that **`Valid` / `Validate` were not widened** — five call sites depend on the strict `AAA-TTTTT` form ([`users`](../../../core/database/users/users.go), [`sources`](../../../core/database/sources/sources.go), [`artifacts`](../../../core/database/artifacts/artifacts.go), [`identity`](../../../core/identity/identity.go), [`search/refpath`](../../../core/search/refpath.go)), so candidate support is purely additive and no existing caller changed.

The two forms are disjoint by length (9 against 11), so no string satisfies both validators. That is load-bearing rather than incidental: `C` is a legal Crockford token character, so `PER-C4N2P` is a valid catalog ref that must never read as a candidate. Both directions are asserted.

`ValidatePrefix` rejects `USR` / `SRC` / `ART` / `CIT` / `OBS`. It does not need a special case for the literal `C` — prefixes are exactly three ASCII letters, so `C` fails on length. A test row documents that rather than carrying a redundant branch.

Docs updated with the step: [`catalog-refs.md`](../../catalog-refs.md) §4 (the "Not implemented yet" row is now the shipped API) and [`add-catalog-ref`](../../../.cursor/skills/add-catalog-ref/SKILL.md) (usage snippet, reserved-prefix guidance, and its stale test command corrected to include `-tags fts5`).

**Verify:**

```bash
go test ./core/ref/
CGO_ENABLED=1 go test -tags fts5 ./...
```
