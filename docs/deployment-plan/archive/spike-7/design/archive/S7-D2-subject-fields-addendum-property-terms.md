# S7-D2 addendum — Property terms (Subject fields deltas)

**Kind:** Claude Design addendum (paste into the **existing** S7-D2 board)  
**Spike:** Provenencia Spike 7  
**Parent brief:** [`S7-D2-subject-fields.md`](S7-D2-subject-fields.md)  
**Implements with:** PR **S7-05** (schema/seed for terms: **S7-01b**)  
**Full rationale:** [`interpretation-graph-ui.md`](../../../../../ideas/interpretation-graph-ui.md) decision **23**; plan step **S7-01b**

> Use this if the S7-D2 board was already started from an earlier brief that assumed five value types and free-text kind/edge Properties (`event_type`, `role`, `relationship_type`). Apply these deltas; do **not** restart the board or redesign the whole IA unless a frame now contradicts these rules.

---

## 1. What changed (one paragraph)

Kind and edge identity are no longer open text. They use a sixth schema value type, **`term`**, backed by `property_terms` rows. **`term`-typed Properties are product/registry-only** — they appear in Subject fields as seeded vocabulary, but researchers **cannot create** a Property with `value_type = term`. Managing which *values* exist for those Properties (birth, father, … plus optional custom rows) is **not** a Subject fields job — that lives in the citation composer term picker (**S7-D4**). Subject fields stays: Properties + bindings to the seven types.

---

## 2. Deltas to apply on the board

### 2.1 Create Property

| Before (if your board assumed this) | Now |
| --- | --- |
| Value type control lists five types | Still **five** choices for researchers: `text`, `integer`, `date`, `name`, `subject` |
| — | Schema also has `term`, but **do not show it** in create/edit Property |

**Frames to revise:** Create Property; any value-type picker; any copy that says “choose how values are stored” with a full schema list.

### 2.2 Seeded Properties that use `term`

These (and similar) will show up in the Property list / bindings as normal seeded rows, with value type **Term** (or product wording you choose):

- `event_type` (on event)
- `role` (on participation)
- `relationship_type` (on relationship)

| Do | Don’t |
| --- | --- |
| Show them like other seeded Properties (origin badge, bindings, locked if registry says so) | Add an “Event types” / “Roles” / “Manage terms” destination or nested browser |
| Make value type read-only / non-editable for seeded `term` Properties | Offer “change value type” or “convert to text” |
| Optional: quiet caption that values are chosen when citing (composer) | Design term CRUD, Add custom, rename, delete here |

**Frames to revise or add (small):** Property detail / inspector for a seeded `term` Property — clarify it’s product vocabulary, not a broken incomplete field.

### 2.3 Out of scope (unchanged emphasis, now sharper)

Still out for Subject fields — call out explicitly on the board if any frame drifted:

- Event types admin
- Roles admin
- Relationship types admin
- Term picker / Add custom / rename / delete user term rows → **S7-D4**

### 2.4 Copy / IA that can stay

Keep your chosen layout (type cards, property-first, matrix, etc.). Property terms do **not** require a new primary navigation axis. This is almost entirely:

1. Create Property value-type list (five, not six)  
2. Honest treatment of a few seeded `term` Properties  
3. Refusal to invent vocabulary browsers for term *values*

---

## 3. Checklist for the designer

- [ ] Create Property value-type control has **exactly five** options — no `term`.
- [ ] At least one frame shows a seeded `term` Property (e.g. `event_type` or `role`) in the list or detail without a “manage values” CTA into a CatalogVocabulary.
- [ ] No frame titled or structured as Event types / Roles / Relationship types.
- [ ] Locked bindings on bridges still work as in the parent brief (unchanged).
- [ ] Notes for S7-05: implementers must refuse `value_type = term` on researcher create even if UI is bypassed.

---

## 4. What you do *not* need to redesign

- Seven-type chrome / asymmetry vs many Properties  
- Search / filter among Properties  
- Binding UX and locked-binding treatment  
- Seeded vs user origin distinction  
- Mounting in the existing Subject fields sidebar destination  

---

## 5. Pointer for implementers (S7-05)

Parent brief + this addendum gate **S7-05**. Schema/API for terms land in **S7-01b** before that PR. Composer term picker is designed in **S7-D4**, not here.
