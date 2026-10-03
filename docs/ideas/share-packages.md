# Share packages

**Status:** idea only — not roadmapped.

## Problem

Researchers often want to hand someone *just one slice* of a project: a Source they transcribed, or later a person/place they have grounded. Full-project export (or “send the whole `.provenencia` folder”) is too heavy. Copy-paste into email loses structure, files, and provenance.

## Idea

A one-button **share** action that packages a chosen subsection of the project into something:

- **Readable** by humans (notes, titles, metadata land as clear text/structure, not opaque blobs);
- **Standards-based** where practical, so non-Provenencia researchers can still use the data;
- **Self-contained** enough that the recipient can take the payload and do what they want with it — import into Provenencia later, or work outside the app entirely.

### Share a Source

Starting scope that matches today’s Source layer:

- Source title, type, description, notes
- Metadata fields / descriptive metadata
- Artifacts under that Source (including fileless placeholders)
- Associated Files (bytes + enough metadata to reopen or re-ingest)

The package should preserve enough context that a recipient understands *what was shared* without needing the rest of the project.

### Share a canonical entity (later)

Same gesture at the Conclusion layer: one button on a Person (or other identity anchor) that grabs the subsection tied to that entity. Exact contents TBD once Conclusion work exists — likely the thin canonical handle plus the Claims / supporting Interpretations / cited Sources the researcher chooses to include. Do not block Source-level sharing on this.

## Why it fits Provenencia

Local-first ownership already implies researchers should be able to extract and move their evidence. Partial share is the complementary gesture to full-project portability: publish or gift a *unit of research*, not only a whole catalog.

Recipients need not be Provenencia users. The package should remain useful as plain research material (readable notes + files + structured metadata).

## Open questions

- Package format: directory zip? single archive with a manifest? dual emit (human-readable folder + machine JSON/XML)?
- Which external standards to prefer for the machine side (GEDCOM partial export, JSON Schema of our own, both)?
- How much Interpretation / Conclusion to attach when sharing a Source (none by default vs optional “include citations that point here”)?
- Attribution: how contributor identity from audit shows up in the package without requiring cloud accounts.
- Import path: first-class “import share package” into an existing project, vs open-as-read-only, vs leave import for later.
- Privacy / redaction: notes and files may contain living-person data; any warn-before-share UX?

## Explicitly out of scope for this note

- Real-time collaboration, permissions, or hosted “share links”
- Replacing full-project backup / sync
- Speccing tables, FFI verbs, or macOS UI for Spike 2

## Related docs

- [`source-layer-data-model.md`](../source-layer-data-model.md)
- [`artifact-file-storage.md`](../artifact-file-storage.md)
- [`user-identity-model.md`](../user-identity-model.md) (attribution / future sharing boundaries)
- [`application-stack.md`](../application-stack.md) (interop and export as long-term goals)
