# Image optimization (archived)

**Done.**

## Decisions

- Do not decode `objects/…` on the SwiftUI render path. `ThumbnailCache` / `CachedThumbnail` load asynchronously; `PVThumbnail` already has a loading state.
- Object-store paths may carry a **MIME-derived extension** so Finder / Quick Look are less generic. Bytes stay content-addressed; the name is still the hash.
