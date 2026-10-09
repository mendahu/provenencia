# A place to manage terms

**Status:** idea, 2026-10-09. Not scheduled. Placement (a panel on the Properties page, or its own page) is undecided.

## The problem

Term-valued Properties (sex at birth, event type, role, relationship type, place relationship type, and any term Property a researcher defines) each hold a vocabulary of terms. The app has no view of them.

- **Adding** a term happens only mid-transcription: the Citation Composer's "add custom term" dialog on an Observation row (`CitationComposerModel.confirmCustomTerm`). It asks for a label and nothing else.
- **Editing, describing, or deleting** a term isn't possible in the app. Go and the FFI already support it (`propertyterms.Update` / `Delete`, `UpdatePropertyTerm` / `DeletePropertyTerm`, with delete-impact and in-use checks), but no Swift code calls them.
- **Direction and inverses can't be set.** A relationship term that reads one way only (parent of) is `directed`, and it can name its `inverse_key` (parent ↔ child), so Promote's matching reads "Mary parent of John" and "John child of Mary" as one relationship ([`promote-graph-alignment.md`](../promote-graph-alignment.md), Direction and Inverse terms). The seeded terms have both. A term the composer creates is always undirected with no inverse, so a researcher's "godparent" matches either way round, and "godparent" never lines up with "godchild".

## The idea

One place to see and manage every term, grouped by its Property:

- List a Property's terms with origin (product, plugin, the researcher's), label, description, and how many Observations use each.
- Add, rename, describe, and delete the researcher's own terms, with the delete-impact preview other deletes use. Product and plugin terms are read-only (`propertyterms.ErrLocked`); `place_relationship_type` takes no new terms at all.
- For relationship-like Properties: mark a term as directed, and pick its inverse from the same Property (the choice is mutual: setting godparent's inverse to godchild sets godchild's too).

The composer dialog stays as it is, the quick path while transcribing. Terms made there can be refined here later. Changing a term's direction or inverse afterwards is safe for matching, which reads both from the catalog on every proposal.

## Open questions

1. **Placement:** a terms panel inside a term-valued Property on the Properties page, or a first-class Terms page that lists every term-valued Property.
2. **Which Properties offer direction and inverse:** only `relationship_type`, or any term Property (a later "contains" could be `part_of`'s inverse if place relationships ever open up).
3. **Merging terms:** two researcher terms that mean the same ("godfather", "god-father"). Rename-only, or a merge that repoints Observations.
4. **Seeded vocabulary changes** stay in `subjectvocab` (create-time Install); this view never edits product terms.

## What it would touch

- **Go:** `propertyterms.Create` / `Update` accept `directed` and `inverse_key` (validate: same Property, both directed, mutual); the FFI term messages carry both.
- **Swift:** store methods for update and delete, a term list and editor in the chosen place, the `propertyTerms` query key it reads, and L10n for the new copy.
