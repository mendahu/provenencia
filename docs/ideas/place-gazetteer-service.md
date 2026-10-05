# Place gazetteer service

**Status:** idea only, not roadmapped. The intended **end state** for place hierarchies, recorded 2026-10-05 while designing place reconciliation ([`conclusion-reconciliation.md`](../conclusion-reconciliation.md) §11). Likely a **paid service**.

## Progressive enhancement

- **The hard way (ships first, free):** the researcher builds place hierarchies by hand, like any other evidence. "Alberta is part of Canada" needs a cited Source that says so. It works, but it's slow for the common chains everyone knows (city → province → country).
- **The easy way (this idea, paid):** the app looks a place up in a Provenencia places service. It offers the place's chain and periods, and on accept writes the same evidence the hard way would: a gazetteer Source and Citation with the retrieval date, the places, and their "part of" links. Nothing in the data model changes; the service only saves typing.

Both paths produce the same rows, so a project built the easy way is still fully evidenced, portable and usable offline.

## Shape

A web API with its **own database of places and stable Provenencia place identifiers**. It sits in front of third-party gazetteers.

- **Fills itself on demand.** The app asks "do we have Toronto?". If the service doesn't, it fetches from its providers (Wikidata, GeoNames, …), stores the result, and answers. The next user gets it from the cache, so the database grows from real research instead of up-front seeding.
- **One interface, swappable providers.** The app speaks one contract. Providers can be added, swapped or merged behind it.
- **Versioned capabilities.**
  - v1: names, kind, "part of" links, and periods (the place's own and each link's).
  - v2 might add polygons.
  - Each snapshot in a project records the version it was fetched under, so the app can offer "this place was fetched under v1; v2 adds boundaries; refetch?". A refetch adds a **new Citation** rather than overwriting; the reconciler sees both.
- **Keys stay on the server.** A provider that needs a developer key or rate limits can't put that key in a Mac app.

## Before building it

- **Start in the app.** A gazetteer interface in Go core with a direct Wikidata provider proves the normalized place record. That record becomes the service's v1 contract. Swap to the service when something forces it: server-side keys, merging providers, rate limits, or polygons.
- **Two version numbers:** the API contract version (which fields exist) and each record's version (what it was fetched under).
- **Ambiguity.** A name lookup returns candidates (Toronto, Ontario vs Toronto, Ohio). The service fills only the one the researcher picks, never a guess.
- **Merging providers is hard.** Deciding that two providers' entries are the same place is its own problem. Crosswalk identifiers help (Wikidata stores GeoNames identifiers). This is where the pull toward building a historical GIS is strongest, so resist it.
- **Licensing travels with each record.**
  - Wikidata: CC0.
  - GeoNames: CC-BY, so attribution must travel with the record.
  - OpenStreetMap-derived data: ODbL, whose share-alike terms could reach the whole database.
  - FamilySearch Places: its terms may forbid caching or redistribution.

  Every record carries its provider, provider identifier and licence.
- **Operating it** means hosting, uptime, cost, security and abuse handling. The app must work fully without it.
- **Privacy:** lookups reveal what a user is researching. Keep the service opt-in and say so.

## Candidate providers

| Provider | Strength | Catch |
| --- | --- | --- |
| Wikidata | CC0; "located in" with start and end dates; many historical jurisdictions | uneven quality; online only |
| GeoNames | downloadable; strong modern administrative hierarchy | little history; CC-BY |
| FamilySearch Places | built for genealogy; historical jurisdictions | access and terms; not open data |
| Who's On First, OpenHistoricalMap, World Historical Gazetteer | open data; some historical boundaries | uneven coverage; more GIS than needed |
