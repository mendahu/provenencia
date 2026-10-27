# Spike 14 — Completed

One short paragraph per landed PR, newest first. Written by the PR that lands it.

**S14-08 — Typed omnibar hit kinds.** `CatalogSearchKind` (`macos/App/Platform/CatalogSearchKind.swift`) spells the omnibar's hit kinds once in Swift; `TestSwiftVocabularyMatchesGo/search_kinds` pins it to `core/search` `Registry`. The row, its presentation and FakeStore's search switch on it exhaustively, and handle kinds map through `EvidencePrimaryKind`. `CatalogSearchHit.kind` stays a string, so an unknown kind still shows with the Source label. No change a researcher sees.
