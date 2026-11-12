# Review — S15-26 The metadata board takes an owner

**Verdict:** pending — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None yet.

## Writer's own calls
- The draft owner (a board on a form) moves from S15-26 to S15-10, its first caller; S15-11 reuses it. The protocol (D-4) is shaped so it conforms with `.empty` effects. Map scope entries for S15-26 and S15-10 updated.
- The section header and intro stay on the Source page; the board is rows, suggestions and the Add form (D-7), because S15-D1 SD-9 draws the Artifact board under a Disclosure.
- One Add dialog and one clear confirm per page, not per board (D-7).
- `FakeStore` keeps `metadataBySource` for Source owners and adds `metadataByOwner` for the others (D-3).
- `getMetadataBoard` waits for S15-11; renaming `SourcePageMetadata*` chrome waits for S15-09.

## Round 1

### X-1 · blocker · STEP-1's check cannot print nothing
`SourceMetadata\(` matches the kept `.mutatedSourceMetadata(` in `CatalogMutation.swift`, `CatalogQueryRegistry.swift`, `SourcePageContext.swift` and `CatalogQueryRegistryTests.swift`.
**Answer:** Fixed — STEP-1's check is now `\b(set|clear|reorder)SourceMetadata\b|dismissSourceMetadataSuggestion|SourceMetadataError` with `--glob '!*.pb.swift'`, and says `.mutatedSourceMetadata` stays.

### X-2 · major · FakeStore for non-Source owners: choices left open
`bumpActivity`, usage count, workspace read, copy and delete read only `metadataBySource`.
**Answer:** Fixed — D-3: `bumpActivity` only for `.source`; usage count, workspace read, Copy Source and delete stay Source-only; the missing-field fallback is the same for every level.

### X-3 · major · Comments and doc strings that name deleted symbols
**Answer:** Fixed — STEP-1 rewrites the store and fake doc comments for an owner (with the dismiss wording given) and the `replaceEntry` comment; STEP-4 rewrites the text editor's comment to `MetadataBoardModel.drafts`.

### X-4 · major · T-6 as written does not compile
**Answer:** Fixed — T-6 keys by `MetadataOwner(level: .artifact, id: "a1")`, seeds `fields: [authorField(), repositoryField()]` through `makeStore`, and asserts field keys and the value.

### X-5 · major · STEP-4 and DoD checks need a GUI
**Answer:** Fixed — STEP-4 adds a runnable check comparing the sorted identifier suffixes at `7d2f3e7` with the new files; seeing them in the running app stays in the manual check.

### X-6 · minor · Choices the plan leaves implicit
**Answer:** Fixed — D-7: the row views keep their names; D-5 lists `isSavingAdd` and `isClearing`, stores `private let owner`, and says `moveSaved` restores `previous` even when the owner is gone (T-3 covers it); STEP-3 says `SourcePageView` needs no edit.
