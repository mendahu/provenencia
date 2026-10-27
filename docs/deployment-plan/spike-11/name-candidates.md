# Name candidates: records Promote should have offered

**Status:** pulled into [Spike 11](deployment-plan.md) on 2026-10-24 (S11-43, S11-44, S11-46, S11-45, S11-47), from two dogfood cases the user raised that day.

Matching today: [`comparison-layers.md`](comparison-layers.md) (layer 1 names, layer 6 candidates), [`matching-model.md`](../../matching-model.md) (knob inventory).

## How the target menu is filled

1. **Shortlist** (`core/shortlist`): records of the row's kind that share a key with the Subject, such as a name word (`name|w:amy`) or an alias set (`name|a:…`).
2. **Candidate gate** (`graphalign.Candidates`): a shortlisted record is a candidate only if subject compare's similarity is at least `alignment.MinSimilarity` (0.4 since S11-43; 0.5 before). The top `TopK` (5) are kept, plus ties up to `TieCap`.
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

- **Lower the gate** (`MinSimilarity`). This only changes who becomes a candidate. Bands come from evidence, and evidence does not read the gate, so no band moves. The row still opens unset, now with the record in its menu ([Promote suggests a record or nothing](#promote-suggests-a-record-or-nothing)). No test or golden flips at 0.4 (map review, 2026-10-24). At 0.4, a Subject's one given name reaches a record with up to three extra given names and a surname.
- **Relax the curve** (count words only one side has at less than full weight). This also raises evidence for every partial name match, because a name's points are its weight times its similarity (Amy: 2.5 × 0.44 = 1.1), so bands move and the goldens move with them.

The gate is the narrower change, so the spike ships it (S11-43) and holds the curve until dogfood shows it is still needed (user, 2026-10-24).

A second cause for this case: if the Source recorded "Amy" as a surname rather than a given name, it conflicts with Long (evidence −3). No gate fixes that. Check the Subject's name parts first.

## Case 2 — "Desiree" against "Des"

The record is never shortlisted. The two names share no key: the product nickname table (from upstream `carltonnorthern/nicknames`) has no Des ↔ Desiree set, and the only built-in shortening is an initial ("J." for James, `IsInitial`). The name comparer calls the first given names `Different`.

With `{desiree, des}` as a nickname set, the pair shortlists and compares at similarity 1.0 (probe, 2026-10-24). Two changes:

- **Des as a nickname of Desiree and Desirae** (S11-44, user 2026-10-24). Two sets, `{desiree, des}` and `{desirae, des}` (a nickname shared by two formal names does not join their sets, the table's rule), join `nicknameSets`, the product's nickname table, which is already a reviewed edit of the upstream list (its header; S11-28: "another found later is a one-line seed edit plus a CacheVersion bump"). Scoring reads the catalog's own alias rows, written once at create (`namealiases.Install`). `AddProductNicknames` copies new product sets into an existing catalog, reading the `nicknames` table, but has no caller today. An existing catalog gets them by a command the researcher runs (user, 2026-10-24): the tables are catalog data, so a code change alone reaches only new catalogs.
- **A prefix is a weaker hit** (S11-45, user 2026-10-24). A given name that is the exact start of another given name, after case and accent folding, is a hit below an alias. This is an initial generalised: an initial is `Contains` at `InitialCredit` (0.5) with `CauseInitial`; a prefix is `Contains` at a new `PrefixCredit` with a new `CausePrefix`, from `PrefixMinLength` letters (2, user 2026-10-24). "Des" reaches "Desiree" and "Desmond" this way, and "Al" reaches "Albert" and "Alice", and a nickname table hit still wins when there is one.

## Prefix rule, in detail

- **Words.** In `valuecompare`'s word comparer (`words.go`, today: exact, initial, alias, then fold and spelling): after alias, fold and spelling, so "Ann" against "Anne" stays a spelling `Near` and the ladders hold. The shorter word, at least `PrefixMinLength` letters, equals the start of the longer one after folding. Relation `Contains`, similarity `PrefixCredit`, cause `CausePrefix`. The broader side is the shorter word (today `name.go` picks it with `IsInitial`, which is wrong for a prefix on side A). A `prefixes` flag on the word options turns it on for names; text compare never gets it.
- **Roles.** The alias tables' role set (`namealiases.aliasPartRoles`): given, nickname and untyped parts. A surname prefix ("Mac" against "MacDonald") is not a hit.
- **Knobs.** `PrefixMinLength` (2, user 2026-10-24: "Al", "Ed" and "Jo" count; S11-46's measurement says what the wider shortlist costs) and `PrefixCredit` (estimate 0.7: above an initial, below a nickname table hit at 0.98) in the registry's value section, with inventory rows.
- **Shortlist** (S11-46). A prefix has to shortlist in both directions: a Subject "Des" must reach a record "Desiree", and a Subject "Desiree" a record "Des". So a record files its whole word and its prefixes from `PrefixMinLength` up under separate tags, and a Subject looks up its whole word against filed prefixes and its own prefixes against filed whole words. Today `valuecompare.WordKeys` returns one list that `shortlist.Keys` puts on both sides; S11-46 splits it into index and lookup lists first. It measures the extra keys on S11-21's fixture (or `bench_test.go`'s), with a cap from the measurement if needed.
- **Evidence wording.** `CausePrefix` joins the cause codes the evidence sheet words, in `ConclusionVocabulary.swift` under `swift_vocabulary_test.go`, with an L10n sentence ("Des may be short for Desiree").
- **The reconciler.** Its name fold (`autoreconcile/names.go`) allows only `Same` and `CauseInitial` pairs (S11-28 widens it to aliases), so a prefix does not fold unless `CausePrefix` joins that allow-list; [open question 2](#open-questions). S11-45 lands after S11-28, which edits the same line.

## What stays later

- **Surname prefixes.** Not hits.
- **Relaxing the name curve** (counting one-side-only words at less than full weight). Held until dogfood after S11-43 shows a fuller record still missing (user, 2026-10-24).
- **A coverage rule for surname-only records** stays in [`ideas/coverage-aware-similarity.md`](../../ideas/coverage-aware-similarity.md). Lowering the gate does not change it: those records tie at similarity 1.

## Decided (user, 2026-10-24)

1. **Gate only.** Lower the gate to 0.4 (S11-43); leave the curve alone for now.
2. **A prefix folds like an initial** in the reconciler: `CausePrefix` joins the fold allow-list, so "Des Long" shows as "Desiree Long" within one entity (S11-45).
3. **New product nicknames reach an existing catalog by a command.** No backfill on open, so `add-seeded-vocabulary`'s rule stands and no marker is stored (S11-44).
4. **Prefixes from two letters** (`PrefixMinLength` 2).

## Promote suggests a record or nothing

Go today gives every row it does not place a target: Skip (`weak`) when a candidate fell short, New (`no_match`) when there was none (`graphalign/rows.go`). Swift drafts neither. A row with no band opens unset (`PromoteFlow.draftTarget`), and the Skip / New it receives survive only in a rare fallback (an unplaced row with a band that was not taken falls through to Skip). The user's call (2026-10-24): remove it (S11-47). A proposal row carries a handle or no target; its alternatives stay. New and Skip are only ever the researcher's choice. A decided New or Skip still goes to Go as a fixed pair (`PromoteGraphAlignmentFixed`), unchanged.

So lowering the gate does not change how a row opens: "Amy" still opens unset, now with "Amy Lee Anne Long" in its menu.

Go's possible-duplicate check (`rows.go`, two look-alike rows that would both be new records) finds its rows by `TargetNew` today; S11-47 finds them as unplaced rows instead, so the "also new" note still works.
