# S2-03 — Artifacts, ingest, and fileless evidence

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 2 (Source layer validation)  
**Implements later as:** PR S2-17 (Swift Artifacts + ingest); informs S2-12 (thumbnails)  
**Depends on:** S2-02 Source detail (Artifacts section lives there)  
**Related briefs:** S2-01 chrome (unchanged), S2-02 catalog fields

Paste this entire document into Claude Design as the requirements for one board/flow.

---

## 1. Objective

Design how researchers **add and manage Artifacts** under a Source: fileless (physical-only) placeholders, ingesting a local file as the Artifact’s primary File, viewing file identity (name / type / size), preview or placeholder, and replacing a scan. This validates Artifact + File storage rules in the UI without building Interpretation.

---

## 2. Domain model (UI must reflect)

### 2.1 Artifact

An **Artifact** is one concrete representation of a Source (a scan, PDF, photo, or a physical object not yet digitized).

| Rule | UI implication |
| --- | --- |
| `ref` (`ART-…`) | Visible on each Artifact; monospace; system-minted; not user-typed. |
| Belongs to one Source | Artifact UI is nested under Source detail — not a global media library. |
| Zero or one **primary File** | UI is not a multi-attachment list on one Artifact. Extra scans = **extra Artifacts**. |
| `file_id` may be null | **Fileless** Artifact is first-class (e.g. “Bible held by Mary Smith”). |
| `description` | Optional short text about this representation (e.g. “recto”, “archival TIFF”). |
| No `artifact_type` taxonomy | Do not invent “scan / photo / pdf” enum chips — MIME/type comes from the File when present. |

### 2.2 File (storage object)

| Rule | UI implication |
| --- | --- |
| Immutable bytes, content-addressed | “Replace file” creates/points to a **new** File; do not imply in-place edit of pixels. |
| `original_filename`, `media_type`, `byte_size` | Show these when a File is attached. |
| Storage path derived from SHA-256 under `objects/` | Do **not** show raw hash paths as the primary label; filename is human-facing. Hash may be advanced/detail only. |
| Bytes never travel in API payloads as design fiction | Ingest is “choose a file on disk”; progress may be shown, but no drag-to-base64 narrative. |
| Primary File deletion unsupported initially | No “Delete file from disk” destructive control in Spike 2. Detach/replace only as product allows — prefer **replace**; omitting detach is OK if replace + fileless cover the stories. |

### 2.3 Derivatives

| Rule | UI implication |
| --- | --- |
| Thumbnails belong to **Files**, not Artifacts | Preview image is a convenience; missing thumbnail → placeholder by media type. |
| Disposable / regenerable | No “manage derivatives” UI. |
| Non-images | PDF/audio may use generic placeholders in Spike 2. |

### 2.4 Mental model (must be obvious in UI)

```text
Source (SRC-…)
  ├── Artifact ART-…  →  File (scan.tiff)     ← representation A
  ├── Artifact ART-…  →  File (scan.jpg)      ← representation B (separate Artifact)
  └── Artifact ART-…  →  (no file)            ← physical-only
```

Wrong model to avoid: one Artifact with many attached files.

---

## 3. Requirements

### 3.1 Artifact list on Source detail

| ID | Requirement |
| --- | --- |
| A-1 | Show a list/grid of Artifacts for the current Source. |
| A-2 | Each item shows **`ART-…`**, optional description, and File summary or “No file” / physical-only treatment. |
| A-3 | Empty state explains representations vs the Source itself in one short sentence. |
| A-4 | Primary actions: **Add without file** and **Add from file…** (labels flexible; both intents required). |

### 3.2 Fileless Artifact

| ID | Requirement |
| --- | --- |
| A-5 | Creating a fileless Artifact requires minimal friction (optional description only). |
| A-6 | Fileless state is a valid resting state — not an error or incomplete warning that blocks save. |
| A-7 | Copy may mention physical custody; do not require repository metadata here (that’s Source metadata). |

### 3.3 Ingest from disk

| ID | Requirement |
| --- | --- |
| A-8 | “Add from file” uses a native file chooser pattern (macOS open panel). Design the **before / choosing / after** states. |
| A-9 | After ingest, show **original filename**, **media type**, and **size** on the Artifact. |
| A-10 | Show **preview** for common images when available; otherwise a media-type placeholder. |
| A-11 | Ingest failure (missing file, unsupported, permission) uses calm error feedback consistent with onboarding toasts. |
| A-12 | Do not imply the app uploads to a cloud; wording is local project ingest. |

### 3.4 Replace primary File

| ID | Requirement |
| --- | --- |
| A-13 | An Artifact with a File supports **Replace file…**. |
| A-14 | UX copy/treatment should not say the old file is destroyed; historical retention is a product rule (old bytes kept). A light confirmation is enough — no need for a version browser in Spike 2. |
| A-15 | After replace, UI shows the new filename/type/size/preview. |

### 3.5 Multiple representations

| ID | Requirement |
| --- | --- |
| A-16 | Board includes a frame with **two Artifacts** under one Source (e.g. TIFF + JPEG) to teach the model. |
| A-17 | No control that attaches a second File to the same Artifact. |

### 3.6 Layout

| ID | Requirement |
| --- | --- |
| A-18 | Remains inside Source detail within the S2-01 workspace content host. |
| A-19 | Does not open a separate “media manager” window as the primary pattern. |

---

## 4. Screen / frame inventory (minimum)

1. Source detail — Artifacts **empty**.
2. Add **fileless** Artifact (flow + resulting row).
3. **Ingest** image — success with preview + filename/type/size.
4. Ingest **non-image** (e.g. PDF) — placeholder treatment.
5. **Replace file** — before/after.
6. Source with **two** file-backed Artifacts side by side in the list.
7. Optional: error toast on failed ingest.

---

## 5. Out of scope

- Editing Source title/type/metadata (S2-02).
- Custom vocabulary (S2-04).
- Citation cropping / transcription (Interpretation).
- Derivative management, waveform UI, PDF page scrubber.
- Historical File version timeline UI.
- Deleting primary Files from disk.
- Drag-import from cloud providers as a special case.

---

## 6. Decisions to record on the board

- Thumbnail: required for dogfood vs placeholder-only (feeds engineering S2-12).
- List vs grid for Artifacts.
- Whether replace needs an explicit confirm dialog.

---

## 7. Acceptance checklist

- [ ] Fileless and file-backed Artifacts are both first-class.
- [ ] Multiple files ⇒ multiple Artifacts (demonstrated).
- [ ] `ART-…` visible; filenames are human-facing.
- [ ] Replace does not read as destructive wipe of history.
- [ ] Preview/placeholder rules are clear for image vs other.
- [ ] No cloud-upload language; no multi-file-per-Artifact control.
- [ ] Nested under Source detail in workspace chrome.
