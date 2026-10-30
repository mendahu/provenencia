# Import Sources from genealogy registries by URL

**Status:** idea only — not roadmapped. Feasibility is unknown and needs research before any design. Captured from dogfooding (2026-10-11).

## Problem

Many genealogists now find records on large hosted services such as Ancestry, FamilySearch, MyHeritage and 23andMe. Getting one of those records into Provenencia today means creating a Source by hand and retyping what the service already shows: title, record type, repository, dates, and a link back.

## Idea

First-class support for adding a Source from a known registry:

1. The researcher copies a record URL from the service and pastes it into the app.
2. Provenencia recognizes the registry from the URL.
3. It reads the record's data and creates a Source from it, with the Source fields filled in.

This would be a convenience for the Source layer first. Whether it could also bring over an Artifact, or draft any Interpretation, is a separate and later question.

## The main uncertainty: access

Much of this content sits behind a login, so a pasted URL can't just be fetched. Options to explore:

- **Official APIs with OAuth.** Where a registry offers one, the researcher signs in once and the app holds a token. What each service actually offers, and on what terms, needs checking. I did not verify any of it.
- **A sign-in inside the app.** The researcher signs in through a web view and the app reads the page they can already see. This is likely to run into each service's terms of use.
- **Parsing the URL alone.** Even without access, a URL often carries a record or collection ID. That could create a Source with a title, a registry and a link, and leave the rest for the researcher.

## Open questions

- Which registries offer a usable API for individual researchers, and under what terms? Some may offer none.
- Do the terms of service allow reading record pages, and storing what is read?
- What would be stored for the token and where (macOS Keychain is the obvious place)?
- How are Source fields mapped from each registry's record types? Each registry may need its own adapter.
- Does an imported Source keep a structured link to the original record, so the citation can be reconstructed later?
- Does this fit alongside the existing Source types, or does a registry record need its own?
- What happens when a link later goes stale, or a service changes its pages?

## Related

- [`source-structure-later.md`](source-structure-later.md): nested Sources and derivation links (parked by Spike 12). A registry record often belongs to a larger collection; for now its details go in the Source metadata.
- [`source-provenance-date.md`](../deployment-plan/spike-13/source-provenance-date.md): first-class provenance dates on Sources.
