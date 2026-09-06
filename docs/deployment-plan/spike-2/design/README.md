# Spike 2 — Claude Design briefs

Hand these files to Claude Design **one at a time**. Each brief is self-contained: product context, numbered requirements, screen inventory, out-of-scope, and acceptance checks. You should not need the rest of the repo to produce a reviewable board.

| Step | Brief | Feeds implementation |
| --- | --- | --- |
| S2-01 | [`S2-01-workspace-chrome.md`](S2-01-workspace-chrome.md) | PR S2-14 |
| S2-02 | [`S2-02-source-catalog.md`](S2-02-source-catalog.md) | PRs S2-15, S2-16 |
| S2-03 | [`S2-03-artifacts-ingest.md`](S2-03-artifacts-ingest.md) | PR S2-17 |
| S2-04 | [`S2-04-extensible-vocabulary.md`](S2-04-extensible-vocabulary.md) | PRs S2-07 seed, S2-18 |

Sprint task list (PR sequence, non-design work): [`../README.md`](../README.md).

## How to use

1. Open the existing Provenencia Claude Design project / design-system bundle used for onboarding (see `macos/App/DesignSystem/README.md` in the repo).
2. Paste **one** brief as the prompt for a new board or flow.
3. Keep visual language aligned with onboarding (parchment neutrals, serif display, Spectral body, iron-gall accent) — do not invent a second brand.
4. Prefer existing design-system components (`Button`, `Field`, `Input`, `Select`, `Toast`, `Icon`, `LogoMark`). Add `SidebarNav` / `Card` / etc. only when the board needs them.
5. When the board is reviewable, mark the matching S2-0N step done on the task list and note any seed/vocabulary decisions in S2-04.

## Shared product facts (all briefs)

- Offline-first macOS genealogy app; no cloud account for this spike.
- After first-run onboarding, the user is “signed in” to a local `*.provenencia` project folder.
- Evidence lives in the **Source** layer; people/events/conclusions are later layers — do not design Interpretation chrome as active product UI in Spike 2.
- Short human ids (`SRC-…`, `ART-…`, `USR-…`) are shown in mono; UUIDs are not primary UI.
