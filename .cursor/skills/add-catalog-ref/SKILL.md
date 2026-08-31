---
name: add-catalog-ref
description: >-
  Adds or wires Provenencia short human refs (USR-F4N2P, SRC-…, PER-C-…) using
  core/ref. Use when adding a ref column, minting identifiers for users/sources/
  artifacts/citations/observations/nodes, catalog UNIQUE refs, identity.json ref,
  or when the user mentions ref, USR-, SRC-, PREFIX-TOKEN, or Crockford tokens.
---

# Add a catalog `ref`

Short user-facing ids are **`{PREFIX}-{TOKEN}`** (and **`{PREFIX}-C-{TOKEN}`** for Interpretation Nodes). Machine identity stays UUIDv7. Authoritative product notes: [`docs/catalog-refs.md`](../../docs/catalog-refs.md). Philosophy: [`docs/data-model-source-interpretation-conclusion.md`](../../docs/data-model-source-interpretation-conclusion.md) §2.

## Use `core/ref`

Package: [`core/ref`](../../core/ref/ref.go).

```go
import "github.com/mendahu/provenencia/core/ref"

r, err := ref.Mint(ref.PrefixSource) // or "SRC"
if err != nil {
	return err
}
if err := ref.Validate(r); err != nil {
	return err
}
```

| Do | Do not |
| --- | --- |
| `ref.Mint(prefix)` on insert | Hand-build strings / `fmt.Sprintf` tokens |
| `ref.Validate` before Upsert | Mint in SQL migrations |
| Retry mint on UNIQUE conflict | Recycle deleted refs |
| Keep UUID as PK / FK | Use `ref` as a foreign key |

Token = 5 chars, Crockford alphabet without `I L O U`. Prefix = exactly three ASCII letters, uppercased.

## Prefixes

**Reserved catalog prefixes** (do not reuse for Node types):

| Constant (preferred) | Prefix | Entity |
| --- | --- | --- |
| `ref.PrefixUser` | `USR` | `users` / install identity |
| `ref.PrefixSource` | `SRC` | sources |
| `ref.PrefixArtifact` | `ART` | artifacts |
| `ref.PrefixCitation` | `CIT` | citations |
| `ref.PrefixObservation` | `OBS` | observations |

Node / canonical prefixes (`PER`, `EVT`, …) come from `node_types.ref_prefix`. Candidate Nodes: `{PREFIX}-C-{TOKEN}` — **extend `core/ref`** with `MintCandidate` when first needed; do not concatenate `-C-` ad hoc without shared validate.

Uniqueness: unique **within the project across all ref-bearing tables** (app rule). Table `UNIQUE(ref)` is necessary but not always sufficient once multiple tables exist—check/retry on insert.

## Checklist for a new ref-bearing table

1. **Migration** — follow [add-catalog-migration](../add-catalog-migration/SKILL.md): `ref TEXT NOT NULL` (+ unique). Nullable + Go backfill only when migrating existing rows.
2. **Query package** — follow [add-catalog-query](../add-catalog-query/SKILL.md): Upsert takes/validates `ref`; mint in the use-case or package helper with the correct prefix.
3. **Backfill** — if opening old catalogs, `EnsureRefs`-style mint for NULL/empty (see `users.EnsureRefs`).
4. **FFI / Swift** — expose `ref` on list/get protos; UI shows `Name (REF)` when names can collide ([add-localized-string](../add-localized-string/SKILL.md)).
5. **Docs** — if adding a **new** prefix, update [`docs/catalog-refs.md`](../../docs/catalog-refs.md) and §2 of the data-model philosophy doc.
6. **Tests** — table-driven: mint shape (`ref.Valid`), upsert rejects bad ref, create/open round-trip preserves ref.

```bash
CGO_ENABLED=1 go test ./core/ref/ ./core/database/...
```

## Reference implementation

Contributor path already ships:

- Mint: `identity.Mint` → `ref.Mint(ref.PrefixUser)`
- Catalog: `users.Upsert(…, ref)`, `users.EnsureRefs`
- Install file: `identity.json` field `ref`

Copy that pattern for `SRC` / `ART` / … — same package, different prefix.

## Do not

- Invent a second random-id helper beside `core/ref`
- Put `ref` generation in Swift (mint in Go; Swift only displays)
- Use `C` as a three-letter `ref_prefix`
- Treat folder slugs / filenames as refs
