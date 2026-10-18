# Spike 10 follow-ups — PR map

PRs from the likeness and depiction area audit (2026-10-18). S10-60 – S10-63 fix the four High items; S10-64 – S10-66 fix Medium items 5, 7 and 8. None depends on another, but S10-60 and S10-66 land before S11-19 / S11-20 read the gallery and carry a depiction through the matching layers. Where two PRs edit the same file, whichever lands second merges `main` and keeps both changes.

## Working rules

- **Spike 10 decisions hold** ([`archive/spike-10/README.md`](../archive/spike-10/README.md)): a depiction is an owned value row (a region plus an image File); a gallery is the reconciled likeness values; a preferred portrait is a display pointer.
- **CDR-13 is amended (user, 2026-10-18).** Drawing a region stays a pointer action, but a likeness can take its region from the keyboard through two buttons: the whole photo or page, and the Citation's own region. No Claude Design brief: the plan describes the two buttons in words (user, 2026-10-18).

## PR map

| ID | Title | Scope | Brief | Depends on | Plan | Status |
| --- | --- | --- | --- | --- | --- | --- |
| S10-60 | The gallery reads only gallery records | `likenesses.ForEntities` admits a likeness record by its own outcome (kept, weak), so a negative, provisional, or denied record of the same picture is never a gallery record, lead, or portrait, and never hides a weak record. The gallery rule goes in the conclusion layer doc §6 (audit CON-1, CON-2, CON-3) | — | — | [plan](prs/S10-60-gallery-member-records.md) | Landed in [#459](https://github.com/mendahu/provenencia/pull/459) |
| S10-61 | Cap region size | `core/locator` refuses a locator over 256 region points (counted per document) or 64 KiB; an oversized stored region reads as hidden through the existing `Readable` rule; Swift mirrors the cap in `isValidPolygon` and stops the freeform tool at it (audit SEC-1) | — | — | [plan](prs/S10-61-region-caps.md) | Landed #462 |
| S10-62 | Refuse symlinked objects in the app | `ProjectFiles.objectURL` refuses a symlinked `objects/`, shard folder, or object, and anything that is not a regular file (audit SEC-2, first half) | — | — | [plan](prs/S10-62-object-symlinks.md) | Landed #464 |
| S10-63 | Keyboard region for a likeness | Two buttons in the Likeness value area set the region to the whole photo (“Use the whole page” on a PDF) or to the Citation’s region; a press that sets a region announces the new caption through one shared announcement helper; Draw moves focus to Cancel and announces drawing (audit UX-1) | — (waived) | — | [plan](prs/S10-63-keyboard-likeness-region.md) | Landed #467 |
| S10-64 | Check a thumbnail before decoding it | `ThumbnailCache` decodes only a PNG or JPEG whose declared pixel size is within the raster limit, else stores missing, and caps concurrent decodes (audit SEC-2, second half; Medium item 5) | — | — | [plan](prs/S10-64-thumbnail-decode-guard.md) | Planned |
| S10-65 | Render the likeness preview off the main thread | The composer cuts a likeness preview once per region, off the main actor, drawing only the face box of a PDF page; an edit with no redraw shows the saved image (audit PRF-1; Medium item 7) | — | — | [plan](prs/S10-65-depiction-preview-off-main.md) | Planned |
| S10-66 | Pure depiction value package | The depiction `Value`, its shape check and read rules move to a pure `core/depictionvalue`; `core/database/depictionvalues` keeps the queries; a matchregistry guard row, an `add-depiction-value` skill and a `depiction-values` rule (audit MNT-1, MNT-4; Medium item 8) | — | — | [plan](prs/S10-66-pure-depiction-value.md) | Planned |
