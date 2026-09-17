# Provenencia — Catalog refs

## Status

Implementation notes for short user-facing identifiers (`ref`). Domain philosophy lives in [`data-model-source-interpretation-conclusion.md`](data-model-source-interpretation-conclusion.md) §2; this file is the **how** for Go, schema, and UI.

---

## 1. Two identities

| Kind | Role | Example |
| --- | --- | --- |
| Machine id | Durable primary key (UUIDv7 `BLOB`) | `019c…` |
| `ref` | Short human handle for UI, search, conflict display | `SRC-F4N2P` |

Never use `ref` as a foreign key. Never recycle a `ref` after delete/merge; old refs must keep resolving to the surviving row when merges exist.

---

## 2. Wire format

```text
{PREFIX}-{TOKEN}
```

- **PREFIX** — exactly three ASCII letters, uppercased (`USR`, `SRC`, `PER`, …).
- **TOKEN** — exactly five characters from Crockford base32 **without** `I`, `L`, `O`, `U` (alphabet `0123456789ABCDEFGHJKMNPQRSTVWXYZ`).

Examples: `USR-F4N2P`, `SRC-3K9M2`, `OBS-2F8Q1`.

### Candidate Nodes (Interpretation)

There is **one ref format**. Interpretation Nodes are distinguished by their own three-letter prefix, not by a special shape:

```text
canonical entity   {PREFIX}-{TOKEN}    e.g. PER-7KD45   (canonical_entities)
candidate Node     {PREFIX}-{TOKEN}    e.g. CPR-7KD45   (nodes)
```

A Node Type therefore carries **two** prefixes — `node_types.ref_prefix` for the Conclusion handle and `node_types.candidate_ref_prefix` for the Interpretation Node — because `canonical_entities` and `nodes` share the same `node_types` vocabulary. Seeded pairs are in [`seeded-vocabulary.md`](seeded-vocabulary.md) §3.1.

Candidate prefixes conventionally start with `C` (`CPR`, `CEV`, `CPL`, …), but that is **not enforced**. Both columns validate identically through `ref.ValidatePrefix`, and nothing in the format tells you which layer a prefix belongs to — that is a registry lookup. The consequence to keep in mind: all Node Type prefixes, canonical and candidate, share one three-letter namespace and must be unique across **both** columns. That is an application rule; no single SQL `UNIQUE` expresses it.

An earlier design used an infix marker (`PER-C-7KD45`) so one prefix could serve both layers. It was dropped because the special shape leaked into everything that touches refs — a second mint function, a second validator, partial-match handling in search — for a distinction that a separate prefix encodes for free.

---

## 3. Reserved and seeded prefixes

**Catalog / contributor (fixed):**

| Prefix | Entity |
| --- | --- |
| `USR` | contributors (`users`, install identity) |
| `SRC` | sources |
| `ART` | artifacts |
| `CIT` | citations |
| `OBS` | observations |

**Node Type prefixes** come from `node_types`, two per row (see [`seeded-vocabulary.md`](seeded-vocabulary.md) §3.1): `ref_prefix` for canonical entities (`PER`, `EVT`, `PLC`, `REL`, `PTN`, `LOC`, `SRN`, …) and `candidate_ref_prefix` for Nodes (`CPR`, `CEV`, `CPL`, `CRL`, `CPA`, `CLO`, `CSR`, …). None of them may collide with the reserved catalog set above, or with each other across either column.

**Proposed (not shipped):** `NAR` for Narrative compositions ([`narrative-layer-data-model.md`](narrative-layer-data-model.md)). Promote to the reserved catalog table above when the layer is roadmapped.

Uniqueness is **across all ref-bearing tables in one project** (application rule). Prefer a project-wide unique index strategy when multiple tables carry `ref`; until then, enforce uniqueness in the insert path (retry mint on conflict).

---

## 4. Go package: `core/ref`

Authoritative mint/validate code: [`core/ref`](../core/ref/ref.go).

```go
r, err := ref.Mint(ref.PrefixSource)
if err != nil { … }
if err := ref.Validate(r); err != nil { … }

// Interpretation Node — an ordinary ref off the candidate prefix
n, err := ref.Mint(nodeType.CandidateRefPrefix) // CPR-7KD45
if err := ref.Validate(n); err != nil { … }
```

| API | Behavior |
| --- | --- |
| `Mint(prefix)` | Normalizes prefix to `A-Z{3}`, appends `-` + random 5-char token |
| `Valid` / `Validate` | The one ref form (`PREFIX-TOKEN`) |
| `ValidatePrefix(prefix)` | Normalizes a Node Type prefix — either column — and rejects the reserved catalog prefixes (`ref.reserved_prefix`) |

There is no candidate-specific mint or validator, and no layer-specific ref grammar. A Node ref and a canonical entity ref are the same shape; only the prefix differs, so every existing consumer of `Mint` / `Valid` already handles both.

`Mint` still accepts reserved prefixes, because `ref.Mint(ref.PrefixSource)` is how catalog rows are minted. Only `ValidatePrefix`, which guards Node Type vocabulary, refuses them.

`ValidatePrefix` does **not** check cross-column uniqueness or enforce the leading `C` convention. Enforce uniqueness in the Node Type write path against both `ref_prefix` and `candidate_ref_prefix`.

**Do not** invent refs in SQL (`hex(id)`, random literals). Mint in Go on insert (or backfill in an `EnsureRefs`-style helper after migration).

---

## 5. Schema pattern

On a ref-bearing table:

```sql
ref TEXT NOT NULL,
-- UNIQUE(ref) on this table; also unique across ref-bearing tables at the app layer
```

- Assign `ref` in the domain `Upsert` / insert path via `ref.Mint(prefix)`.
- Reject empty / malformed refs (`ref.Validate`) before write.
- Backfill nullable legacy columns in Go after migrate (see `users.EnsureRefs`), not with unstable SQL placeholders.

Install-local identity mirrors the contributor ref in `identity.json` (`ref` field) so the same `USR-…` is reused across projects on one Mac.

---

## 6. UI

- Show `Display name (USR-F4N2P)` (or label + ref) when two rows can share a name.
- Prefer localized format helpers (see onboarding contributor copy) over concatenating in views.
- Do not surface the UUID as the primary human identifier.

---

## 7. Agent skill

When adding `ref` to a new catalog entity, follow [`.cursor/skills/add-catalog-ref/SKILL.md`](../.cursor/skills/add-catalog-ref/SKILL.md).
