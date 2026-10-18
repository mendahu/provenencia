# Spike 10 follow-ups — PR map

Four independent PRs from the likeness and depiction area audit (2026-10-18). Each fixes one High item. None depends on another, and none blocks Spike 11, but S10-60 lands before S11-19 / S11-20 read the gallery.

## Working rules

- **Spike 10 decisions hold** ([`archive/spike-10/README.md`](../archive/spike-10/README.md)): a depiction is an owned value row (a region plus an image File); a gallery is the reconciled likeness values; a preferred portrait is a display pointer.
- **CDR-13 is amended (user, 2026-10-18).** Drawing a region stays a pointer action, but a likeness can take its region from the keyboard through two buttons: the whole photo or page, and the Citation's own region. No Claude Design brief: the plan describes the two buttons in words (user, 2026-10-18).

## PR map

| ID | Title | Scope | Brief | Depends on | Plan | Status |
| --- | --- | --- | --- | --- | --- | --- |
| S10-60 | The gallery reads only gallery records | `likenesses.ForEntities` admits a likeness record by its own outcome (kept, weak), so a negative, provisional, or denied record of the same picture is never a gallery record, lead, or portrait, and never hides a weak record. The gallery rule goes in the conclusion layer doc §6 (audit CON-1, CON-2, CON-3) | — | — | [plan](prs/S10-60-gallery-member-records.md) | Draft |
| S10-61 | Cap region size | `core/locator` refuses a locator over 256 region points (counted per document) or 64 KiB; an oversized stored region reads as hidden through the existing `Readable` rule; Swift mirrors the cap in `isValidPolygon` and stops the freeform tool at it (audit SEC-1) | — | — | [plan](prs/S10-61-region-caps.md) | Planned |
| S10-62 | Refuse symlinked objects in the app | `ProjectFiles.objectURL` refuses a symlinked `objects/`, shard folder, or object, and anything that is not a regular file (audit SEC-2, first half) | — | — | [plan](prs/S10-62-object-symlinks.md) | Draft |
| S10-63 | Keyboard region for a likeness | Two buttons in the Likeness value area set the region to the whole photo or page or to the Citation's region; Draw moves focus to Cancel and announces drawing (audit UX-1) | — (waived) | — | [plan](prs/S10-63-keyboard-likeness-region.md) | Draft |
