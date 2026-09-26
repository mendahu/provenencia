# Deployment Plan — Spike 8

Pause-and-refine: make **Source → Evidence graph** data entry cheaper. Stories are independent improvements to that flow, not one schema epic. Authoritative composer/locator: [Spike 7 archive](../archive/spike-7/). Dogfood seed: [`docs/dogfood/ux.md`](../../dogfood/ux.md).

## Status

**Open.** Landings go in [`completed.md`](completed.md). More stories will be added under the same spike.

> **Goal of this spike:** cut the tedium of entering real research without reopening the Interpretation model. Slices land independently. Later stories join this plan as they are scoped.

## Goal (dogfood bar)

Grow this list as stories land. **By spike close**, every checked story below must be true in the app.

1. **Composer rethink** — the citation composer is a **Citation document**: pick Artifact and Citation in-form (no create-time pre-screen); see **all** Observations on that Citation; each row names a **Subject**; Save with zero Observations is allowed; Add property can **reuse** an existing Citation. Graph stays a place. Entry points only pre-select. **This story ships before Auto Transcribe and PDF Find/paste.**
2. **Auto Transcribe** — in the **new** composer, on an **image** Artifact, a control fills **transcription** from Vision OCR of the image (or the region polygon when one is set). The researcher can edit and Save as today. Full-image / oversized jobs warn and can still proceed. No Observation writes. **PDF:** this button stays disabled (text-layer path is bar items 3–4).
3. **PDF Find** — on a PDF in the composer, a Find field on the viewer tool strip jumps to a keyword hit with a highlight. Image-only PDFs (no text layer) fail honestly. Notes: [`pdf-text-find.md`](pdf-text-find.md).
4. **PDF select + paste transcription** — default PDF pointer is text select; scroll/trackpad pans (Preview / PDFKit). **Paste transcription from selection** fills the transcription field. No Vision on PDF pages.
5. **Graph visual enhancements** — conflict + negated row badges; always-on **jump to the Source page**; cited **bridge sentences** prefer endpoint `name` / `event_type` / `toponym`, then working label; **Add property** on bridge cards (extra non-edge rows). More items may join **S8-D3** / **S8-06**.
6. **Source page enhancements** — the Source detail page has an **Open Evidence graph** control for the same Source (disabled with no Artifact). Catalog metadata matches the shipped quick-add / quick-edit (dates are **text**); saved rows can be **deleted**; `url` values are **clickable links**. Credibility sits at the **bottom** of the left overview column; the title edit pencil sits **mid-title** (both already on the S8-D4 board). More items may join **S8-D4** / **S8-07**.
7. **Sources list refresh** — the Sources list shows **subject** and **observation** counts per Source so a worked Evidence graph is obvious next to an empty one. Counts live on their own cache keys (one Source invalidates; the list payload does not). More items may join **S8-D5** / **S8-08**.
8. **Delete paths** — erase is allowed when the target **exists**, extra gates pass, and inbound **resources** are empty. Facets **CASCADE**. Other resources **refuse with a named list** (refs + links). Shared DeleteImpact recipe, then **one screen at a time** (Source page, Source Types, Source Fields, Subject Fields, composer, graph). **Observation row delete already ships in S8-11** and moves onto the recipe with the composer pass. Contract: [`docs/catalog-deletes.md`](../../catalog-deletes.md).
9. **Composer and Connect simplification** — each Observation row saves, reverts, and deletes on its own (delete asks first; nothing is ever deleted by omission). The Citation fields have their own **Save citation**. Leaving the composer with unsaved work always asks. Connect goes straight to the composer, where the connection is **one compact row** (read-only endpoints + role / relationship type) that saves onto a new **or existing** Citation. Endpoints are fixed; a wrong one means discard (or delete the bridge) and connect again. Bridges are named by their cited sentence. A person / event / place can be created from a composer row. The audit log can reconstruct every Interpretation create, edit, and delete. **Ships before items 2, 3, and 4.**

Further bar items: TBD (additional data-entry stories).

## Design track

**Composer, graph, Source-page, and Sources-list chrome is designed in Claude Design before the matching UI PRs.** **S8-D7 first, then S8-D8** — Auto Transcribe and PDF Find/paste design against the S8-D8 form. Briefs: [`design/`](design/).

| Step | Brief | Covers | Gates |
| --- | --- | --- | --- |
| **S8-D7** | Citation composer rethink | Flexible Citation document; compact Artifact vs robust Citation (ref + transcription); inline simple observations; empty Save | **S8-10** |
| **S8-D8** | Composer and Connect simplification | Row-level Observation Save / Revert / Delete; Save citation; unsaved-work guard; one-row connection in the composer (no graph sheet; endpoints fixed); computed bridge names; New person / event / place from a row | **S8-11** |
| **S8-D1** | Auto Transcribe in the composer | Button, progress, replace confirm, large-page warning + proceed, failure copy | **S8-01** |
| **S8-D2** | PDF Find + select + paste | Tool-strip Find; I-beam + scroll-to-pan; paste-from-selection vs Auto Transcribe row | **S8-03**, **S8-04**, **S8-05** |
| **S8-D3** | Evidence graph visual enhancements | Conflict + negated; Source-page jump; richer bridge sentences; Add property on bridges | **S8-06** |
| **S8-D4** | Source page enhancements | Restore shipped metadata + dates as text + delete + clickable `url` (jump already on the board); more page items join this brief | **S8-07** |
| **S8-D5** | Sources list design refresh | Subject + observation counts; more list items join this brief | **S8-08** |
| **S8-D9** | Shared delete confirm / blocked notice | DeleteImpact recipe: confirm if allowed, notice if blocked | **S8-13** |
| **S8-D12** | Source page resource delete | Source + Artifact trash; instances DeleteImpact | **S8-14** |
| **S8-D13** | Source Types delete | DeleteImpact on this page | **S8-15** |
| **S8-D14** | Source Fields delete | DeleteImpact on this page | **S8-16** |
| **S8-D15** | Subject Fields delete | DeleteImpact on this page | **S8-17** |
| **S8-D11** | Citation composer delete chrome | Citation + Observation row; instances DeleteImpact | **S8-18** |
| **S8-D10** | Evidence graph delete chrome | Card trash; instances DeleteImpact | **S8-19** |

## PR sequence

```text
   design                         build
─────────────               ──────────────────────────────────────────

S8-D7  Composer rethink
  │
  └────── gates ──────────▶ S8-10  Flexible composer (identity + multi-subject
                              │     rows + empty Save). FIRST composer PR.
                              │
S8-D8  Composer + Connect     │
       simplification         │
  │                           ▼
  └────── gates ──────────▶ S8-11  Row-level commits, unsaved guard, Connect in
                              │     the composer, audit + write integrity.
                              │     SECOND composer PR (after S8-10).
                              │
                              ├──────▶ S8-D1 / S8-01  (design + build after S8-11)
                              └──────▶ S8-D2 / S8-03…S8-05  (after S8-11)

S8-D1  Auto Transcribe UI
  │
  └────── gates ──────────▶ S8-01  Vision + crop + fill transcription
                              │     (images only; after S8-10)

S8-D2  PDF Find / select / paste
  │
  └────── gates ──────────▶ S8-03  PDFKit live page + I-beam default (scroll pans)
                              │     (after S8-10; MUST precede Find and paste)
                              ├──────▶ S8-04  Find field + highlight + page jump
                              └──────▶ S8-05  Paste transcription from selection
                                              (after S8-03; prefer after S8-01
                                               so one transcription action row)

S8-D3  Graph card visuals
  │
  └────── gates ──────────▶ S8-06  Graph chrome (badges, Source jump, bridge copy + Add property)
                              │     (parallel; no composer / PDF dependency)

S8-D4  Source page enhancements
  │
  └────── gates ──────────▶ S8-07  Open Evidence graph from the Source page
                              │     + metadata dates as text (drop DateValue UI)
                              │     (parallel; pair with S8-06 graph → page)

S8-D5  Sources list refresh
  │
  └────── gates ──────────▶ S8-08  List chrome + per-Source graph-progress counts
                              │     (parallel; do not fold counts into sourcesList)

                            S8-09  Cross-resource FKs → NO ACTION
                              │
                              ▼
                            S8-12  Delete-impact registry + GetDeleteImpact

S8-D9 ─────────────────────▶ S8-13  DeleteImpact recipe (confirm / notice)

Then one screen at a time (brief → PR). Each PR cuts over or adds
that surface’s official Delete and presents DeleteImpact.

S8-D12 ────────────────────▶ S8-14  Source page (Source + Artifact)
S8-D13 ────────────────────▶ S8-15  Source Types
S8-D14 ────────────────────▶ S8-16  Source Fields
S8-D15 ────────────────────▶ S8-17  Subject Fields (properties)
S8-D11 ────────────────────▶ S8-18  Composer (Citation + Observation row)
S8-D10 ────────────────────▶ S8-19  Evidence graph (Subject / bridge)

                            S8-99  Dogfood close / docs
```

**Order notes**

- **S8-D7 / S8-10 first** among composer work. Do not implement **S8-01** or **S8-03…S8-05** on the old form. **S8-D1** / **S8-D2** boards wait until the rethink layout is agreed.
- **S8-D8 / S8-11 next**, before **S8-D1** / **S8-01** and **S8-D2** / **S8-03…S8-05**. S8-11 replaces the composer's save model (footer Save → row commits + Save citation) and removes the graph's Connect sheet. Auto Transcribe and PDF paste land on the S8-11 citation fields section, not the S8-10 footer. **S8-D1** / **S8-D2** boards redraw against **S8-D8** frames.
- **S8-11 ships single Observation delete from the composer.** Later delete PRs reuse `observations.Delete` (and its audit shape). Do not add a second Observation delete.
- **S8-03** remounts PDF from raster → PDFKit. Find and paste cannot ship on `ArtifactMediaViewport` bitmaps.
- **S8-04** and **S8-05** are parallel after **S8-03**.
- **S8-01** does not block **S8-D2**. **S8-05** should follow **S8-01** when both touch the transcription `PVField`.
- **S8-D3** / **S8-06** are independent of OCR and PDF remount. Freeze the **S8-06** bundle on the brief before that PR starts.
- **S8-D4** / **S8-07** are independent of OCR, PDF remount, and **S8-06**. Freeze the **S8-07** bundle on the brief before that PR starts. The two jumps (graph ⇄ page) should use the same location helpers and product name.
- **S8-D5** / **S8-08** are independent of OCR, PDF remount, and the jump pair. Freeze the **S8-08** bundle on the brief before that PR starts. Counts must not ride `sourcesList`.
- **Delete work is linear.** **S8-09** schema → **S8-12** registry → **S8-D9** / **S8-13** shared modal. Then **one screen at a time**: brief, then the PR that cuts over or adds that surface’s `Delete` and presents DeleteImpact. Paste **S8-D9 first**, then the next unpaved screen brief. No parallel engine-cutover PR — the writer moves in the screen PR. Boards do not invent allow/refuse — §S8-09.1–§S8-09.4.

---

## Checklist

- [x] S8-D7 — Design: Citation composer rethink → [`completed.md`](completed.md)
- [x] S8-10 — Flexible citation composer (identity + multi-subject + empty Save) → [`completed.md`](completed.md)
- [x] S8-D8 — Design: Composer and Connect simplification → [`completed.md`](completed.md)
- [x] S8-11 — Composer and Connect simplification + Interpretation write integrity → [`completed.md`](completed.md)
- [x] S8-D1 — Design: Auto Transcribe in the citation composer → [`completed.md`](completed.md)
- [x] S8-01 — Vision OCR + Auto Transcribe button → [`completed.md`](completed.md)
- [x] S8-D2 — Design: PDF Find, text selection, paste transcription → [`completed.md`](completed.md)
- [x] S8-03 — PDFKit live viewer + I-beam default → [`completed.md`](completed.md)
- [x] S8-04 — PDF Find in the tool strip → [`completed.md`](completed.md)
- [x] S8-05 — Paste transcription from PDF selection → [`completed.md`](completed.md)
- [x] S8-D3 — Design: Evidence graph visual enhancements → [`completed.md`](completed.md)
- [x] S8-06 — Graph visual enhancements (badges, Source jump, bridge copy + Add property) → [`completed.md`](completed.md)
- [x] S8-D4 — Design: Source page enhancements → [`completed.md`](completed.md)
- [x] S8-07 — Source page enhancements (Evidence graph jump + metadata dates as text + delete + clickable `url` + brief bundle) → [`completed.md`](completed.md)
- [x] S8-D5 — Design: Sources list refresh → [`completed.md`](completed.md)
- [x] S8-08 — Sources list refresh (graph-progress counts + brief bundle) → [`completed.md`](completed.md)
- [x] S8-09 — Cross-resource FKs: Artifact / Citation / property terms no longer CASCADE → [`completed.md`](completed.md)
- [ ] S8-12 — Delete-impact registry + `GetDeleteImpact` → [`completed.md`](completed.md)
- [ ] S8-D9 — Design: shared delete confirm / blocked notice → [`completed.md`](completed.md)
- [ ] S8-13 — DeleteImpact recipe → [`completed.md`](completed.md)
- [ ] S8-D12 — Design: Source page resource delete → [`completed.md`](completed.md)
- [ ] S8-14 — Source + Artifact delete → [`completed.md`](completed.md)
- [ ] S8-D13 — Design: Source Types delete → [`completed.md`](completed.md)
- [ ] S8-15 — Source Types delete → [`completed.md`](completed.md)
- [ ] S8-D14 — Design: Source Fields delete → [`completed.md`](completed.md)
- [ ] S8-16 — Source Fields delete → [`completed.md`](completed.md)
- [ ] S8-D15 — Design: Subject Fields delete → [`completed.md`](completed.md)
- [ ] S8-17 — Subject Fields delete → [`completed.md`](completed.md)
- [ ] S8-D11 — Design: Citation composer delete chrome → [`completed.md`](completed.md)
- [ ] S8-18 — Composer Citation + Observation row delete → [`completed.md`](completed.md)
- [ ] S8-D10 — Design: Evidence graph delete chrome → [`completed.md`](completed.md)
- [ ] S8-19 — Graph delete chrome → [`completed.md`](completed.md)
- [ ] S8-99 — Dogfood close / docs (after later stories, or when we choose to close)

---

## S8-D7 — Design: Citation composer rethink

Claude Design board for a **flexible Citation document**: in-form Artifact + Citation identity (replace the create-time pre-screen), Observation list across subjects, empty Save, reuse. **Density is the problem** — the board must explore layout ideas, not a four-list stack. Brief: [`design/archive/S8-D7-composer-rethink.md`](design/archive/S8-D7-composer-rethink.md). Gates **S8-10**. **Done.** Layout A + 1500pt two-column; C/D rejected.

Does **not** design Auto Transcribe (**S8-D1**) or PDF Find/paste (**S8-D2**) — only leave homes on the new chrome.

---

## S8-10 — PR: Flexible citation composer

Replace the subject-locked composer with one place whose document is the Citation. Entry points (Add property, pencil, Connect) only pre-select Artifact / Citation / focused row / subject. Remove the Artifact pre-screen. Show every Observation on the Citation; each row names a Subject. Save with zero Observations. Add property can reuse a Citation on that Artifact.

| | |
| --- | --- |
| **In** | In-form Artifact + Citation identity per **S8-D7**; drop `CitationComposerArtifactPicker` as a phase; load **all** observations (stop filtering `$0.subjectID == subjectID`); subject control on add/edit row; empty Save; reuse from Add property; `ListByArtifact` on the client (FFI if needed); create-without-observations (relax `CreateWithObservations` or add `Create`); location `subjectId` / `citationId` as focus; Connect edge rows unchanged; L10n + VoiceOver; tests. |
| **Out** | Vision / PDF remount / Find / paste; minting Subjects; project-wide Citation search; graph 3-pane; Foundation Models; `WorkspaceLocation.observationId` unless restore requires it. |
| **Testable** | Two-subject Citation shows both rows; pencil does not hide the other subject; Add property can attach to an existing `CIT-…`; Save with empty observations persists; switching Artifact on a dirty saved Citation confirms abandon; one-Artifact source skips a wizard; Connect still submits two fixed edges; Back still returns to the graph. |
| **Depends on** | **S8-D7**. Shipped composer + graph handoff. **Not** S8-01…S8-09. |

---

## S8-D8 — Design: Composer and Connect simplification

Claude Design board that **extends** the S8-10 composer and the shipped graph. It replaces the bundled footer Save with row-level Save / Revert / Delete plus **Save citation**. It adds an unsaved-work guard on every exit, moves Connect's role / relationship-type choice from a graph sheet into a single compact **connection row** in the composer (read-only endpoints; saves onto a new or existing Citation; endpoints cannot be changed), names bridges by their computed sentence, and adds **New person / event / place** to the row subject picker. Brief: [`design/archive/S8-D8-composer-connect-simplification.md`](design/archive/S8-D8-composer-connect-simplification.md). Gates **S8-11**. **Done.** Row-level commits + one-row connection (endpoints fixed); location connections have no term; citation wording, not "reading".

Does **not** design Auto Transcribe (**S8-D1**), PDF Find / paste (**S8-D2**), graph badges / sentence wording (**S8-D3**), or subject / Citation delete chrome (**S8-D10**, **S8-D11**).

---

## S8-11 — PR: Composer and Connect simplification + Interpretation write integrity

One PR that implements the **S8-D8** board **and** the engine, audit, FFI, and client fixes from the S8-10 adversarial review that are still relevant once S8-D8's save model is in. The review found that the composer deletes saved Observations by omission, that edits are lost silently on most exits, that Connect loses the connection when an existing Citation is picked, that the audit log cannot reconstruct what was asserted, and that the server does not check that a bridge's edges point at the chosen endpoints.

**Done.** Row-level Observation commits, Save citation, Connect in the composer, engine-locked edges, attach-to-existing-Citation bridges, full-state create/delete audit, and a navigation leave guard.

**This section is the implementation contract.** Where it names a function, file, field number, error code, or audit field, implement exactly that. If something here turns out to be wrong in the code, stop, fix this section in the same PR (with a one-line reason), then continue. Do not silently diverge.

| | |
| --- | --- |
| **In** | Everything in §S8-11.1 through §S8-11.9 below. |
| **Out** | Subject / bridge / whole-Citation delete and cascades (**S8-09**); undo / ⌘Z; autosave; guarding app quit, window close, project switch, sign-out; rewriting stored labels on existing bridges; creating bridges or `source` subjects from the composer; generalizing `EvidenceBridgeEdgeSummary` sentence templates; any catalog migration; Auto Transcribe / PDF work; graph badges (**S8-06**). |
| **Testable** | §S8-11.10 (every row is a required test). |
| **Depends on** | **S8-D8** agreed. **S8-10** shipped. **Not** S8-01…S8-09. |

### S8-11.1 Frozen decisions

These are decided. Do not reopen them in the PR.

1. **Observations are committed one at a time.** The bundled update RPC `METHOD_UPDATE_CITATION_WITH_OBSERVATIONS` (71) and `observations.ApplyForCitationTx` are **deleted**. Nothing in the product deletes an Observation except an explicit `DeleteObservation` call. **Later exception (§S8-09.1 / 17, S8-19):** official `subjects.Delete` may release **connection facets** (edge rows + disambiguation) with the bridge. Still never delete-by-omission or by Citation cascade.
2. **First commit creates the Citation.** On a New Citation, the first of {row Save, Save citation, Save connection} creates the Citation from the **current citation draft**. Row Save uses `CreateCitationWithObservations` with exactly one draft. Save citation uses it with zero drafts. Save connection uses `CreateCitedBridge` without `citation_id`.
3. **Row commits never write the citation fields** on a saved Citation, and Save citation never writes rows. The only exception is decision 2.
4. **A saved Citation's Artifact is immutable.** `UpdateCitation` has no `artifact_id`. Switching Artifact in the composer means "work on a Citation on another Artifact". It is allowed only when there is no unsaved document work.
5. **Connect stays one atomic write**, with one audit revision `create_cited_bridge`. It may attach to an existing Citation (`citation_id`). The server computes the bridge position, validates that the edge Observations bind the chosen endpoints, and never stores a client label.
6. **Edge rows are engine-locked.** An Observation is an **edge row** when its subject's type is a bridge type and its Property key is one of that bridge type's edge keys in `subjectvocab`. Edge rows are only created by `CreateCitedBridge`. Edge rows are **immutable** from Observation write paths: `DeleteObservation` and `UpdateObservation` on an edge row always fail with `observations.edge_locked`, and so does inserting an edge Property through any other write. A connection is one unit. A wrong endpoint is fixed by deleting the bridge (**S8-19** connection facets) and connecting again, never by editing an edge.
7. **The disambiguation row (role / relationship_type) is required only by `CreateCitedBridge`.** In the engine it is an ordinary term Observation after that (editable, deletable). In the composer it is shown **inside the connection row** (§S8-11.7), where it can be changed but not deleted.
8. **Audit follows `docs/audit-revision-history.md` literally.** Creates record the complete initial row. Deletes record the complete prior row. Updates record changed fields only. Child value rows (`date_values`, `name_values`) and notes get their own `audit_changes` entries. No revision is ever recorded with zero changes.
9. **Primary subject create is one transaction with its position.** `CreateSubject` takes an optional placement. The graph's create and the composer's "New person…" both use it. `SetSubjectPosition` remains for drags only.
10. **Bridges are named by their computed sentence.** New bridges store `label = NULL`. The bridge Edit dialog edits description only and sends the bridge's **current stored label unchanged**, so no audit noise is produced.
11. **The unsaved-work guard lives in `WorkspaceNavigation`** as a single optional leave guard (§S8-11.6). It covers `go(to:)`, `goBack()`, `goForward()`, and `go(toIndex:)`. It does not cover `fallbackToSectionRoot()`, `attachProject`, app quit, or window close.
12. **No schema migration.** Every change here fits the existing tables.
13. **Location connections have no term.** An event ↔ place bridge (`location`, `DisambiguationNone`) is created from its two edges alone. Its connection row:
    - keeps the two-line layout, with the muted text "No role or type" in place of the term picker
    - enables **Save connection** immediately, and sends only the two edge Observations
    - is never "touched", so a pending one never triggers the leave guard
    - once saved, shows only its `CLO-…` ref: no Save / Revert, no ⋯, no Delete, no polarity
    
    Participation (`role`, label "Role") and relationship (`relationship_type`, label "Relationship") keep the term picker.
14. **The S8-D8 board is authoritative for layout; this plan is authoritative for copy.** Where the board says "reading" ("Save reading", "switch reading", "Reading saved"), ship the citation wording in §S8-11.7 instead.

### S8-11.2 Protobuf and FFI contract (`api/proto/engine.proto`)

Run `./scripts/generate-proto.sh` after editing and commit generated Go + Swift. Add each new method to `api/ffi/dispatch.go` as a single `case` calling `handlers.<Name>` (see [`add-ffi-handler`](../../../.cursor/skills/add-ffi-handler/SKILL.md)).

**Method enum**

| Change | Exact edit |
| --- | --- |
| Remove bundled update (commit 7, §S8-11.11) | Delete `METHOD_UPDATE_CITATION_WITH_OBSERVATIONS = 71;`. Add `reserved 71;` and `reserved "METHOD_UPDATE_CITATION_WITH_OBSERVATIONS";` next to the existing `reserved 27;` lines. Delete `UpdateCitationWithObservationsRequest` / `Response` messages, the handler, its test, the `dispatch.go` case, and the Swift store method. |
| Add | `METHOD_UPDATE_CITATION = 75;` |
| Add | `METHOD_UPDATE_OBSERVATION = 76;` |
| Add | `METHOD_DELETE_OBSERVATION = 77;` |
| Add | `METHOD_GET_SUBJECT_FIELDS_WORKSPACE = 78;` |

**New messages**

```proto
message UpdateCitationRequest {
  string project_dir = 1;
  string user_id = 2;
  string citation_id = 3;
  string locator_json = 4;
  string transcription = 5;
  string description = 6;
  bool transcription_uncertain = 7;
  string transcription_note = 8;
}
message UpdateCitationResponse { Citation citation = 1; }

message UpdateObservationRequest {
  string project_dir = 1;
  string user_id = 2;
  Observation observation = 3; // id required; citation_id ignored (read from the stored row)
}
message UpdateObservationResponse { Observation observation = 1; } // listed shape (property_key/label/value_type, date, name)

message DeleteObservationRequest {
  string project_dir = 1;
  string user_id = 2;
  string observation_id = 3;
}
message DeleteObservationResponse {}

message GetSubjectFieldsWorkspaceRequest { string project_dir = 1; }
message SubjectTypeFieldsGroup {
  string subject_type_id = 1;
  repeated SubjectTypeField fields = 2;          // same rows ListSubjectTypeFields returns
  SubjectTypePresentation presentation = 3;      // same as GetSubjectTypePresentation; unset when subjectvocab.TypeByKey has none
}
message GetSubjectFieldsWorkspaceResponse {
  repeated Property properties = 1;
  repeated SubjectType types = 2;
  repeated SubjectTypeFieldsGroup groups = 3;
}

message ConnectEdge {
  string property_key = 1;        // e.g. "person", "event", "place", "related_to"
  string endpoint_type_key = 2;   // subject type key the edge binds, e.g. "person"
}
```

`SubjectTypeField`, `SubjectTypePresentation`, `Property`, and `SubjectType` are the existing messages. Reuse them; do not add parallel shapes.

**Changed messages**

| Message | Exact edit |
| --- | --- |
| `ConnectRule` | Add `repeated ConnectEdge edges = 7;`, ordered **A then B** (same order as `edge_property_keys`). Keep `edge_property_keys = 4` unchanged. |
| `CreateSubjectRequest` | Add `bool has_placement = 7; int64 grid_x = 8; int64 grid_y = 9;`. |
| `CreateSubjectResponse` | Unchanged (subject). Position is implied by the request. |
| `CreateCitedBridgeRequest` | Add `string citation_id = 19;` (commit 5). Remove `label = 7`, `grid_x = 9`, `grid_y = 10`, and add `reserved 7, 9, 10; reserved "label", "grid_x", "grid_y";` (commit 7). |
| `CreateCitedBridgeResponse` | Unchanged shape. `observations` must be the **listed** shape, so the client gets `property_key` and labels without another round trip. In the handler, after `connect.CreateCitedBridge` returns and inside the same `withProjectCatalog` closure, call `observations.ListByCitation(c, res.Citation.ID)`, keep only the rows whose ids are in `res.Observations`, and map them with `listedObservationProto`. |

**Handlers** (`api/ffi/handlers/`)

- `citations.go`: add `UpdateCitation` (commit 5). Delete `UpdateCitationWithObservations` (commit 7).
- `observations.go`: add `UpdateObservation` and `DeleteObservation`. Collapse `observationToInput` and `observationDraftToInput` into one internal `valueInput(...)` helper that both call, so the value-field parsing exists once.
- `subject_defs.go`: add `GetSubjectFieldsWorkspace` next to `ListSubjectTypeFields` / `GetSubjectTypePresentation`. `ListConnectRules` (same file) fills `edges`.
- `subjects.go`: `CreateSubject` passes placement through.
- `connect.go`: map `citation_id`; stop reading label / grid.

**New error code** (`core/apperr/apperr.go`): `CodeObservationsEdgeLocked = "observations.edge_locked"`, `KindConflict`. Swift: `L10n.Errors` case + `error.observations.edge_locked` in `Localizable.xcstrings` ("This observation connects a bridge to its endpoints. Change the endpoint instead of editing or deleting it." — final copy on the board).

### S8-11.3 Engine contract (Go, `core/database/…`)

Follow [`add-catalog-query`](../../../.cursor/skills/add-catalog-query/SKILL.md) and [`use-catalog-session`](../../../.cursor/skills/use-catalog-session/SKILL.md). Every write below is one SQLite transaction that includes its `audit.Record`.

**`audit`**

- `audit.Record` returns `ErrInvalid` when `len(rev.Changes) == 0`. In commit 1, run `rg -n "audit\.Record\(" core --glob '!*_test.go'` (25 call sites today) and make each one skip `Record` when it has no changes. `subjects.Update` already does; `citations.UpdateWithObservations` does not (fix it too; it lives until commit 7).
- Add `audit.FullRow(fields map[string]any) map[string]FieldDiff` (old = nil, new = value) and `audit.DeletedRow(fields map[string]any) map[string]FieldDiff` (old = value, new = nil). These helpers keep every key present, with JSON `null` for unset columns. Use them for every create and delete in this PR.

**`subjectvocab`**

- Add `type ConnectEdge struct { PropertyKey, EndpointTypeKey string }` and `Edges []ConnectEdge` on `ConnectRule`, filled from the seed matrix, ordered A then B, written here as *property key* binds *endpoint type*. Participation: `person` binds `person`, then `event` binds `event`. Relationship: `person` binds `person`, then `related_to` binds `person`. Location: `event` binds `event`, then `place` binds `place`. Check these against `seedConnect`'s `EdgePropertyKeys` order and keep that order.
- Add `func EdgeEndpoint(bridgeTypeKey, propertyKey string) (endpointTypeKey string, ok bool)`. It returns ok for any non-refused rule whose `BridgeTypeKey` matches.
- `observations` will import `subjectvocab`. If that creates an import cycle, move `ConnectEdge` / `EdgeEndpoint` into a new leaf package `core/database/connectedges` that both import, and note it here.

**`subjects`**

- `InsertTx(tx, in CreateInput) (Subject, audit.Change, error)` no longer records a revision and no longer takes `userID`. Its change uses `audit.FullRow` with `id, ref, source_id, subject_type_id, label, description`.
- `Create(c, userID, in CreateInput, placement *Placement) (Subject, error)`, where `type Placement struct{ GridX, GridY int64 }`. In one tx it calls `InsertTx`, then `subjectpositions.SetTx` when `placement != nil`, then records `create_subject` with the one change. Positions stay unaudited (layout).
- Add `GetTx(tx, id)` (exported) for `connect`.
- Use `errors.Is(err, sql.ErrNoRows)` throughout (it already does).

**`subjectpositions`**

- Fix the `SetTx` doc comment (it mentions a `skipExistsCheck` parameter that does not exist).
- Add `GetTx(tx, subjectID) (Position, error)` for the midpoint.

**`citations`**

- Extract `normalizeCreateInput(in *CreateInput) error` (trim + `len(ArtifactID)==16` + `locator.Validate`). `CreateWithObservations` and `InsertWithObservationsTx` both call it once. Delete the duplicated trimming.
- `InsertWithObservationsTx(tx, in, obs []observations.Input, opts observations.InsertOptions) (CreateResult, []audit.Change, error)` no longer records. Its changes are: the citation `FullRow` (`id, ref, artifact_id, locator_json, transcription, description, transcription_uncertain, transcription_note`), one `citation_note` `FullRow` per note (`id, citation_id, body`), then the observation changes from `observations.InsertManyTx`.
- `CreateWithObservations(c, userID, in, obs)` records `create_citation_with_observations` with those changes. It passes `InsertOptions{AllowEdgeRows: false}`.
- **New** `Update(c, userID, citationID []byte, in CitationFieldsInput) (Citation, error)`, where `CitationFieldsInput` = `LocatorJSON, Transcription, Description, TranscriptionUncertain, TranscriptionNote`. It trims, validates the locator, loads the previous row in the tx, updates only the citation columns, and records `update_citation` with changed fields only. With zero changes it commits nothing and returns the stored row. It never touches `citation_notes` or observations.
- Delete `UpdateWithObservations` in commit 7 (§S8-11.11).
- Ref-mint retry loop: when an insert error is **not** a unique conflict, return `mapConstraint(err)` (constraint violation → `ErrInvalid`, anything else → the raw error). Today it always returns `ErrInvalid`, which hides real database errors. Same fix in `observations.insertOne`.
- Use `errors.Is(err, sql.ErrNoRows)` everywhere in the package.

**`observations`**

- In commit 7 (§S8-11.11), delete `ApplyForCitationTx`, delete `listRowsByCitationTx` if it has no other caller, and delete `DeleteByCitationTx` if `rg DeleteByCitationTx` finds no caller outside tests.
- `type InsertOptions struct { AllowEdgeRows bool }`. `insertOne` looks up the subject's type key and the Property key. When `subjectvocab.EdgeEndpoint(typeKey, propKey)` is ok and `!opts.AllowEdgeRows`, it returns `ErrEdgeLocked`.
- `insertOne` returns `[]audit.Change`: the observation `FullRow` (`id, ref, citation_id, subject_id, property_id, polarity, value_text, value_integer, value_date_id, value_name_id, value_subject_id, value_term_id`), then a `date_value` `FullRow` when it inserted a date, a `name_value` `FullRow` when it inserted a name, and one `observation_note` `FullRow` per note.
- `AddToCitation` passes `AllowEdgeRows: false`.
- Replace `resolveValue`'s 7-value return with `type resolvedValue struct { Text string; Integer *int64; DateID, NameID, SubjectID, TermID []byte; DateChange, NameChange *audit.Change }`. It returns `(resolvedValue, error)`. When it calls `datevalues.UpdateTx` / `namevalues.UpdateTx`, it first loads the prior value row and sets `DateChange` / `NameChange` to an update change with changed fields only (nil when nothing changed). When it inserts, it sets a create change.
- **New** `Update(c, userID []byte, in Input) (Listed, error)`. `in.ID` is required. In one tx:
  1. Load the stored row by id (`sqlGetRow`, new). Missing → `ErrInvalid`.
  2. Classify edge-ness from the **stored** subject + Property (decision 6). If it is an edge row → `ErrEdgeLocked`, even when the request changes nothing. Also classify the **requested** subject + Property: moving an ordinary row onto an edge Property of a bridge → `ErrEdgeLocked`.
  3. Run the existing `updateOne` logic (binding check, `resolveValue`, notes when `in.Notes != nil`, release date / name values).
  4. Changes: the `observationUpdateChange` diff (if any), plus `DateChange` / `NameChange` (if any), plus a `date_value` / `name_value` `DeletedRow` for any value row that `releaseDateValue` / `releaseNameValue` deleted (they must now return the deleted row's fields), plus note changes.
  5. Zero changes → commit without a revision. Otherwise record `update_observation`.
  6. Return the listed shape (re-read with `sqlListSelect … WHERE o.id = ?` in the tx).
- **New** `Delete(c, userID, id []byte) error`. In one tx: load the stored row. Missing → `ErrInvalid`. Edge row → `ErrEdgeLocked`. Load its notes. Delete notes (one `observation_note` `DeletedRow` each), delete the row (observation `DeletedRow` with every column), release its date / name value rows (a `DeletedRow` for each one actually deleted). Record `delete_observation`. **Never** delete the Citation, even if it becomes empty.

**`connect`**

- `CreateInput`: remove `Label`, `GridX`, `GridY`. Add `CitationID []byte` (nil = new Citation).
- **All reads happen inside the transaction** (`subjects.GetTx`, `subjecttypes` / `properties` Tx variants — add `GetByIDTx` / `LookupTx` where missing, each the existing query on `*sql.Tx`).
- Replace `sameSource` with `bytes.Equal`.
- Validation, in this order (any failure → `ErrInvalid` unless noted):
  1. From and To exist, are different subjects, and both have `source_id == in.SourceID`.
  2. `rule := subjectvocab.Connect(fromType, toType)`. Refused → `ErrRefused`. `in.BridgeTypeKey` non-empty and different from `rule.BridgeTypeKey` → `ErrInvalid`.
  3. **Bind endpoints to edges.** `rule.Edges` is `[A, B]`. When `A.EndpointTypeKey != B.EndpointTypeKey`, bind each edge to whichever of From / To has that type. When they are equal (person–person relationship), bind A → From and B → To.
  4. **Inputs must be exactly** one row per edge plus, when `rule.Disambiguation` is not `""` / `"none"`, exactly one row whose Property key is `rule.Disambiguation`. Anything else is invalid: extra rows, missing rows, duplicate keys, or a non-empty `SubjectID` on any row.
  5. Each edge row: `ValueSubjectID` equals its bound endpoint's id, and polarity is positive (or empty).
  6. Disambiguation row: `ValueTermID` set (term-on-property is checked later by `resolveValue`).
  7. When `CitationID` is set: the Citation exists, its Artifact's `source_id == in.SourceID`, and every citation field in `in.Citation` is zero-valued (`ArtifactID` empty, `LocatorJSON`, `Transcription`, `Description`, `TranscriptionNote` empty, `TranscriptionUncertain` false, `Notes` empty). When it is not set, `in.Citation` goes through `normalizeCreateInput`.
  8. Both endpoints have positions (`subjectpositions.GetTx`). A missing position is invalid, because the UI always places subjects.
- **Midpoint:** `gridX = floorDiv(fromX + toX + 1, 2)`, `gridY = floorDiv(fromY + toY + 1, 2)`, where `floorDiv` rounds toward −∞. This equals today's client `GraphCanvasGridMapping` midpoint for any integers, including negatives. Add a table test with negative and odd sums.
- Writes: `subjects.InsertTx` (label empty → NULL), `subjectpositions.SetTx`, then either `citations.InsertWithObservationsTx(..., InsertOptions{AllowEdgeRows: true})` (new Citation) or `observations.InsertManyTx(tx, citationID, obs, InsertOptions{AllowEdgeRows: true})` (attach). Set each row's `SubjectID` to the bridge id before inserting.
- Record **one** revision `create_cited_bridge` containing: the subject change, the citation (+ notes) changes when new, and every observation / value change.

**`GetSubjectFieldsWorkspace`** (handler-level composition, no new package): `properties.List`, the same subject-type list `ListSubjectTypes` returns, `subjectvocab.ListBindings(c, typeID)` per type, and `subjectvocab.TypeByKey(type.Key)` for the presentation, all inside **one** `withProjectCatalog` call. A type with no registry entry leaves `presentation` unset. Any other error fails the RPC. The client's `try?` goes away.

**Go idioms in touched files:** `errors.Is` for `sql.ErrNoRows`; no function returns more than three values; no duplicated trim / validate blocks; `bytes.Equal` for id comparison.

### S8-11.4 Swift Platform (`macos/App/Platform/`)

- `GenealogyStore` protocol:
  - **Remove** `updateCitationWithObservations`.
  - **Add** `updateCitation(projectDir:userID:citationID:locatorJSON:transcription:description:transcriptionUncertain:transcriptionNote:) -> CatalogCitation`.
  - **Add** `updateObservation(projectDir:userID:observation: CatalogObservation) -> CatalogObservation`.
  - **Add** `deleteObservation(projectDir:userID:observationID:)`.
  - **Add** `getSubjectFieldsWorkspace(projectDir:) -> SubjectFieldsSnapshot`.
  - **Change** `createSubject(...)` to add `placement: CatalogGridCell?` (new `struct CatalogGridCell: Sendable, Equatable { var gridX: Int64; var gridY: Int64 }`).
  - **Change** `createCitedBridge(...)`: remove `label`, `gridX`, `gridY`; add `citationID: String?`. Return `(CatalogSubject, CatalogCitation, [CatalogObservation])` with listed observations.
- `CatalogConnectRule`: add `var edges: [CatalogConnectEdge]` (`propertyKey`, `endpointTypeKey`). Update `productMatrix` to match the Go seed exactly. Keep the comment that the table mirrors Go.
- `GoStore`: map every change. `FakeStore`: implement the same semantics as Go, because model tests rely on it:
  - edge rows refuse every delete and update with the `observations.edge_locked` error
  - `createCitedBridge` computes the same midpoint, binds endpoints the same way, and supports `citationID`
  - `createSubject` stores the placement
  - `updateCitation` changes citation fields only
- New `Platform/InterpretationValueKinds.swift`: `enum ObservationPolarity: String { case positive, negative }` and `enum PropertyValueType: String { case text, integer, date, name, subject, term }`. `Features/CitationComposer/` and `Features/EvidenceGraph/` must not contain the bare literals `"positive"`, `"negative"`, `"text"`, `"integer"`, `"date"`, `"name"`, `"subject"`, or `"term"` as value-type or polarity comparisons. Wire structs keep `String` fields; compare with `.rawValue`.
- `CatalogQueryRegistry.load(.subjectFieldsWorkspace)` calls `getSubjectFieldsWorkspace` once (no per-type loop, no `try?`).

### S8-11.5 Session cache (`Features/Workspace/Session/WorkspaceSession.swift`)

Add a per-key **generation** so a stale load can never overwrite newer data or clear a newer load's in-flight marker.

- `private var generations: [CatalogQueryKey: Int] = [:]`. Change `inFlight` to `[CatalogQueryKey: (generation: Int, task: Task<Void, Never>)]`.
- `ensureQuery` bumps the generation when it starts a load and captures it. In the task, after `await loader()`: if `generations[key] != captured`, return **without** touching the handle or `inFlight`. Otherwise apply the result, and clear `inFlight[key]` only if its stored generation equals `captured`.
- `invalidate(_:)` bumps the generation (then cancels and clears as today).
- `setQueryValue(_:value:)` bumps the generation and cancels / clears any in-flight load for that key, so an optimistic patch (graph drag) cannot be overwritten by an older load.
- `readyValue` loops on `inFlight[key]?.task` as today. With the guard it can only wait on the newest load.

### S8-11.6 Navigation leave guard (`Features/Workspace/WorkspaceNavigation.swift`)

Follow [`add-workspace-location`](../../../.cursor/skills/add-workspace-location/SKILL.md).

```swift
enum PendingNavigation: Equatable {
    case location(WorkspaceLocation)
    case back
    case forward
    case index(Int)
}

@MainActor
protocol WorkspaceLeaveGuard: AnyObject {
    /// Return true to hold this navigation. The guard must later call
    /// `resumeHeldNavigation()` or `cancelHeldNavigation()`.
    func shouldHoldNavigation(_ pending: PendingNavigation) -> Bool
}
```

- `@ObservationIgnored weak var leaveGuard: (any WorkspaceLeaveGuard)?` (views never observe it) and `private(set) var heldNavigation: PendingNavigation?`.
- `go(to:)`: when `location == currentLocation`, proceed without asking (coalesce). Otherwise ask the guard. Held → store `heldNavigation` and return.
- `goBack()`, `goForward()`, `go(toIndex:)`: ask the guard the same way when a move is possible.
- `resumeHeldNavigation()`: take `heldNavigation`, clear it, and perform it **without** asking the guard again. `cancelHeldNavigation()` clears it.
- `fallbackToSectionRoot()` and `attachProject(...)` never ask the guard.
- `CitationComposerView` sets `navigation.leaveGuard = model` in `.onAppear` and clears it in `.onDisappear` only when `navigation.leaveGuard === model`.
- The composer model conforms. It holds when there is unsaved document work **or** a touched pending connection (definitions in S8-D8 §2.3), and sets `pendingLeave` (an `Identifiable` item carrying the counts) that drives the `.pvConfirm(item:)`. **Discard changes** → `navigation.resumeHeldNavigation()`. **Keep editing** → `navigation.cancelHeldNavigation()`.

### S8-11.7 Composer (`Features/CitationComposer/`)

**Split `CitationComposerModel` (1,700 lines) into these files.** Each type is `@MainActor @Observable final class` unless marked struct. Keep every file under ~500 lines.

| File | Owns |
| --- | --- |
| `CitationComposerModel.swift` | Coordinator: `entry`, `phase`, `prepare()`, Artifact / Citation identity, the menus' enabled state, `WorkspaceLeaveGuard` conformance, `pendingLeave`, the new-subject dialog, and composing the three child models. No row or citation-field logic. |
| `CitationFieldsDraft.swift` | Citation fields (`locator`, `transcription`, `transcriptionUncertain`, `transcriptionNote`, `description`), their saved baseline, `isDirty`, `saveCitation()`. |
| `CitationObservationRows.swift` | `[ObservationRow]`, per-row state + baseline + error, and `commit(rowID:)`, `revert(rowID:)`, `requestDelete(rowID:)`, `confirmDelete()`, add-draft, and the inline edit mutators. |
| `CitationConnections.swift` | `[ConnectionRow]`: at most one **pending** row (built once from a Connect entry: from / to ids + labels, bridge type key, term) plus one **saved** row per bridge whose edges are on this Citation. Each row carries `termProperty: CatalogProperty?`, resolved from the rule's `disambiguation`, and **nil for location**. Owns `isTouched`, `canSave`, `saveConnection()`, `discard()`, and role edit / `commitRole(connectionID:)` / `revertRole(connectionID:)`. With `termProperty == nil`: `canSave` is always true, `isTouched` is always false, and the role commands are unavailable. |
| `CitationComposerVocabulary.swift` (struct) | Built from (fields snapshot, graph snapshot, connect rules, cached terms): `propertiesByID`, `propertiesByTypeID`, `graphSubjects` (bridges labelled with `EvidenceBridgeEdgeSummary.sentence(for:in:)`), `termsByPropertyID`, `isEdge(subjectID:propertyID:)`, `propertyOptions(forSubjectID:)`. The coordinator rebuilds it only when an input changes (store it; do not recompute in computed properties per render). |

**Connection rows, not edge rows.** Edge Observations **never** become `ObservationRow`s. When a Citation loads, `CitationConnections` groups its Observations:
1. Every Observation where `vocabulary.isEdge(subjectID:propertyID:)` is true is grouped by its bridge `subject_id`. Each group becomes one saved `ConnectionRow`. The row shows the bridge sentence (`EvidenceBridgeEdgeSummary.sentence(for:in:)` from the graph snapshot) and has no endpoint controls.
2. For each such bridge whose rule has a `disambiguation` (skip this step for location), the **first** Observation (by ref) on this Citation with that bridge as subject and Property key == the rule's `disambiguation` becomes the row's role (`rolePersistedID`, `roleTermID`, baseline). Any further role Observations on that bridge stay ordinary `ObservationRow`s.
3. Everything else is an ordinary `ObservationRow`.

A saved participation / relationship connection with no role Observation on this Citation shows an empty role picker. Committing it calls `addObservationsToCitation` with the bridge as subject. A saved location connection never shows a picker.

**`ObservationRow`** deletes `isConnectFixed` (there is no endpoint variant). It gains:
- `var persistedRef: String?` — the `OBS-…` ref, set together with `persistedID` from every create / add / load response. It never changes after that and is shown per S8-D8 CS-2b. `ConnectionRow` carries `bridgeRef` (the bridge subject's ref) for the same purpose.
- `var state: RowState` (`.draft`, `.saved`, `.edited`, `.saving`, `.error(String)`)
- `var baseline: RowValues?` (nil for drafts). `RowValues` is the value-bearing subset: subject, property, polarity, value fields, date draft, name draft.

State is derived on every edit: a draft stays `.draft`; a saved row whose values differ from `baseline` becomes `.edited`; if equal, `.saved`.

**Commands** (each one sets the row to `.saving` first and ignores re-entry while saving):

| Command | Condition | Store call | On success | On failure |
| --- | --- | --- | --- | --- |
| Row Save | Citation is New | `createCitationWithObservations(citation draft, [row draft])` | Set `citationID`; citation baseline = current citation fields; row `.saved` with `persistedID` + `persistedRef` + baseline; `session.apply(.savedCitation(sourceId:))`; reload listed citations for the Artifact | Row `.error(message)`; nothing else changes |
| Row Save | Citation saved, row draft | `addObservationsToCitation(citationID, [row draft])` | Row `.saved` with `persistedID` + `persistedRef`; `session.apply(.savedCitation)` | Row `.error` |
| Row Save | Citation saved, row edited | `updateObservation(row as CatalogObservation)` | Row `.saved`, baseline = values; `session.apply(.savedCitation)` | Row `.error` |
| Row Revert | draft | — | Remove the row | — |
| Row Revert | edited / error | — | Restore baseline → `.saved` | — |
| Delete… | saved / edited ordinary row | `.pvConfirm(item:)` then `deleteObservation(persistedID)` | Remove row; `session.apply(.savedCitation)`; Citation stays even if empty | Row `.error`; row stays |
| Save citation | Citation is New | `createCitationWithObservations(citation fields, [])` | Set `citationID`; citation baseline; `session.apply(.savedCitation)`; reload listed citations | Citation error under the citation fields |
| Save citation | Citation saved | `updateCitation(...)` | Citation baseline = current; `session.apply(.savedCitation)` | Citation error |
| Save connection | pending connection row, and (term set **or** `termProperty == nil`) | `createCitedBridge(citationID: current citationID or nil, citation fields only when nil, observations: [edge A, edge B] + [term] only when `termProperty != nil`)` | When the Citation was New: set `citationID` + citation baseline. The pending row becomes a **saved** `ConnectionRow` in place (bridge id + ref from the response; role = the returned term Observation, or none for location). No `ObservationRow`s are added. `session.apply(.savedCitation)` (also invalidates the graph). | Error on the connection row |
| Role Save | saved participation / relationship connection row, role edited | role persisted → `updateObservation(role)`; no role yet → `addObservationsToCitation([role draft on the bridge])` | Role baseline = current; `session.apply(.savedCitation)` | Error on the connection row |
| Role Revert | saved connection row, role edited | — | Restore role baseline | — |
| Discard | pending connection row | — | Remove the pending row; nothing written | — |
| New person / event / place | row subject picker | `createSubject(type, label, description, placement: slot.cell)` where `slot = EvidenceGraphPlacement.composerSlot(in: graphSnapshot, landingColumn: landingColumn)` | Store `landingColumn = slot.landingColumn`; `session.apply(.mutatedSourceGraph)`; set the row's subject | Error in the dialog; dialog stays open; `landingColumn` unchanged |

**Placement for composer-created subjects.** New file `Features/EvidenceGraph/EvidenceGraphPlacement.swift`: a pure `enum EvidenceGraphPlacement` with no store or session access, unit-tested on its own. New subjects **stack in a landing column to the right of the graph**.

Constants — derive them, do not hard-code the numbers:

| Name | Definition | Value today |
| --- | --- | --- |
| `spacing` | `GraphCanvasGridMapping.defaultSpacing` | 40pt |
| `horizontalStep` | `ceil((EvidenceSubjectCard.width + spacing) / spacing)` — clears the widest card (264pt subject; bridge is 236pt) plus one cell of gap | 8 cells |
| `verticalStep` | `ceil((2 * EvidenceSubjectCard.approximateHalfHeight + 2 * spacing) / spacing)` — one new card plus room to grow | 4 cells |
| `minCell` | smallest x such that `(x + 0.5) * spacing - EvidenceSubjectCard.width / 2 >= spacing`; smallest y such that `(y + 0.5) * spacing - approximateHalfHeight >= spacing / 2` | x = 4, y = 1 |
| `maxCell` | largest x such that `(x + 0.5) * spacing + EvidenceSubjectCard.width / 2 <= contentSize.width`; same for y with `approximateHalfHeight` and `contentSize.height` | x = 96, y = 98 |

`contentSize` is today `private static` on `EvidenceGraphView` (4000×4000). Move it to `EvidenceGraphPlacement.contentSize` and have the view read it from there.

`static func composerSlot(in snapshot: SourceGraphSnapshot, landingColumn: Int64?) -> (cell: CatalogGridCell, landingColumn: Int64)`:

1. `placed` = the cells of every primary in `snapshot.subjects` and every bridge in `snapshot.bridges`.
2. **Empty graph:** return cell `(minCell.x, minCell.y + 2)` = (4, 3), with landing column 4.
3. **Pick the column.**
   - When `landingColumn` is nil (first create in this composer visit): `column = min(max(gridX in placed) + horizontalStep, maxCell.x)`.
   - Otherwise `column = landingColumn`.
4. **Pick the row.**
   - When no card in `placed` has `gridX == column`: `row = max(min(gridY in placed), minCell.y)`, i.e. level with the topmost card.
   - Otherwise `row = max(gridY of cards with gridX == column) + verticalStep`.
5. **Wrap.** If `row > maxCell.y` and `column < maxCell.x`: move to `column = min(column + horizontalStep, maxCell.x)` and recompute the row with step 4.
6. **Clamp.** If the row is still `> maxCell.y` (the rightmost column is full), clamp to `maxCell.y`. Overlap is then accepted, since it only happens after roughly 24 subjects in the last column, and the researcher can drag.
7. Return the cell and the column it landed in.

The coordinator keeps `landingColumn: Int64?` for the life of this composer visit, so the second and later creates stack under the first. It starts nil on every visit. A later visit therefore opens a new column to the right of everything, including earlier visits' columns, and step 5 wraps it back inside the canvas.

The snapshot must include the previous create before the next one. `session.apply(.mutatedSourceGraph)` reloads the graph, and the New… option is disabled while that reload is in flight (`graphHandle.status == .loading`).

**Behaviour rules:**
- Changing a row's subject to one whose type does not allow the row's Property: clear Property and value in the **current** values only, keep `baseline`, and show the Property field error. **Save** is disabled until valid. The `buildObservationUpdates` "skip empty property" path is gone, because there is no bundle any more.
- The name / date dialog's confirm writes into the row's current values only (row becomes `.edited` / stays `.draft`). The row's Save commits.
- Artifact and Citation `PVSelect`s are disabled while there is unsaved document work (hint per S8-D8). Delete `ArtifactAbandon`, `pendingArtifactAbandon`, and the abandon confirm.
- Identity switches (`selectCitation`, `selectArtifact`) run through one stored `identityTask: Task<Void, Never>?`. Starting a new switch cancels the previous one. After every `await` inside a switch, check `Task.isCancelled` and return without applying.
- Switching Citation or Artifact **keeps** the pending connection row (endpoints + term) and does not reset it. Saved connection rows are rebuilt from the newly loaded Citation.
- The connection row never offers endpoint editing, edge Property labels, Delete, or polarity. The only editable value is the role term, and a location row has none.
- **Copy (authoritative over the board, decision 14).** Add each string to `L10n` + `Localizable.xcstrings` per [`add-localized-string`](../../../.cursor/skills/add-localized-string/SKILL.md):

  | Surface | String |
  | --- | --- |
  | Citation fields button | "Save citation" |
  | Citation fields status line | New Citation: "First save creates the citation". Dirty: "Citation has unsaved changes". Clean: "Citation saved". |
  | Identity menus disabled hint | "Save or discard changes to switch citation" |
  | Footer unsaved summary / guard Confirm body | "Unsaved: the citation, 2 observations, a new connection" (plural-aware; omit absent parts) |
  | Row state badges | "New", "Edited", "Saving", "Not saved". A Saved row shows only its ref. |
  | Row ⋯ polarity item | "Negate observation" on a positive row, "Affirm observation" on a negative row |
  | Name / date dialog | Confirm "Apply", subtitle "Updates the row only. The row's Save commits it to CIT-…" |
  | Connection term label | "Relationship" (relationship), "Role" (participation) |
  | Location connection term slot | "No role or type" (muted) |
  | Connection row VoiceOver | "Connection, {sentence}. {Relationship / Role} {term or 'not chosen'}." — or "Location." — then "Not saved" / "Edited" / "Saved, {ref}" |
- Footer: **Done** only (goes through `navigation.go(to: graphLocation())`, so the guard applies) plus the unsaved summary. Remove the footer Save, the post-save `session.noticeToast` from the composer, and `didSubmit` / `submitAttempted` / `formError` (errors are per row / per citation / per connection now).
- Remove `applyEntryOverlay`'s Connect branch, `entryConnectPrefill`, and `connectEdgePrefillRows`. The pending connection row is created once from the entry and lives until saved or discarded.
- `CitationComposerEntry.connect` drops `termID`, `gridX`, `gridY`. Update `identityKey`.
- Delete `CitationComposerModel.endpointTypeKey(forEdgePropertyKey:)`, `edgePropertyKeys(typeKey:rules:)` duplication, and `bridgeSentence(...)`. Use `CatalogConnectRule.edges` and `EvidenceBridgeEdgeSummary`.
- The wrong error message for a row with no subject (`dialogPropertyRequired`) becomes a dedicated subject-required row error.

### S8-11.8 Evidence graph (`Features/EvidenceGraph/`) and location

- **Remove** the disambiguation sheet:
  - delete `EvidenceConnectDisambiguationForm.swift` and its Xcode target entry
  - delete `PendingDisambiguation` and every `disambiguation*` member, `confirmDisambiguation`, `cancelDisambiguation`, `selectDisambiguationTerm`, `isDisambiguating`, `canConfirmDisambiguation`, and `midpointCell`
  - delete the sheet presentation in `EvidenceGraphView` and the L10n keys used only by it
- `completeConnectPair`: a valid pair sets `pendingComposerHandoff` directly. The unreachable `do/catch` goes away with the term loading.
- `WorkspaceLocation`: remove `connectDisambiguationTermId`, `connectGridX`, `connectGridY` (properties, `CodingKeys`, init params, decoder, `==`). Persisted history JSON that still has those keys must decode (Swift ignores unknown keys). Add a test that decodes a legacy JSON string with all three keys.
- `connectComposerLocation`: build the title from `EvidenceBridgeEdgeSummary.sentence(kind:…, term: nil)`, using endpoint labels bound via `CatalogConnectRule.edges` (not the hard-coded `switch kind` on person / event / place).
- `SourceGraphSnapshot.build` takes `rules: [CatalogConnectRule]`. `citedEndpoints` reads A / B property keys from the matching rule's `edges` instead of hard-coded keys. Update every caller (graph model, composer vocabulary, tests).
- `confirmPrimaryCreate`: one `store.createSubject(..., placement: CatalogGridCell(gridX: pendingGridX, gridY: pendingGridY))`. Remove the follow-up `setSubjectPosition`.
- Bridge **Edit** dialog: Description only. `confirmEdit` for a bridge sends the bridge's **stored** `subject.label` unchanged, plus the new description.
- **Bridge name fallback** (S8-D8 §2.4) lives in `EvidenceBridgeEdgeSummary`. Every surface that names a bridge calls it: the card, `CitationComposerVocabulary.graphSubjects`, the connection row, and the composer location title.
  - Change the entry point to `sentence(for bridge: SourceGraphPlacedBridge, in snapshot: SourceGraphSnapshot) -> String`, so it can resolve endpoints.
  - **Noun per endpoint:** take the edge Observation's `valueSubjectID` and find that subject in `snapshot.subjects`.
    1. Tier 1 is the identity Observation. It is **S8-06's**; if S8-06 has not landed, skip it.
    2. Tier 2 is the endpoint's `subject.label`, trimmed, when non-empty.
    3. Tier 3 is `L10n.EvidenceGraph.bridgeNounTypeAndRef(type: placed.typeLabel, ref: subject.ref)`, a new key (`"%@ %@"`, e.g. "Person CPR-F4N2P").
    
    When the endpoint is not in the snapshot, fall through to the whole-name fallback.
  - **Role:** unchanged (term display). Missing role → the existing role-less template.
  - **Whole-name fallback** applies when either endpoint noun cannot be resolved:
    - Tier A is the bridge's stored `subject.label`, trimmed, when non-empty.
    - Tier B is `L10n.EvidenceGraph.bridgeNameKindAndRef(phrase: phrase(kind:term:), ref: subject.ref)`, a new key (`"%@ · %@"`).
  - The function must never return an empty string. Delete the old `sentence(for:)` without a snapshot.
- Primary create / edit dialog: add the working-label caption (S8-D8 CS-14).
- `prepare()`: remove `_ = fields` / `_ = sources`. `EvidenceGraphView.task` stops warming query keys itself (the model's `prepare()` already does). Move `refreshSourceChrome` into the model as computed `sourceTitle` / `sourceRef` from the `sourcesList` handle, and delete the view's duplicated key properties.
- **Drag persistence:** add `private var positionGeneration: [String: Int]` and `private var confirmedPositions: [String: CatalogGridCell]` (seeded from the snapshot on first touch). Each `commitDrag` / `moveSubject` bumps the subject's generation. On success, set `confirmedPositions`. On failure, revert to `confirmedPositions[subjectID]` **only if** the generation still matches, then toast.

### S8-11.9 Docs to update in the same PR

- `docs/audit-revision-history.md`: add a short "Interpretation action types" list (`create_subject`, `update_subject`, `delete_subject`, `create_citation_with_observations`, `update_citation`, `add_observations`, `update_observation`, `delete_observation`, `create_cited_bridge`) and note that `name_value` changes carry `parts` as one JSON array field.
- `docs/macos-client-patterns.md`: one paragraph on the leave guard and the row-commit pattern.
- `.cursor/skills/add-workspace-location/SKILL.md`: mention `WorkspaceLeaveGuard` for places with unsaved work.
- `completed.md`: S8-11 entry. This file: tick S8-11.
- S8-D11 notes that single Observation delete is S8-11's. Later delete PRs reuse `observations.Delete`.

### S8-11.10 Required tests

Go (`CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...`):

| Package | Test |
| --- | --- |
| `audit` | `Record` with zero changes → `ErrInvalid`. `FullRow` / `DeletedRow` keep null keys. |
| `citations` | Create audit has all eight citation fields (nulls included) + note rows + full observation rows. `Update` changes only citation fields, records changed fields only, records nothing on a no-op, never touches notes. A non-unique insert error is not turned into `ErrInvalid`. |
| `observations` | `Update`: date edit in place records a `date_value` update change; name edit records a `name_value` change; no-op → no revision; released date recorded as a `date_value` delete. `Delete`: records every column + notes + released values; Citation still exists afterwards (including when it was the last row). Edge rows: delete → `edge_locked`; any update (including a no-op and a `value_subject_id` change) → `edge_locked`; moving an ordinary row onto a bridge edge Property → `edge_locked`. `AddToCitation` with an edge Property on a bridge subject → `edge_locked`. Updating and deleting the role Observation of a bridge still works (ordinary row in the engine). |
| `connect` | Location (event ↔ place, either order) with exactly the two edge rows → ok, and no term Observation is written. Location with an extra term row → invalid. Endpoint mismatch → invalid. From/To in either order bind by type. Person–person binds A→From, B→To. Extra row / missing term / non-empty SubjectID → invalid. Attach to an existing Citation: no new `citations` row; one revision; wrong-Source Citation → invalid; non-empty citation fields with `citation_id` → invalid. Midpoint table incl. negative + odd sums. Exactly one `create_cited_bridge` revision containing subject + citation + observation changes. A failing insert rolls back (no subject row). |
| `subjects` | `Create` with placement writes subject + position in one tx and one revision. Without placement, no position row. |
| `subjectvocab` | `EdgeEndpoint` for every bridge type; `ListConnectRules` edges in A, B order. |
| `api/ffi/handlers` | `runRPC` tables for `UpdateCitation`, `UpdateObservation`, `DeleteObservation`, `GetSubjectFieldsWorkspace`, changed `CreateSubject` / `CreateCitedBridge` / `ListConnectRules`. `observations.edge_locked` surfaces as that code. |
| `api/ffi` (`dispatch_test.go`) | Method 71 → the unknown-method error (router-only test, per `add-ffi-handler`). |

Swift (`xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'`, Swift Testing per [`add-swift-test`](../../../.cursor/skills/add-swift-test/SKILL.md)):

| Suite | Test |
| --- | --- |
| `WorkspaceSessionTests` | A slow older load finishing after a newer one does not overwrite the value or clear the newer `inFlight`. `setQueryValue` during an in-flight load wins. `readyValue` waits for the newest load. |
| `WorkspaceNavigationTests` | Guard holds `go(to:)` / back / forward / index. Resume performs the held move once. Cancel drops it. Same-location `go(to:)` does not ask. `fallbackToSectionRoot` does not ask. Weak guard released → navigation proceeds. |
| `CitationComposerModelTests` (rewrite) | First row Save on New creates the Citation with the citation fields and one Observation, and the row then exposes the returned `OBS-…` ref. Draft rows expose no ref; loaded rows expose their stored ref; a saved connection row exposes the bridge ref. Second row Save calls add, not create. Edited row Save calls update only. Revert draft removes it; revert edited restores it. Delete asks, then calls delete; Citation stays. Incompatible subject change never calls delete, and Save is disabled. Save citation on New → empty Citation; on saved → `updateCitation` only (no row writes). Menus disabled with unsaved document work; enabled with only a pending connection. Leave guard holds for dirty citation fields / edited row / touched connection; does not hold for an untouched connection or empty drafts. Pending connection on New and on an existing Citation (passes `citationID`, no citation fields). Switching Citation keeps the pending connection. A pending location connection: `canSave` is true with no term, `isTouched` stays false (the leave guard does not hold for it alone, and still holds for other unsaved work), and Save connection sends exactly two observations. A saved location connection exposes no role commands. Participation rows use the "Role" label; relationship rows use "Relationship". The polarity menu item reads "Negate observation" / "Affirm observation" by polarity. The name / date dialog confirm only updates the row (state becomes Edited; no store call). Loading a Citation with a bridge's edges yields one saved connection row (no `ObservationRow` for any edge) with the role attached; a second role on the same bridge stays an ordinary row. Save connection turns the pending row into a saved connection row (no rows appended). Role change on a saved connection calls `updateObservation` (or add when missing); the connection row exposes no delete and no endpoint edit. New person creates with `EvidenceGraphPlacement.composerSlot` placement and fills the row. A second New person in the same visit reuses `landingColumn`. New… is disabled while the graph reload is in flight. Two quick Citation switches apply only the last. |
| `EvidenceGraphModelTests` | A valid Connect pair hands off immediately (no disambiguation state); the location has no term / grid keys. Primary create calls `createSubject(placement:)` once and never `setSubjectPosition`. Drag failure after a newer drag does not revert the newer move. Bridge edit sends the stored label unchanged. |
| `EvidenceGraphPlacementTests` | The constants equal their formulas (8, 4, x 4 / y 1, x 96 / y 98 for today's sizes). Empty graph → (4, 3). First create is `horizontalStep` right of the rightmost card and level with the topmost card. The new card's frame (width × `2·approximateHalfHeight`, centered on the cell) does not intersect any existing card's frame, for both a subject and a bridge as the rightmost card. A second create with the returned `landingColumn` stacks `verticalStep` below the first. Column full → wraps to the next column at the top row. Rightmost column full → clamps to `maxCell.y`. The rightmost card near the canvas edge → column clamps to `maxCell.x`. The result is always inside `minCell…maxCell`. |
| `EvidenceBridgeEdgeSummaryTests` | Endpoint with a working label → label noun. Endpoint with a blank label → "Type REF" noun. Endpoint missing from the snapshot, bridge with a stored label → stored label. Same with no stored label → "Kind phrase · REF". Missing role → role-less template. No input combination returns an empty string. |
| `WorkspaceLocation` / `SourceGraphSnapshotTests` | Legacy JSON with `connectDisambiguationTermId` / `connectGridX` / `connectGridY` decodes. Endpoints come from `rule.edges`. |

Also run `python3 scripts/check-localizable-xcstrings.py` and discard accidental catalog churn.

### S8-11.11 Commit order

Land as these commits, in order. Each commit builds, and its tests pass, before the next starts.

1. **Audit helpers + zero-change guard.** `audit.FullRow` / `DeletedRow`; `Record` rejects empty; fix every call site. Tests.
2. **Engine: full-state audit on existing creates.** `subjects.InsertTx` / `citations.InsertWithObservationsTx` / `observations.insertOne` return changes; callers record; `normalizeCreateInput`; ref-retry error mapping; `resolvedValue` struct. Tests.
3. **Engine: edge classification + observation Update / Delete.** `subjectvocab` edges + `EdgeEndpoint`; `InsertOptions`; `observations.Update` / `Delete`; `citations.Update`. **Do not delete** `ApplyForCitationTx` / `UpdateWithObservations` yet (the app still calls them); make them record full-state changes via the new helpers so they stay correct until commit 7. Tests.
4. **Engine: Connect hardening.** Tx-only reads, endpoint binding, exact-inputs rule, attach, midpoint, single revision; `subjects.Create` placement. The `CreateCitedBridge` and `CreateSubject` handlers are updated to the new Go signatures in this commit: they stop passing `label` / `grid_x` / `grid_y` (those proto fields still exist and are now ignored), and pass `nil` placement / `nil` citation id. Tests.
5. **Proto + FFI (additive).** Add methods 75–78, `ConnectEdge` / `edges`, `CreateSubjectRequest` placement, and `CreateCitedBridgeRequest.citation_id` (field 19). Handlers + `dispatch.go` + `runRPC` tests. Regenerate. The Swift protocol / `GoStore` / `FakeStore` gain the new methods. `createSubject` gains `placement:` and `createCitedBridge` gains `citationID:`; update their one call site each to pass `nil`. **Nothing is removed from the proto in this commit.**
6. **Session generation guard + navigation leave guard.** §S8-11.5, §S8-11.6. Tests.
7. **Composer rewrite + removal of the bundled path.** §S8-11.7 in full. In the same commit, delete everything the old footer Save used:
   - `METHOD_UPDATE_CITATION_WITH_OBSERVATIONS` (reserve 71 + name) and its messages
   - its handler, handler test, and `dispatch.go` case
   - `citations.UpdateWithObservations` and `observations.ApplyForCitationTx`
   - the Swift `updateCitationWithObservations` in protocol / `GoStore` / `FakeStore`
   - `CreateCitedBridgeRequest` fields 7 / 9 / 10 (reserve numbers + names) and the matching Swift parameters
   
   Regenerate. Tests.

   Between commits 7 and 8 the graph's disambiguation sheet still exists, and `WorkspaceLocation` still carries `connectDisambiguationTermId` / grid keys. The rewritten composer **ignores** them: `CitationComposerEntry(location:)` stops reading them in this commit, and the pending connection starts with an empty term. This is expected. Do not add a bridge to carry the old term through.
8. **Graph + location.** §S8-11.8. Tests.
9. **Docs + L10n cleanup.** §S8-11.9; prune dead L10n keys; xcstrings check.

### S8-11.12 Do not

- Do not keep any code path that deletes an Observation because it was missing from a payload.
- Do not add autosave, save-on-blur, or ⌘Z.
- Do not let `DeleteObservation` delete a Citation.
- Do not trust the client for bridge labels, bridge positions, or endpoint binding.
- Do not record audit revisions outside the write's transaction, or with zero changes.
- Do not add a DesignSystem component. S8-D8's inventory is binding.
- Do not add a catalog migration.
- Do not bump `VERSION` unless this PR is cut as a release.

---

## S8-D1 — Design: Auto Transcribe in the citation composer

Claude Design board for the **transcription** field: Auto Transcribe control, in-progress state, replace confirm, large-page / slow-job warning that can still proceed, and failure/empty states. Brief: [`design/archive/S8-D1-auto-transcribe.md`](design/archive/S8-D1-auto-transcribe.md). Gates **S8-01**. **Done.** Trailing secondary **Auto transcribe**; combined replace + whole-page confirm; field locked while running; image-only.

Does **not** design Observation auto-fill, LLM extract, PDF OCR, or PDF Find (**S8-D2**).

---

## S8-01 — PR: Vision OCR + Auto Transcribe

On-device Vision (`VNRecognizeTextRequest`) fills the composer **transcription** textarea for **image** Artifacts only. Crop in memory from the loaded `NSImage` and the region locator when present. No temp file, no catalog write until the researcher Saves. PDF / audio / video: do not run Vision.

**Done.** Reusable `Features/OCR` module; composer wires image-only enablement, locator crop, replace / whole-page confirms, and a running lock. No PDF OCR, Observations, Live Text, or new kit primitive.

| | |
| --- | --- |
| **In** | Protocol-shaped OCR seam (tests do not call Vision); `CGImage` from the already-loaded **image**; bounding-box crop (optional mask later); Auto Transcribe control per **S8-D1**; replace confirm if transcription is non-empty; preflight warn on artifact-only / huge bitmap / dense-image heuristics, with proceed; L10n; skip PDF / audio / video / no image. |
| **Out** | PDF OCR; PDF page raster → Vision; PDF Find / select / paste (**S8-03…S8-05**); Foundation Models; writing Observations; persisted crop objects; `RecognizeDocumentsRequest` (macOS 26); raising the deployment target. |
| **Testable** | Fake recognizer fills / fails / empty; crop uses region vs full page; preflight flags large page; replace does not overwrite without confirm; button disabled while running. |
| **Depends on** | **S8-D1**, **S8-10**, **S8-11**. Locators (S7-06 / S7-07). Do not land on the pre-rethink form or the S8-10 footer-Save form; the control sits on the S8-11 citation fields section. |

---

## S8-D2 — Design: PDF Find, text selection, paste transcription

Claude Design board for the PDF **tool strip** (Find), **cursors** (I-beam default; scroll pans like Preview), and transcription **Paste from selection**. Brief: [`design/archive/S8-D2-pdf-text-find.md`](design/archive/S8-D2-pdf-text-find.md). Gates **S8-03**, **S8-04**, **S8-05**. **Done.** Find is a trailing search IconButton + under-strip row (wrap). Pan is Preview-style scroll. PDF replaces Auto transcribe with Paste from selection.

Does **not** design image OCR, PDF thumbnails, or `text_quote` locators ([`text-quote-locators.md`](../../ideas/text-quote-locators.md)).

---

## S8-03 — PR: PDFKit live page + I-beam default

Replace the composer PDF **raster** (`displayImage` / `ArtifactMediaViewport`) with a **PDFKit-backed** page so `PDFSelection` exists. Default drag **selects text**. Pan is the `PDFView` scroll view (trackpad, wheel, scrollbars), same as Preview — not a hand tool or modifier (per **S8-D2**). Region overlay (S7-07) still draws in page space. Image viewer unchanged.

**Done.** Live `PDFView` in `ArtifactPDFViewport`; I-beam + scroll-to-pan; region overlay remapped to media-box unit-square. No Find, paste, or image-pan change.

| | |
| --- | --- |
| **In** | Live `PDFDocument` / `PDFView` (or equivalent) for PDF Artifacts; I-beam default; scroll-to-pan; cursors; region tools exclusive with select; zoom/page chrome still work; no-text-layer is paintable (select does nothing useful). |
| **Out** | Find UI (**S8-04**); paste button (**S8-05**); Vision; changing image pan; Source-page viewer. |
| **Testable** | PDF path no longer depends on `displayImage` for hit-testing text; image path unchanged; region draft still normalizes; scroll/trackpad still moves the page; selecting text does not pan. |
| **Depends on** | **S8-D2**, **S8-10**, **S8-11**. Locators S7-07. **Not** S8-01. Do not remount the pre-rethink viewer slot. |

This is the load-bearing remount. Do not start S8-04 / S8-05 until it lands.

---

## S8-04 — PR: PDF Find

Find field on `ArtifactViewerToolChrome` (PDF only). `PDFDocument.findString` → highlight + page jump + next/previous.

**Done.** Composer under-strip Find; `findString` highlight + wrap + page jump. No paste, Vision, or Source-page Find.

| | |
| --- | --- |
| **In** | Keyword field; Find / next / prev; current-hit highlight; sync `model.page`; no-match and no-text-layer copy; L10n; disable while a region tool is drawing if the board says so (Find itself may stay). |
| **Out** | Source-page Find; `text_quote` locators ([`text-quote-locators.md`](../../ideas/text-quote-locators.md)); OCR fallback; image Find. |
| **Testable** | Fake/document fixture: hit changes page; wrap/stop per board; empty query / no hits; hidden on image Artifacts. |
| **Depends on** | **S8-D2**, **S8-03**. |

---

## S8-05 — PR: Paste transcription from PDF selection

Transcription-row control for PDF: copy current `PDFSelection` string into `transcription`. Replace confirm if the field is non-empty (same as **S8-01**).

**Done.** PDF transcription slot is Paste from the I-beam selection; replace confirm; no Vision. Image row stays Auto Transcribe.

| | |
| --- | --- |
| **In** | Button/label per **S8-D2**; disabled with no selection / no text layer / `inert`; replace confirm; does not write Observations; image row stays Auto Transcribe. |
| **Out** | Vision on PDF; auto-running paste; locator writes. |
| **Testable** | Selection string fills the field; empty selection disabled; replace does not overwrite without confirm; image Artifact does not show an enabled Paste. |
| **Depends on** | **S8-D2**, **S8-03**. **Prefer after S8-01** so the transcription `PVField` action row is designed once. |

---

## S8-D3 — Design: Evidence graph visual enhancements

Claude Design board for a **bundled** graph-chrome pass. Items: **conflict** + **negated** row marks; always-on **Source-page jump**; **bridge sentences** that prefer endpoint identity Properties; **Add property** on bridge cards. Brief: [`design/archive/S8-D3-graph-visuals.md`](design/archive/S8-D3-graph-visuals.md). Gates **S8-06**. **Done.** Ochre conflict bracket + “Not” prefix (not `PVBadge`); ghost **Jump to Source page**; identity-Observation nouns; Add property + extra non-edge rows on bridges.

Does **not** design the composer rethink (**S8-D7**) or denied-lines. Source-page → graph is **S8-D4**. **Descoped:** incomplete-bridge chrome, collapse/expand, density filters, undo, unplaced tray, minimap.

---

## S8-06 — PR: Graph visual enhancements

**Done.** One Evidence graph chrome pass against **S8-D3**. Competing Observations stay as separate rows. Duplicate `propertyKey` → ochre gutter bracket on the run; `polarity = negative` → “Not” prefix (may stack). Header **Jump to Source page** opens the same Source’s detail page. Cited bridge sentences prefer each endpoint’s identity Observation (`name` / `event_type` / `toponym`), then working label. Bridge cards get **Add property** (reuse `composerLocation(for:)`) and show extra **non-edge** Observation rows.

| | |
| --- | --- |
| **In** | Conflict + negated per **S8-D3**; Source jump via existing `sourcePageLocation` + `go(to:)`; `EvidenceBridgeEdgeSummary` reads endpoint Observations from the snapshot; Add property + extra rows on bridges; L10n + VoiceOver; card height / hit tests. Prefer `PVBadge`. |
| **Out** | Merge / resolve; schema or FFI; denied-line drawing; Source-page → graph (**S8-07**); composer rethink (**S8-D7** / **S8-10**). Incomplete-bridge chrome, collapse/expand, filters, undo, tray, minimap are **descoped**. |
| **Testable** | Two `name`s → both conflict; negative singleton → negated only; jump location is `.page` for the same `sourceId` and Back returns to `.graph`; relationship sentence uses NameValue form when present and label when not; participation uses `event_type`; location uses `toponym`; bridge Add property opens composer for that bridge (not the connect Citation); extra non-edge row visible + editable; edge keys not duplicated as rows; height/a11y follow the new sentence. |
| **Depends on** | **S8-D3**. Shipped cards + snapshot + `sourceSurface`. **Not** S8-01…S8-05 / **S8-07**. |

---

## S8-D4 — Design: Source page enhancements

Claude Design board for a **bundled** Source-page chrome pass. **This board redraws metadata from the shipped Mac section** (quick-add / quick-edit), drops DateValue catalog chrome, **keeps a delete on saved values**, and styles saved `url` values as **external links**. **Jump to Evidence graph is already drawn** — do not brief it again; **S8-07** implements it. More page items join this brief (and **S8-07**) as they are scoped. Brief: [`design/archive/S8-D4-source-page.md`](design/archive/S8-D4-source-page.md). Gates **S8-07**.

Does **not** redesign the jump, keep the board’s date-entry experiments, graph chrome, source-to-source commentary ([`source-to-source-relationships.md`](../../ideas/source-to-source-relationships.md)), composer DateValue chrome, a first-class provenance date ([`source-provenance-date.md`](../../ideas/source-provenance-date.md)), composer rethink (**S8-D7**), or Sources-list counts (**S8-D5**).

---

## S8-07 — PR: Source page enhancements

One Source-page pass against **S8-D4**. Add a control that opens this Source’s Evidence graph (`sourceSurface: .graph`). Same `hasArtifact` rule as the list split-row. Keep the shipped metadata quick-add / quick-edit; treat date fields as text; **delete** a saved value via `ClearSourceMetadata`. Saved `url` values are **links** that open in the default browser; save requires a host-shaped value (`https://…` **or** `www.url.com`; no reachability). Drop `data_type = date` / `date_value_id` from source metadata. Pin **Credibility** to the bottom of the left overview column (description + credibility | metadata) — already drawn on the board; the app currently stacks it immediately under the description. Nudge the **title edit** pencil from the top-right of the title block toward mid-title (already on the board). Further items listed on the brief at PR start ship here.

| | |
| --- | --- |
| **In** | Jump control from the **existing** S8-D4 frames (`SourcesListNavigation.graphLocation`, `hasArtifact`, `go(to:)`, **Evidence graph** copy); shipped metadata chrome + dates as text (SP-1…SP-3); **delete** on saved rows → `clearSourceMetadata` (SP-4); saved `url` values as links → default browser (SP-6); **`url` shape check** on Set (SP-7: `http`/`https` + host **or** scheme-less host like `www.url.com`; same rule in Go `validateValue`); catalog migration + FFI so `source_metadata` is `value_text` only (`data_type` `text` \| `url`); seed `record_date` / `issue_date` as text; search projector stops joining `date_values` for metadata; **Credibility** aligned to the **bottom** of the left overview column (`overviewColumns` in `SourcePageView` — today a `VStack` under Description); title **edit** pencil vertically nearer the **middle** of the title (`SourcePageIdentityHeader` / `PVInlineEdit` currently `.top`); L10n + VoiceOver; any other SP items frozen on the brief. Prefer `PVButton` / `PVIconButton`. |
| **Out** | Graph chrome (**S8-06**); opening the graph with zero Artifacts; source-to-source commentary ([`source-to-source-relationships.md`](../../ideas/source-to-source-relationships.md)); composer / Observation DateValue; first-class provenance-date column or list sort ([`source-provenance-date.md`](../../ideas/source-provenance-date.md)); deleting a Source Fields vocabulary row; redesigning credibility chips / argument; restyling the identity header beyond the title-pencil alignment; in-app browser / WebView; URL reachability / allow-lists; composer rethink (**S8-10**); list counts / refresh (**S8-08**). |
| **Testable** | Source with an Artifact → location is `.graph` for the same `sourceId`; Back returns to `.page`; no Artifact → control disabled and does not navigate. Date-named metadata saves and edits as text; SetSourceMetadata rejects a DateValue; existing `date` fields / `date_value_id` migrate to text (keep `value_text`; if only a DateValue existed, flatten a readable phrase into `value_text` then drop the FK). Delete clears a saved value; a type suggestion returns to the dashed list; an extra field is gone. `https://example.com` and `www.url.com` save and look like links (scheme-less opens as `https://` + the string); `htps://…`, `file:/tmp`, and strings with no host fail save with the field error (engine too). When the metadata column is taller than Description, Credibility sits at the bottom of the left column (spacer / bottom alignment), not flush under Description. Title edit sits mid-title (or first-line center), not pinned to the top-right of the title block. Composer date Observations still insert DateValues. |
| **Depends on** | **S8-D4**. Shipped Source page + `sourceSurface`. **Not** S8-01…S8-06 / **S8-08**. |

---

## S8-D5 — Design: Sources list refresh

Claude Design board for a **bundled** Sources-list chrome pass. First item: **subject** + **observation** counts per Source (optional uncited). More list items join this brief (and **S8-08**) as they are scoped. Brief: [`design/archive/S8-D5-sources-list.md`](design/archive/S8-D5-sources-list.md). Gates **S8-08**.

Does **not** design graph or Source-page chrome, commentary, or folding counts into `CatalogSource`.

---

## S8-08 — PR: Sources list refresh

One Sources-list chrome pass against **S8-D5**. Show graph-progress counts so a worked Evidence graph is obvious next to an empty one. Counts live on their **own** session cache identity (per `sourceId` or a patchable map — follow `.citationCounts`). A canvas write invalidates **that Source’s** numbers only. Further items listed on the brief at PR start ship here.

| | |
| --- | --- |
| **In** | List chrome per **S8-D5**; Go aggregates (subjects on this Source, minus `source` type; Observations that mean “filled out”); new `CatalogQueryKey` + registry `invalidateOn` for `.createdSubject` / `.createdCitation` / `.addedObservations` (and delete if a mutation tag exists); warm with the Sources place; L10n + VoiceOver; any other SL items frozen on the brief. |
| **Out** | Counts on `GetSources` / `CatalogSource`; invalidating `sourcesList` on graph writes; a Subjects list destination; commentary; graph / Source-page chrome. |
| **Testable** | Worked Source shows non-zero subjects + observations; empty Source shows honest zero; no-Artifact row stays blocked; creating a subject or Observation updates **only** that Source’s count key; `sourcesList` handle does not refetch from the canvas write. |
| **Depends on** | **S8-D5**. Shipped split-row list + session cache. **Not** S8-01…S8-07 / **S8-09**. |

---

## S8-D9 — Design: Shared delete confirm / blocked notice

Rethink of resource-delete confirmation. One recipe: **confirm** when Impact is allowed, **notice** when it is not. Brief: [`design/S8-D9-impact.md`](design/S8-D9-impact.md). Gates **S8-13**. Paste **first**.

Does **not** place trash on any screen. Later boards instance this recipe.

---

## S8-D12 — Design: Source page resource delete

Enhancement of the shipped Source page. Delete Source and Delete Artifact via DeleteImpact. Metadata-row delete stays a facet Confirm. Brief: [`design/S8-D12-source-delete.md`](design/S8-D12-source-delete.md). Gates **S8-14**.

---

## S8-D13 — Design: Source Types delete

Enhancement of the shipped Source Types page. Brief: [`design/S8-D13-source-types-delete.md`](design/S8-D13-source-types-delete.md). Gates **S8-15**.

---

## S8-D14 — Design: Source Fields delete

Enhancement of the shipped Source Fields page. Brief: [`design/S8-D14-source-fields-delete.md`](design/S8-D14-source-fields-delete.md). Gates **S8-16**.

---

## S8-D15 — Design: Subject Fields delete

Enhancement of the shipped Subject Fields page. Brief: [`design/S8-D15-subject-fields-delete.md`](design/S8-D15-subject-fields-delete.md). Gates **S8-17**.

---

## S8-D11 — Design: Citation composer delete chrome

Enhancement of the S8-D8 composer. **Delete citation** when empty; notice when any Observation remains. Observation **row** Delete… moves onto DeleteImpact (inbound empty today). Brief: [`design/S8-D11-composer-delete.md`](design/S8-D11-composer-delete.md). Gates **S8-18**.

---

## S8-D10 — Design: Evidence graph delete chrome

Enhancement of shipped card trash. Trash on every card; DeleteImpact confirm or notice (including endpoint-only G2). Brief: [`design/S8-D10-graph-delete.md`](design/S8-D10-graph-delete.md). Gates **S8-19**.

---

## S8-09 — PR: Cross-resource FKs (`NO ACTION`)

One format step. The three FKs that treat a **different resource** as a child: `artifacts.source_id`, `citations.artifact_id`, and `property_terms.property_id`. SQLite cannot `ALTER` the action — rebuild those tables (`PRAGMA foreign_keys=OFF`, copy, rename, `foreign_key_check`). Follow [`add-catalog-migration`](../../../.cursor/skills/add-catalog-migration/SKILL.md) (`000030.sql`). Update the CREATE TABLE copies in `interpretation-layer-data-model.md` / `source-layer-data-model.md`.

**Do not** strip CASCADE from facets (notes, metadata, positions, …). Those stay. Existing `Delete` functions keep working.

| | |
| --- | --- |
| **In** | FKs marked **flip** in §S8-09.3: rebuild `artifacts` (`source_id` → `NO ACTION`), `citations` (`artifact_id` → `NO ACTION`), and `property_terms` (`property_id` → `NO ACTION` — terms are their own resource). Partial unique indexes on `observations.value_date_id` and `observations.value_name_id` (`WHERE … IS NOT NULL`) so two Observations cannot share a value row. Keep `file_id` and remaining facet CASCADEs. Keep `sources.primary_artifact_id … ON DELETE SET NULL`. |
| **Out** | Other facet CASCADEs; unique on `value_subject_id` / `value_term_id` (those **are** shared); `deleteimpact` (**S8-12**); `citations.Delete`; new FFI / chrome; Source / Artifact delete UI; `VERSION` bump. |
| **Testable** | Deleting a Source row while an Artifact exists fails. Deleting an Artifact while a Citation exists fails. Deleting a Source that has only notes / metadata still succeeds and those rows vanish (CASCADE unchanged). Empty Citation still dies with its Artifact until this lands — after it, that delete fails (the point). Second Observation attaching an existing `value_date_id` / `value_name_id` fails unique. In-place date/name edit on the same Observation still works. Deleting a property while a term row exists fails (no term CASCADE). `subjects.Delete` / `observations.Delete` unchanged. |
| **Depends on** | §S8-09.1. Shipped catalog. **Not** S8-12…S8-19. |

---

## S8-09.1 Frozen delete policy

These are decided. Later PRs and briefs implement them; they do not reopen them.

1. **Buckets are named on the FK, including the wrong-way ones.** A table may be a **resource** (Source, Artifact, Citation, Observation, Subject, later Claim / Narrative) or a **facet** of one resource (notes, metadata, layout, positions, name parts, vocab joins). **Resource → resource** is `NO ACTION`. **Facet → parent** is `ON DELETE CASCADE` (child points at parent; SQLite deletes the child). **Owned outbound** is the other way: the parent row holds `parent.col → child.id` (`artifacts.file_id`, `observations.value_date_id`). SQLite will not delete the child when the parent goes. The registry must still list that edge so official `Delete` releases the child. The app must not be the only gate on a **resource** pointer.

2. **The app explains the schema; it does not replace it.** `Impact` is a **speakable report** of inbound **resources** that block the delete (see §S8-09.4) — not a boolean and not a bare count. CASCADE facets are not blockers. Owned outbound are not blockers either — they are work the official `Delete` performs after the parent row is gone (or in the same tx). A successful erase still runs in a transaction the FK graph would accept.

3. **The map is catalog-wide.** Register every **resource** a researcher might delete, even when this spike offers no UI: Sources, Artifacts, Citations, Observations, Subjects, later Claims / canonical entities / Narratives. Vocabulary rows that already use `ErrInUse` are resources too. Adding a table is “classify it, then register its inbound FKs.”

4. **Erase iff inbound *resources* are empty.** CASCADE facets go with the parent in SQLite. Owned outbound go with the parent in the official `Delete` (always, or only when unused — see §S8-09.3). Do not require the researcher to delete notes before a Source. Do not CASCADE an Artifact because its Source is going.

5. **No cross-resource cascade** except the named **connection-facet** split in (17). Do not add “delete this Citation and its Observations.”

6. **How the map stays honest (S8-12).** Implement §S8-09.3 as data. Every live FK is tagged `resource` / `facet` / `optional` / **`ownedOutbound`** / **`connectionFacet`**. Every **resource** inbound edge also has a **list probe** (§S8-09.4). A `PRAGMA foreign_key_list` test fails if an FK is **unregistered**, mis-tagged, or a resource edge is missing a list probe. Soft refs in the reserved table get stub probes. `Impact.allowed` (inbound) iff a raw resource `DELETE` would succeed with FKs on, **or** the only remaining pointers are registered `connectionFacet` predicates (official `Delete` releases those first, then the parent `DELETE` succeeds).

7. **`SET NULL` is only for optional back-pointers.** Today: `sources.primary_artifact_id`. Clearing a primary thumb does not delete the Artifact.

8. **Owned outbound is how the registry deletes wrong-way children.** Each such edge names: parent table + column, child table, and `release` = `always` (exclusive facet) or `ifUnused` (pool). After the parent row is deleted, `Delete` walks those edges. `ifUnused` counts remaining inbound pointers (other Artifacts, later Claims) before dropping the child. Nested owned outbound runs next (file → `file_derivatives`). Do not invert the FK or `CASCADE` Artifact←file.

9. **Edge-lock is separate from Impact inbound.** `observations.Delete` / `Update` of an edge row still return `observations.edge_locked`. The **only** writer that may remove an edge row is official `subjects.Delete` releasing **connection facets** (17).

10. **Wrong type is erase + place.** No `UPDATE subject_type_id`. Audit records the erase.

11. **Every official resource / vocab `Delete` goes through the register.** No private `sqlInUse`. Preview is `GetDeleteImpact`; the writer re-runs `Impact` in the same tx. New tables register + follow [`add-catalog-delete`](../../../.cursor/skills/add-catalog-delete/SKILL.md) in the same PR. Facet / field writes are exempt. **Writers with no UI this spike still cut over in S8-12** (`propertyterms.Delete`).

12. **`Impact` first verifies the row exists.** Missing id → `gate = not_found`, `allowed = false`. Never treat “no inbound rows” as allowed when the parent is gone.

13. **Extra gates are not inbound counts.** `not_found`, `edge_locked`, `infra`, and `origin_locked` refuse before (or without) listing groups. `observations.Delete` order: load → `edge_locked` → `Impact` → `DELETE` parent → `ReleaseOwned` → audit / `searchindex.Delete`. **Origin lock (decided):** plugin-origin vocab is never erasable (future plugin manager). Seeded (`provenencia`) source types and source fields may erase when unused. Seeded properties and seeded / plugin terms are locked. User-origin vocab may erase when unused.

14. **Recipe owns speakable copy for `via` / `kind` / overflow.** Engine returns stable keys. Unknown `via` still lists refs + titles. Screen PRs only write target copy (“Delete SRC-…”).

15. **Rich refuse is a fetch, not a write.** The UI always calls `GetDeleteImpact` first (that RPC returns the report). Confirm calls `Delete` only when `allowed`. `Delete` still re-runs `Impact` in-tx as a safety net. If that refuses, FFI returns a **generic** `*.in_use` (or extra-gate) code — no Impact proto on `Error`, no refs in `apperr` params. A blocked write after a clear preview is a race or a client bug; toast is enough. The client may call `GetDeleteImpact` again; this spike does not require that.

16. **Domain packages register title + location projectors** for each blocker kind. `deleteimpact` composes them; it does not own kitchen-sink JOINs. S8-12 extends proto `WorkspaceLocation` to the fields Swift already has ([`add-workspace-location`](../../../.cursor/skills/add-workspace-location/SKILL.md)).

17. **Connection facets are the one predicate split.** `observations.subject_id` is a single SQLite FK. For a **bridge** subject it is two logical edges: **resource inbound** = ordinary Observations on that card (Add property extras); **`connectionFacet`** = edge-locked rows on that subject **plus** the Connect rule’s disambiguation row (`role` / `relationship_type`) on that subject. Connection facets do **not** block `Impact`. Official `subjects.Delete` releases them (notes, date/name, audit, FTS) **before** deleting the subject — the only path that may delete an edge row. Endpoints stay. The Citation stays (empty Citation is legal). `observations.value_subject_id` is always resource inbound (G2). This is not a general Observation cascade and not SQLite `CASCADE`.

The agreed table list is **§S8-09.3**. Durable contract: [`docs/catalog-deletes.md`](../../catalog-deletes.md). **S8-09** flips the FKs marked **flip** (Artifact/Citation plus `property_terms.property_id`). **S8-12** registers every row and cuts over no-UI writers. Each **screen PR** cuts over or adds that surface’s `Delete`.

**Open decisions:** none. FFI refuse carriage is (15).

### S8-09.2 Path matrix (erase vs refuse)

| ID | Path | Surface | Decision |
| --- | --- | --- | --- |
| G1 | Uncited primary (no Observation as `subject_id` or `value_subject_id`) | Graph | **Erase** + confirm |
| G2 | Primary that only appears as a bridge endpoint (`value_subject_id`) | Graph | **Refuse** (looks uncited; is not) |
| G3 | Cited primary | Graph | **Refuse** |
| G4 | Cited primary that is also an endpoint | Graph | **Refuse** |
| G5 | Uncited bridge (Connect never persists one) | Graph | **Erase** if it exists |
| G6 | Cited bridge | Graph | **Erase** + confirm when only **connection facets** remain (edges + disambiguation). **Refuse** if extra Observations sit on the bridge (Add property). Endpoints stay. |
| G7–G8 | Observation row on a card | Graph | **Out** — opens composer (C1) |
| C1–C4, C6 | Ordinary / last / subject-valued / extra-role Observation | Composer | **Erase** via Impact (inbound empty today; **S8-18** uses DeleteImpact so later claim pins refuse here) |
| C5 | Role inside a saved connection row | Composer | **Out** — no Delete on the connection row |
| C7 | Edge Observation | Composer | **Refuse** `edge_locked` (not drawn as a row) |
| C8 | Pending connection Discard | Composer | Not a catalog delete |
| C9 | Saved connection row | Composer | **Out** — no Delete on the row. Undo is graph trash on the bridge (G6). |
| C10 / C12 | Whole Citation, zero Observations | Composer | **Erase** + confirm |
| C11 | Whole Citation with any Observation | Composer | **Refuse** with Impact list |
| C13 | Subject minted from the row picker | Composer | **Out** — lives on the graph (G1) |
| S1 | Source with no Artifact and no Subject | Source page | **Erase** + confirm (**S8-14**) |
| S2 | Source with Artifact(s) and/or Subject(s) | Source page | **Refuse** with Impact list |
| A1 | Artifact with no Citation | Source page | **Erase** + confirm (file `ifUnused`) |
| A2 | Artifact with Citation(s) | Source page | **Refuse** with Impact list |
| VT1 | Unused source type | Source Types | **Erase** + confirm (**S8-15**) |
| VT2 | Source type in use | Source Types | **Refuse** — trash stays offered; `usedBy` must match Impact totals |
| VF1 | Unused metadata field | Source Fields | **Erase** + confirm (**S8-16**) |
| VF2 | Field with saved values | Source Fields | **Refuse** — trash stays offered; `usedBy` must match Impact totals |
| VP1 | Unused property | Subject Fields | **Erase** + confirm (**S8-17**) |
| VP2 | Property used on Observations | Subject Fields | **Refuse** — trash stays offered; `usedBy` counts Observations, not type-bindings |

### S8-09.4 Impact report (speakable refuse)

Callers do **not** write per-screen “what points at this?” SQL. They call `Impact(kind, id)` and render the report.

**Do not** give every parent a kitchen-sink JOIN of all child tables. The inbound **resource** edges already on the register each carry two probes:

| Probe | Purpose |
| --- | --- |
| `count` | `SELECT COUNT(*)` for that FK. Sum == 0 → erase allowed. |
| `list` | Same `WHERE`, `ORDER BY ref`, capped (20). Returns the rows the UI will name and link. |

Facets and owned outbound have **no** list probe. They never appear in the refuse.

**Report shape**

```text
Impact
  allowed          bool          // true iff the row exists, extra gates pass, and every resource-edge count is 0
  gate             enum          // ok | inbound | not_found | edge_locked | infra | origin_locked
  groups[]                       // inbound only; empty when gate ≠ inbound
    via            string        // FK identity: observations.citation_id
    kind           string        // observation | subject | citation | …
    total          int           // un-capped count (so “and 14 more” is honest)
    listed[]
      id, ref                    // machine + human id
      title                      // short label (see projectors)
      location                   // WorkspaceLocation the UI can go(to:)
```

`Impact` loads the parent first. Missing → `not_found`, not an empty allowed report. `GetDeleteImpact` on infra / skip kinds is `infra` (or invalid kind), never allowed.

`Delete` re-runs `Impact` in the same tx (safety net). The **report proto is only on `GetDeleteImpact`**. `Delete*` refuse is a generic coded error (`*.in_use`, `origin_locked`, …). Do not attach the report to `Error`. Do not stuff refs into `apperr` params. The UI must not call `Delete` when the preflight is blocked.

**Why not one generic `SELECT *`?** Policy only needs the FK. The refuse sentence needs a **ref**, a **title**, and a **place**. Those are per-*kind*, not per-parent:

| Blocker kind | Title | Location (this spike) |
| --- | --- | --- |
| Observation | Property key + short value (or the card sentence already used on the graph) | Composer for that Observation (`source` + `citation` + `subject` + `observation`) |
| Subject | Working label | Evidence graph for its Source, that card selected |
| Citation | `CIT-…` (+ Artifact ref if useful) | Composer for that Citation |
| Artifact / Source | Title / filename | Source page (Artifact later: same page, that Artifact) |
| Vocab (type, field, property, term) | Seeded / user label | Existing vocabulary place |

That title query may join a couple of columns (Observation → property key). It is a **blocker projector**, not “fetch the accompanying tables.” Do not return full Observation / Citation protobufs.

**`via` is why G2 is speakable.** A Subject has two inbound edges. `observations.subject_id` is “cites this card.” `observations.value_subject_id` is “uses this as an endpoint.” Same Observation kind, two groups, so the refuse can say *used as an endpoint in OBS-…* instead of a lying “uncited” trash.

**Links.** Reuse `WorkspaceLocation` / `go(to:)`. Observation / Citation / Subject are not omnibar kinds today — **do not** force `add-searchable-kind` to get delete links. A mapper lives next to the register (`locationFor(kind, listed row)`). The search proto `WorkspaceLocation` is still Source-page-shaped; S8-12 extends it (or a dedicated blocker location) with the composer fields Swift already has (`source_surface`, `citation_id`, `observation_id`, `subject_id`). Same destination type the graph already builds by hand.

**Cap + overflow.** List at most 20 per group; `total` is always the real count. Stable order = `ref`. Dedup inside a group by id. The same Observation must **not** be collapsed across `via`s (a row that both cites and endpoints would be two reasons).

**What the UI says.** Engine returns facts (`via`, `kind`, `ref`, `location`). **S8-13** owns L10n for group headings, overflow (“and 14 more”), and unknown-`via` fallback. Screen PRs own only the target sentence (*Delete CIT-7KD45?*). Example: *You can’t delete CIT-7KD45 because 3 Observations still belong to it* + tappable `OBS-…` rows.

When Claims ship, add an inbound resource edge + list probe on Observation. `Impact(observation, id)` then lists `CLM-…` without a new composer query.

### S8-09.3 Catalog delete register

Binding for **S8-09** and **S8-12**. Every live catalog table is here. Adding a table later means a new row in this list (and a registry probe) in the same PR.

**Legend.** *Bucket:* resource = independently addressable; facet = child→parent `CASCADE`; owned outbound = parent→child, released by official `Delete`; vocab / pool / infra as below. *Flip* = S8-09 rebuilds that FK.

#### Resources (researcher or later product delete)

| Table | Blocked by (inbound resources) | CASCADE facets | Owned outbound (registry releases) | S8-09 | S8-12 | UI this spike |
| --- | --- | --- | --- | --- | --- | --- |
| `sources` | `artifacts.source_id` (**flip**); `subjects.source_id` | notes, metadata, credibility, layout | — | Flip artifact FK | Register | **S8-14** |
| `artifacts` | `citations.artifact_id` (**flip**); `primary_artifact_id` is SET NULL | — | `file_id` → `files` (`ifUnused`, then derivatives) | Flip citation FK | Register | **S8-14** |
| `citations` | `observations.citation_id` | `citation_notes` | — | — | Register | **S8-18** |
| `observations` | later claim/narrative pins | `observation_notes` | `value_date_id` → `date_values` (`always`); `value_name_id` → `name_values` (`always`, parts CASCADE) | Unique on those two columns | Register | **S8-18** (row already S8-11) |
| `subjects` | **Resource inbound:** non-facet `observations.subject_id` (ordinary / extra rows); always `observations.value_subject_id`. **Not blockers:** `connectionFacet` rows on a bridge (edge-locked + disambiguation) — official `Delete` releases them first. | `subject_positions` | Connection facets (bridge only; not SQLite CASCADE) | — | Register split + release helper | **S8-19** |

#### Vocabulary (already `ErrInUse` in places)

| Table | Bucket | Blocked by | Facets that CASCADE | S8-09 | S8-12 | UI this spike |
| --- | --- | --- | --- | --- | --- | --- |
| `source_types` | Vocab | `sources.source_type_id` | `source_type_metadata_fields` (type side) | — | Register | **S8-15** |
| `source_metadata_fields` | Vocab | `source_metadata.field_id` | `source_type_metadata_fields` (field side); `source_metadata_layout` (field side) | — | Register | **S8-16** |
| `source_credibility_grades` | Vocab | `source_credibility_assessments.credibility_grade_id` | — | — | Register | Out |
| `subject_types` | Vocab | `subjects.subject_type_id` | `subject_type_fields` (type side) | — | Register | Out |
| `properties` | Vocab | `observations.property_id`; **`property_terms.property_id`** (terms are their own resource). `subject_type_fields.property_id` is facet CASCADE (today’s `sqlInUse` also treats bindings as in-use — **S8-17** drops that). Origin: plugin never; seeded locked; user + unused + no terms. | `subject_type_fields` (property side) | Flip term FK | Register | **S8-17** |
| `property_terms` | Vocab | `observations.value_term_id`; also **blocks the parent property** | — | Flip `property_id` → `NO ACTION` | Register + cut over `propertyterms.Delete` | Out (no Terms page this spike — a property that still has terms cannot be erased here) |

#### Facets (not `Impact` blockers; CASCADE with parent)

| Table | Parent | FK action | S8-09 | S8-12 |
| --- | --- | --- | --- | --- |
| `source_notes` | `sources` | `CASCADE` | Keep | Classify inbound only |
| `source_metadata` | `sources` | `CASCADE` on `source_id`; `NO ACTION` on `field_id` (vocab) | Keep | Classify inbound only |
| `source_credibility_assessments` | `sources` | `CASCADE` on `source_id`; `NO ACTION` on grade (vocab) | Keep | Classify inbound only |
| `source_metadata_layout` | `sources` / field | `CASCADE` on both | Keep | Classify inbound only |
| `source_type_metadata_fields` | type + field | `CASCADE` on both | Keep | Classify inbound only |
| `citation_notes` | `citations` | `CASCADE` | Keep | Classify inbound only |
| `observation_notes` | `observations` | `CASCADE` | Keep | Classify inbound only |
| `subject_positions` | `subjects` | `CASCADE` | Keep | Classify inbound only |
| `subject_type_fields` | type + property | `CASCADE` on both | Keep | Classify inbound only |

#### Owned outbound (parent points at child — registry must release)

SQLite will not delete these when the parent goes. **S8-12** stores them on the parent’s `Delete` path. `always` = exclusive (the unique indexes make a second Observation impossible). `ifUnused` = count remaining FKs to that child, including other parents of the same kind.

| Parent.column | Child | `release` | Then | S8-12 |
| --- | --- | --- | --- | --- |
| `observations.value_date_id` | `date_values` | `always` | — | Helper in **S8-12**; `observations.Delete` cut over in **S8-18** |
| `observations.value_name_id` | `name_values` | `always` | `name_value_parts` CASCADE with the name | Same |
| `artifacts.file_id` | `files` | `ifUnused` | then `file_derivatives` for that file | Later Artifact `Delete`; register now |
| `files` (after an `ifUnused` hit) | `file_derivatives` | `always` for that file | both `source_file_id` and `derived_file_id` | Nested after file release |

When Claims ship, add `reconciliation_claims.value_date_id` / `value_name_id` as owned outbound (`always`) on the claim, plus the same partial unique indexes.

#### Value rows (owned outbound of the Observation that created them)

Interpretation does not share these between Observations. Create inserts a new row; an edit `UpdateTx`s that same id. `name_value_parts` CASCADE with the name. Docs that say “shared” mean the **type** is reused by Conclusion later. Conclusion must mint **new** DateValue/NameValue rows on the claim (`structured-date-model.md` §4).

**S8-09** enforces Observation exclusivity in SQLite:

```sql
CREATE UNIQUE INDEX observations_value_date_id_uidx
  ON observations(value_date_id) WHERE value_date_id IS NOT NULL;
CREATE UNIQUE INDEX observations_value_name_id_uidx
  ON observations(value_name_id) WHERE value_name_id IS NOT NULL;
```

Partial (`WHERE IS NOT NULL`) so non-date / non-name Observations can all have NULL. Do **not** unique `value_subject_id` or `value_term_id` — many Observations may name the same person or the same term.

These are **owned outbound**, not CASCADE facets. Official Observation `Delete` must go through the registry release, not a one-off only `observations` knows about. When Claims ship, add the same partial unique + owned-outbound edges on the claim. Mint-new-row still prevents Observation and Claim from sharing one id.

| Table | Bucket | Policy | S8-09 | S8-12 |
| --- | --- | --- | --- | --- |
| `date_values` | Owned outbound of Observation (later: Claim) | Exclusive per Observation via unique index. Released `always` on Observation delete. | Unique on `observations.value_date_id` | Owned-outbound edge |
| `name_values` | Owned outbound of Observation (later: Claim) | Same. | Unique on `observations.value_name_id` | Owned-outbound edge |
| `name_value_parts` | Facet of `name_values` | `CASCADE` with the name when the name is released. | Keep | Nested after name release |

#### Pool (content-addressed; shareable)

| Table | Bucket | Policy | S8-09 | S8-12 |
| --- | --- | --- | --- | --- |
| `files` | Pool + owned outbound of Artifact | Unique `checksum_sha256`. Release `ifUnused` after Artifact delete (other `artifacts.file_id` or derivative FKs still count). | — | Owned-outbound edge on `artifacts` |
| `file_derivatives` | Owned outbound of `files` | No CASCADE today. When a file is actually released, delete derivative rows that name it (either column). Do not CASCADE Artifact←file. | — | Nested after file release |

#### Infrastructure (no researcher delete)

| Table | Bucket | Policy | S8-12 |
| --- | --- | --- | --- |
| `users` | Infra | Install/project identity. `project.updated_by` / `audit_transactions.user_id` point here. Do not offer delete in this spike. | Register so a future delete cannot skip them |
| `project` | Infra | Singleton (`id = 1`). Not deleted. | Skip `Impact`; still list in the pragma fixture |
| `audit_transactions` | Infra | Append-only history. **Never** CASCADE when a domain row is erased. | Skip |
| `audit_changes` | Facet of a revision | `NO ACTION` → transaction. Not a product delete. | Skip |
| `catalog_search_docs` | Projection | No FK. Writers / `searchindex.Delete` drop the doc. | Skip (not an FK) |
| `catalog_search_fts` / `_trigram` | Projection | External-content FTS over docs. | Skip |
| `catalog_search_meta` | Infra | Singleton projection version. | Skip |

#### Reserved (not in the catalog yet — stub probes in S8-12)

| Planned table | Bucket | Expected inbound | Action when it ships |
| --- | --- | --- | --- |
| `sameness_claims` | Resource | two `subjects` | `NO ACTION` |
| `sameness_claim_evidence` | Facet of claim **and** pin on Observation | `observation_id` → observations | **`NO ACTION` on the Observation** (claim delete may CASCADE the exhibit row) |
| `reconciliation_claims` | Resource | entity + property | `NO ACTION` |
| `reconciliation_claim_evidence` | Same as sameness evidence | `observation_id` | **`NO ACTION` on the Observation** |
| Narrative link (`target_kind` + id) | Soft ref | observation / citation / source / entity | App probe; not a real FK until designed |

---

## S8-12 — PR: Delete-impact registry

Engine-only (plus proto decode). After **S8-09**, FKs already refuse; this PR makes the refuse speakable. **No screen chrome.** Screen writers move in the matching screen PR. **No-UI writers cut over here** (`propertyterms.Delete`).

| | |
| --- | --- |
| **In** | `core/database/deleteimpact` **is §S8-09.3 + §S8-09.4 as code**. `Impact` returns the speakable report (existence + extra gates including `origin_locked` + inbound groups). Register every resource / vocab / pool **including list probes**, title/location **projectors**, owned-outbound + **connection-facet** release helpers; stub reserved pins. FFI `GetDeleteImpact` for **every** registered kind (including Out UI). Proto `WorkspaceLocation` matches Swift composer/graph fields. Cut over **no-UI** writers (`propertyterms.Delete`). Swift decode of the Impact proto (no chrome). |
| **Out** | Screen `Delete` rewrites that have a later PR; any chrome; changing FK actions; putting Impact on `Error`. |
| **Testable** | Missing id → `not_found`. Infra kind → not allowed. Plugin vocab → `origin_locked`. Seeded property → `origin_locked`. Seeded unused source type → allowed. `GetDeleteImpact` on a Citation with rows lists those `OBS-` refs + composer locations, `total` matches; >20 inbound → listed capped, `total` uncapped. G2 subject report has a `value_subject_id` group. Bridge with only connection facets → allowed (raw `DELETE` of the subject still fails until release). Bridge with an extra Observation → inbound group. Edge Observation Get + `observations.Delete` stay `edge_locked`. Pragma test fails if an FK is **unregistered** or mis-tagged, a resource edge has no list probe, or an owned-outbound FK is missing from the parent’s release list. `ifUnused` walker unit-tested. `propertyterms.Delete` uses Impact (no `sqlInUse`). Get works for subject types / grades / terms. Property with a term row → inbound `property_terms`, not allowed. |
| **Depends on** | **S8-09**. |

---

## S8-13 — PR: DeleteImpact recipe

One kit pass against [`design/S8-D9-impact.md`](design/S8-D9-impact.md).

| | |
| --- | --- |
| **In** | `DesignSystem/Recipes/DeleteImpact/` (`PV*` type): confirm branch composes `.pvConfirm`; notice branch lists groups / refs / overflow / link action. **Recipe L10n** for `via` / `kind` headings, overflow, unknown-`via` fallback, extra-gate copy (`not_found`, `edge_locked`, `infra`, `origin_locked`). Previews, VoiceOver, Swift tests. |
| **Out** | Wiring any screen; engine; target-noun copy (“Delete this Source?”). |
| **Testable** | Allowed fixture → confirm copy + destructive action; blocked fixture → no destructive button, three named rows, tap invokes `go(to:)`; overflow shows remainder; unknown `via` still lists refs; `not_found` / `edge_locked` / `origin_locked` have no destructive button; a11y names erase vs blocked. |
| **Depends on** | **S8-D9**. **S8-12** mocked or landed for the payload shape. |

---

## S8-14 — PR: Source + Artifact delete

One Source-page pass against [`design/S8-D12-source-delete.md`](design/S8-D12-source-delete.md).

| | |
| --- | --- |
| **In** | New `sources.Delete` + `artifacts.Delete` through the register (file `ifUnused`). Source page trash + DeleteImpact. Leave the page after Source erase. Metadata-row Confirm unchanged. |
| **Out** | Vocab / composer / graph; restyling the recipe. |
| **Testable** | Empty-ish Source erases (notes CASCADE); Source with Artifact or Subject: preflight notice names rows and **does not** call `Delete`; Artifact with no Citation erases; Artifact with Citations: preflight notice; shared file survives a first Artifact erase; cover Artifact erase `SET NULL`s `primary_artifact_id`. Audit + searchindex on both writers. If `Delete` is invoked while inbound (race / misuse) → generic `in_use`, no Impact on the error. After Source erase, Back to that page is `missingDeepId` (no crash). |
| **Depends on** | **S8-D12**, **S8-12**, **S8-13**. |

---

## S8-15 — PR: Source Types delete

One pass against [`design/S8-D13-source-types-delete.md`](design/S8-D13-source-types-delete.md).

| | |
| --- | --- |
| **In** | `sourcetypes.Delete` through the register. DeleteImpact on Source Types. Trash stays offered when in use. |
| **Out** | Any other page. |
| **Testable** | Unused user / seeded type erases; plugin type is `origin_locked`; in-use notice names `SRC-…`; trash is not disabled solely by `used_by`; `used_by` equals that Impact `total`. Audit + searchindex. |
| **Depends on** | **S8-D13**, **S8-12**, **S8-13**. |

---

## S8-16 — PR: Source Fields delete

One pass against [`design/S8-D14-source-fields-delete.md`](design/S8-D14-source-fields-delete.md).

| | |
| --- | --- |
| **In** | `sourcefields.Delete` through the register. DeleteImpact on Source Fields. Trash stays offered when in use. |
| **Out** | Any other page. |
| **Testable** | Unused user / seeded field erases; plugin field is `origin_locked`; in-use notice names Sources with a saved value; trash is not disabled solely by `used_by`; `used_by` equals that Impact `total`. Audit + searchindex. |
| **Depends on** | **S8-D14**, **S8-12**, **S8-13**. |

---

## S8-17 — PR: Subject Fields delete

One pass against [`design/S8-D15-subject-fields-delete.md`](design/S8-D15-subject-fields-delete.md).

| | |
| --- | --- |
| **In** | `properties.Delete` through the register. DeleteImpact on Subject Fields. Type-bindings are not blockers (**cutover:** today’s `sqlInUse` / `usedBy` count `subject_type_fields` — drop that). `usedBy` becomes Observation count. Terms **are** blockers (named `…` rows); no Terms page — that notice is honest. Origin: user + unused + no terms; seeded / plugin → `origin_locked`. |
| **Out** | Any other page; a Terms page. |
| **Testable** | Unused user property with no terms erases; in-use notice names `OBS-…`; property bound only to a type still erases; property with unused terms notices those terms; seeded property is `origin_locked`; `usedBy` matches Observation `total`. Audit + searchindex. |
| **Depends on** | **S8-D15**, **S8-12**, **S8-13**. |

---

## S8-18 — PR: Composer delete

One composer pass against [`design/S8-D11-composer-delete.md`](design/S8-D11-composer-delete.md).

| | |
| --- | --- |
| **In** | New `citations.Delete` through the register. `observations.Delete` cut over (owned outbound; **order:** load → `edge_locked` → Impact → delete → release). DeleteImpact for Citation erase and Observation row. Identity fallback after Citation erase. Composer recovers if the Artifact / Citation under it was deleted from the Source page (`missingDeepId` / identity fallback). |
| **Out** | Other screens; connection-row Delete. |
| **Testable** | Empty Citation erases; Citation with rows: preflight notice those `OBS-` refs, no `Delete` call; Observation row still erases (Impact empty); last row then Delete citation works. Edge Observation Get + Delete stay `edge_locked`. Audit + searchindex. Stale Artifact/Citation location does not crash. Mis-ordered `Delete` while inbound → generic `in_use`. |
| **Depends on** | **S8-D11**, **S8-12**, **S8-13**. |

---

## S8-19 — PR: Graph delete

One graph pass against [`design/S8-D10-graph-delete.md`](design/S8-D10-graph-delete.md).

| | |
| --- | --- |
| **In** | `subjects.Delete` through the register (release **connection facets** before the subject). Trash on cited and uncited cards; DeleteImpact (G1 / G5 / G6-facets-only confirm; G2–G4 and G6-with-extras notice). |
| **Out** | Other screens; card-level Observation trash; Change type; deleting endpoints with the bridge. |
| **Testable** | G1 still erases; G2 notice lists endpoint Observations; location / edges-only bridge erases and those edge rows are gone; endpoints remain; extra Observation on a bridge notices that `OBS-…`; `observations.Delete` on an edge still `edge_locked`. Audit + searchindex. After Subject erase, Back to that card is `missingDeepId` (no crash). |
| **Depends on** | **S8-D10**, **S8-12**, **S8-13**. |

---
## S8-99 — Dogfood close / docs

Honesty pass against the [goal bar](#goal-dogfood-bar) once the cluster is enough (or we stop adding stories). Record in [`completed.md`](completed.md); archive the spike. SemVer only if cutting a product release.

---

## Scope boundary

| In | Out |
| --- | --- |
| **Composer rethink** (Citation document; multi-subject rows; reuse; empty Save) | Four-list file browser; Option A overlay |
| **Composer + Connect simplification** (row-level Observation commits; Save citation; unsaved-work guard; one-row connection in the composer with fixed endpoints; computed bridge names; **New person / event / place from a row**; full-state audit; engine-enforced Connect edges) | Autosave; ⌘Z; app-quit / window-close guards; creating bridges or `source` subjects from the composer |
| Composer transcription assist (after **S8-10**) | Auto Observations / subjects / connect |
| Vision on **image** Artifacts | PDF OCR; audio / video OCR |
| In-memory crop from locator | Object-store crop files |
| Warn + proceed on large images | Hard reject / Apple “too many words” (does not exist) |
| PDF **Find** + **select** + **paste transcription** | PDF Vision / OCR; `text_quote` locators ([`text-quote-locators.md`](../../ideas/text-quote-locators.md)); Source-page Find; **PDF Artifact thumbnails** (glyph stays; parked in [`artifact-pdf-thumbnails.md`](../../ideas/artifact-pdf-thumbnails.md)) |
| Graph **conflict** + **negated** badges; Source-page jump; richer bridge sentences; **Add property** on bridges | Denied-line drawing; merge/resolve; **descoped** leftovers (incomplete bridges, collapse/expand, filters, undo, tray, minimap, Subject types stub, user-minted name parts) |
| Source page **Open Evidence graph** + shipped metadata + **dates as text** + **delete saved value** + **clickable `url`** (more page items via **S8-D4**) | Source-to-source commentary (`mentions` / `remark`, placeholder + merge — [`source-to-source-relationships.md`](../../ideas/source-to-source-relationships.md)); first-class Source provenance date / list sort ([`source-provenance-date.md`](../../ideas/source-provenance-date.md)); DateValue on Observations (stays); in-app browser |
| Sources list **graph-progress counts** (more list items via **S8-D5**) | Folding counts into `sourcesList`; a Subjects list destination |
| Interpretation **delete paths** (FKs **S8-09**, registry **S8-12**, recipe **S8-13**, then one screen: Source/Artifact **S8-14**, Source Types **S8-15**, Source Fields **S8-16**, Subject Fields **S8-17**, composer **S8-18**, graph **S8-19**) | Change type UI; adopt/import; undo; Citation→Observation cascade; plugin-vocab delete (future plugin manager); Terms page; Impact proto on `Error` |
| More data-entry stories as added | Remaining Spike 7 leftovers unless pulled in |

---

## Gotchas

1. **Transcription ≠ Observation** — OCR dumps into the citation reading only ([`interpretation-graph-ui.md`](../../ideas/archive/interpretation-graph-ui.md)).
2. **Vision does not refuse a newspaper page** — it usually succeeds slowly or with junk. Large-page honesty is **our** preflight (pixels / no region / post-pass observation density), not a `VNError`.
3. **Locator y-down vs Vision ROI y-up** — crop in image pixels from [`ArtifactRegionGeometry`](../../../macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift); do not pass a polygon into `regionOfInterest` (rect only).
4. **Crop the source raster**, not the zoomed viewport bitmap.
5. **No file middleman** — `ProjectFiles.objectURL` → image `NSImage` → `CGImage` → crop → `VNImageRequestHandler`.
6. **PDF is not an OCR input** — S8-01 disables Auto Transcribe. S8-05 pastes a PDFKit selection. Image-only PDFs get neither Vision nor fake text.
7. **Hide Vision behind a protocol** — `FakeStore` / unit tests inject a recognizer.
8. **S8-10, then S8-11, before any other composer chrome** — Auto Transcribe and PDF Find/paste must not land on the subject-locked form or on the S8-10 footer-Save form. Graph / Source-page / list / delete do not wait on S8-10 or S8-11. Graph PRs that touch Connect (**S8-06**) must rebase onto S8-11 if it has landed, because S8-11 removes the disambiguation sheet.
8b. **More stories do not wait on S8-01** unless they share the composer transcription chrome.
9. **PDF Artifact thumbs are not this spike** — rows keep the file-type glyph. If the idea returns, generate in Go (`core/derivatives`), not Swift/PDFKit. Parked: [`artifact-pdf-thumbnails.md`](../../ideas/artifact-pdf-thumbnails.md).
10. **S8-03 before Find/paste** — raster `displayImage` has no `PDFSelection`. Region overlay must remount with the live page or locators break.
11. **Pan vs select** — today’s unnamed click-drag pan will fight I-beam. Locked: Preview / PDFKit — drag selects, scroll pans. No hand tool or modifier-drag.
12. **Conflict is a count, not a verdict** — badge when `propertyKey` appears ≥ 2 times on that card. Values may match. Do not write a schema flag or a resolve action.
13. **Negated is polarity, not a missing line** — `polarity = negative` on the Observation. Italic-danger today is not enough; S8-D3 designs an explicit mark. Do not draw a ghost connect edge.
14. **Incomplete bridges are descoped** — Connect is atomic and the UI cannot write person-without-event. Do not add half-line chrome.
15. **Bridge nouns come from the endpoint card** — prefer that subject’s `name` / `event_type` / `toponym`, then `subjects.label`. Do not keep using only the edge row’s working-label display once a name exists.
16. **Graph → page is S8-06; page → graph is S8-07** — reuse `sourcePageLocation` and `SourcesListNavigation.graphLocation`. Same `hasArtifact` gate as the list.
17. **Collapse/expand is descoped** — do not hide the bridge sentence behind a disclosure.
18. **Bridge Add property is a new Citation** — `composerLocation(for: bridgeID)`, not `composerLocationForBridgeCitation` (that edits the connect Citation). Extra rows omit edge keys already in the sentence.
19. **Unplaced tray is descoped** — canvas create and Connect always write a position. The snapshot **omits** subjects with no row. Adopt/import (**20**) is descoped with the tray.
20. **Filters / undo / minimap are descoped** — density stays a dogfood note; undo can return if ⌘Z becomes a real pain; minimap was scope-creep.
21. **List counts are their own cache** — do not hang subject/observation numbers on `CatalogSource`. Invalidate the one `sourceId` (same pattern as `.citationCounts`). Canvas writes already name `sourceId`.
22. **Delete is exists + extra gates + inbound-*resource*-empty erase, or a refuse with a named Impact list** — see [`docs/catalog-deletes.md`](../../catalog-deletes.md) and §S8-09.1 / §S8-09.4. Facets CASCADE (notes, metadata). File behind an Artifact is owned outbound `ifUnused`. Do not CASCADE Artifact with Source or Citation with Artifact. Do not hide cited trash and let SQLite fail raw. Missing id is `not_found`, not allowed. No Change type; wrong type is delete + place.
23. **Composer document is the Citation** — `subjectId` / `citationId` on the location are focus, not locks. Do not filter observations to the entry subject. Do not keep the Artifact pre-screen.
24. **Empty Citation is a policy change** — schema allows it; Swift Save and `CreateWithObservations` (≥1) do not. S8-10 must relax the engine path.
25. **Two lists only** — Citations on this Artifact; Observations on this Citation. Subjects stay on the graph (plus a row-level picker, which since **S8-11** can create a person / event / place — never a bridge).
26. **Nothing deletes by omission** (S8-11) — an Observation is removed only by `DeleteObservation`. Do not reintroduce a "send the whole list and diff" write.
27. **First commit mints the Citation** (S8-11) — on a New Citation, row Save / Save citation / Save connection creates it from the current citation draft. After that, row commits and Save citation are independent.
28. **A connection is one unit** (S8-11) — only `CreateCitedBridge` creates edge Observations, and they are immutable from Observation write paths (`edge_locked`). The composer shows edges + role as one connection row (no row Delete). Undo is **graph trash on the bridge** (**S8-19**): connection facets (edges + disambiguation) are released with the subject; extra Add-property rows still block; endpoints and the Citation stay. A wrong endpoint on a **pending** connection = discard. The UI lock is a courtesy; the engine is the rule.
29. **Bridge names are computed, never blank** (S8-11) — show `EvidenceBridgeEdgeSummary.sentence(for:in:)`. Noun per endpoint: identity Observation (S8-06) → working label → "Type REF". Edges unreadable: stored legacy label → "Kind phrase · REF". New bridges store no label; bridge Edit sends the stored label unchanged.
30. **The server places bridges** (S8-11) — midpoint `floorDiv(a + b + 1, 2)` per axis from the endpoints' stored positions. Do not carry grid cells through `WorkspaceLocation`.
31. **Audit is lossless** (S8-11) — creates and deletes record every column (nulls included), child `date_value` / `name_value` / note rows get their own changes, and zero-change revisions are rejected. See `docs/audit-revision-history.md` §3.2.
32. **Unsaved work is guarded at navigation** (S8-11) — `WorkspaceNavigation.leaveGuard`. Places with drafts implement `WorkspaceLeaveGuard`; do not add per-button dirty checks.
33. **Catalog dates are text** (S8-07) — `source_metadata` has no `date_value_id`. Do not reopen `data_type = 'date'` so a later sort can reuse `publication_date` / `issue_date` keys; that sort is a first-class Source attribute ([`source-provenance-date.md`](../../ideas/source-provenance-date.md)). Observation DateValues stay. Do not remove `DateValueEditorForm` from the composer.
34. **Delete value ≠ dismiss suggestion** (S8-07) — saved rows call `ClearSourceMetadata`. Empty type suggestions keep the existing dismiss ×. Do not use `deleteMetadataField` (that removes vocabulary).
35. **Credibility is bottom-aligned** (S8-07) — already on the S8-D4 board; the app bug is `overviewColumns` stacking Description then Credibility with no spacer. Pin Credibility to the bottom of that left column so it tracks the metadata column’s height. Do not restyle the chips or argument.
36. **Title edit is mid-title** (S8-07) — already on the S8-D4 board; the app pins the pencil with `PVInlineEdit` `HStack(alignment: .top)` so it sits in the top-right of a display-sized title. Match the board: closer to vertical center of the title (first line if wrapped). Do not restyle the header or change metadata-row InlineEdit alignment unless a shared API can do both without moving those pencils.
37. **`url` metadata is a shape check, then a browser open** (S8-07) — Set accepts `http`/`https` with a host **or** a scheme-less host (`www.url.com`): if there is no scheme, parse `https://` + the typed string and require a host. Store what was typed. Same rule in Go `validateValue`. No network. Click opens the default browser (`https://` prefix when the saved string has no scheme). Do not add a `PVLink` primitive. Reject `file:` / `javascript:` / junk with no host.

---

## Definition of done

- [ ] Checklist stories complete (or explicitly descoped)
- [ ] Dogfood bar items for landed stories met
- [ ] Design briefs archived under `design/archive/`
- [ ] [`docs/deployment-plan/README.md`](../README.md) points at archive when closed
