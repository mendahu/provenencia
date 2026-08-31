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

Typed Interpretation Nodes use an extra `C` segment so they do not look like concluded entities:

```text
canonical / catalog   {PREFIX}-{TOKEN}      e.g. PER-7KD45
node (candidate)      {PREFIX}-C-{TOKEN}    e.g. PER-C-7KD45
```

`C` is a fixed application marker, **not** a `ref_prefix`. Do not register `C` as a three-letter prefix.

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

**Node / canonical type prefixes** come from `node_types.ref_prefix` (see [`seeded-vocabulary.md`](seeded-vocabulary.md)): `PER`, `EVT`, `PLC`, `REL`, `PTN`, `LOC`, `SRN`, …. Those must not collide with the reserved catalog set above.

Uniqueness is **across all ref-bearing tables in one project** (application rule). Prefer a project-wide unique index strategy when multiple tables carry `ref`; until then, enforce uniqueness in the insert path (retry mint on conflict).

---

## 4. Go package: `core/ref`

Authoritative mint/validate code: [`core/ref`](../core/ref/ref.go).

```go
r, err := ref.Mint(ref.PrefixSource)
if err != nil { … }
if err := ref.Validate(r); err != nil { … }
```

| API | Behavior |
| --- | --- |
| `Mint(prefix)` | Normalizes prefix to `A-Z{3}`, appends `-` + random 5-char token |
| `Valid` / `Validate` | Catalog form only (`PREFIX-TOKEN`) today |
| Candidate mint | **Not implemented yet** — add `MintCandidate` / validate when Node tables land; do not hand-roll strings |

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
