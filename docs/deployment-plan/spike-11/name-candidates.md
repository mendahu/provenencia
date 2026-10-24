# Name candidates: records Promote should have offered

**Status:** pulled into [Spike 11](deployment-plan.md) on 2026-10-24 (S11-43, S11-44, S11-45), from two dogfood cases the user raised that day.

Matching today: [`comparison-layers.md`](comparison-layers.md) (layer 1 names, layer 6 candidates), [`matching-model.md`](../../matching-model.md) (knob inventory).

## How the target menu is filled

1. **Shortlist** (`core/shortlist`): records of the row's kind that share a key with the Subject, such as a name word (`name|w:amy`) or an alias set (`name|a:…`).
2. **Candidate gate** (`graphalign.Candidates`): a shortlisted record is a candidate only if subject compare's similarity is at least `alignment.MinSimilarity` (0.5). The top `TopK` (5) are kept, plus ties up to `TieCap`.
3. **Menu:** the proposed record, then up to `alignment.Alternatives` (3) other candidates. With no candidate, a Subject with values opens New (`no_match`) and an empty menu.

## Case 1 — "Amy" against "Amy Lee Anne Long"

The record is shortlisted (they share `amy`), and the names agree: evidence +1.1, relation `Contains`. But the gate drops it. Name similarity is what the paired words earn over the weight of every word on both sides (`valuecompare/name.go` step 9; weights: surname 1.5, first given 1, other given 0.5):

| Record | Similarity | Candidate at 0.5 |
| --- | --- | --- |
| Amy | 1.00 | yes |
| Amy Long | 2 / 3.5 = 0.57 | yes |
| Amy Lee Long | 2 / 4 = 0.50 | yes (exactly at the gate) |
| Amy Lee Anne Long | 2 / 4.5 = 0.44 | **no** |
| Amy Lee Anne Marie Long | 2 / 5 = 0.40 | no |

Measured with the shipped registry on 2026-10-24. Similarity should fall as the record has more names the Subject doesn't (user, 2026-10-24). It shouldn't fall so far that a fuller record is never offered.

**Two levers:**

- **Lower the gate** (`MinSimilarity`). This only changes who becomes a candidate. Bands come from evidence, and evidence does not read the gate, so no band moves. At 0.4, a Subject's one given name reaches a record with up to three extra given names and a surname.
- **Relax the curve** (count words only one side has at less than full weight). This also raises evidence for every partial name match, because a name's points are its weight times its similarity (Amy: 2.5 × 0.44 = 1.1), so bands move and the goldens move with them.

The gate is the narrower change, so the spike ships it (S11-43) and holds the curve until dogfood shows it is still needed ([open questions](#open-questions)).

A second cause for this case: if the Source recorded "Amy" as a surname rather than a given name, it conflicts with Long (evidence −3). No gate fixes that. Check the Subject's name parts first.

## Case 2 — "Desiree" against "Des"

The record is never shortlisted. The two names share no key: the product nickname table (from upstream `carltonnorthern/nicknames`) has no Des ↔ Desiree set, and the only built-in shortening is an initial ("J." for James, `IsInitial`). The name comparer calls the first given names `Different`.

With `{desiree, des}` as a nickname set, the pair shortlists and compares at similarity 1.0 (probe, 2026-10-24). Two changes:

- **Des as a nickname of Desiree** (S11-44). The upstream list is not edited (S11-28's rule). The product adds a reviewed set of its own. Scoring reads the catalog's own alias rows, written once at create (`namealiases.Install`). `AddProductNicknames` copies new product sets into an existing catalog but has no caller today, so S11-44 also decides how an existing catalog gets them.
- **A prefix is a weaker hit** (S11-45, user 2026-10-24). A given name that is the exact start of another given name, after case and accent folding, is a hit below an alias. This is an initial generalised: an initial is `Contains` at `InitialCredit` (0.5) with `CauseInitial`; a prefix is `Contains` at a new `PrefixCredit` with a new `CausePrefix`, from `PrefixMinLength` letters. "Des" reaches "Desiree" and "Desmond" this way, and a nickname table hit still wins when there is one.

## Prefix rule, in detail

- **Words.** In `valuecompare`'s word comparer, beside the initial rule: the shorter word, at least `PrefixMinLength` letters, equals the start of the longer one after folding. Relation `Contains`, similarity `PrefixCredit`, cause `CausePrefix`. An alias hit is checked first.
- **Roles.** Given names and nicknames only. A surname prefix ("Mac" against "MacDonald") is not a hit.
- **Knobs.** `PrefixMinLength` (estimate 3: "Des", "Ben", "Kat"; two-letter forms such as "Al" or "Ed" would reach too many names) and `PrefixCredit` (estimate 0.7: above an initial, below a nickname table hit at 0.98) in the registry's value section, with inventory rows.
- **Shortlist.** A prefix has to shortlist, so a given-name word is also indexed under its prefixes from `PrefixMinLength` up (`desiree` → `des`, `desi`, `desir`, `desire`) and looked up under the word itself. S11-45 measures the extra keys on S11-21's fixture (or `bench_test.go`'s), with a cap from the measurement if needed.
- **Evidence wording.** `CausePrefix` joins the cause codes the evidence sheet words, in `ConclusionVocabulary.swift` under `swift_vocabulary_test.go`, with an L10n sentence ("Des may be short for Desiree").
- **The reconciler.** It folds on `Contains`, as with "J. Robins" into "James Robins". A prefix inside one entity would fold the same way ("Des Long" shows as "Desiree Long"); [open question 2](#open-questions).

## What stays later

- **Two-letter prefixes and surname prefixes.** Not hits.
- **A coverage rule for surname-only records** stays in [`ideas/coverage-aware-similarity.md`](../../ideas/coverage-aware-similarity.md). Lowering the gate does not change it: those records tie at similarity 1.

## Open questions

1. **The curve.** Proposed: lower the gate to 0.4 (S11-43) and leave the curve until dogfood shows a fuller record still missing. The alternative is to also count one-side-only words at a reduced weight, which raises bands and needs the goldens re-tuned.
2. **Prefix in the reconciler.** Proposed: a prefix folds like an initial within one entity. The alternative is that the reconciler ignores `CausePrefix` and keeps both spellings.
3. **How an existing catalog gets new product nicknames.** Proposed: on project open, when the product nickname list has changed since the catalog last copied it, `AddProductNicknames` runs once. The alternative is a command the researcher runs.
