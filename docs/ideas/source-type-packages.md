# Source type packages

**Status:** idea, 2026-11-09. Not scheduled. Raised while planning Spike 15, which seeds the *Evidence Explained*-style types, fields and citation formats as the product default instead ([`spike-15/evidence-explained-citation-map.md`](../deployment-plan/spike-15/evidence-explained-citation-map.md) § 6.8).

## The idea

A seeded Source type is only data: a name, its suggested metadata fields at each level, and its three citation formats. A set of them is a citation style. So a style could be an installable **package**: the core app owns the mechanism (metadata, levels, formats, the engine), and the *Evidence Explained*-style set becomes one package among others (another citation style, another language, a researcher's own types exported and shared). Vocabulary rows already carry `origin`, which allows `plugin:<id>` ([`seeded-vocabulary.md`](../seeded-vocabulary.md) § 1.1).

## Holes to close first

- **Fields are shared, types are not.** Two packages that both define `author` would make two "Author" fields under different origins, and values typed into one are invisible to the other's formats. Likely answer: fields match by key on install and form one pool; packages own types and formats.
- **A type mixes "what kind of record" with "which style".** A Source has one type, so switching styles means re-typing Sources, and one census cannot have two styles. Likely answer, if it ever matters: a package can add formats to an existing type.
- **Lifecycle.** Updating a format the researcher edited, and uninstalling types that Sources still use (a type in use cannot be deleted). Simplest honest version: install copies the data in, and from then on it is the researcher's, with no updates or uninstall.
- **Installing into an existing project** is the same operation Spike 15 descoped (§ 6.4 of the map).
- **Naming and trademark.** An installable "Evidence Explained" package looks like an official product. Name it generically, or get the publisher's permission (Spike 15 working rule, [`spike-15/README.md`](../deployment-plan/spike-15/README.md#working-rules)).
- **First run.** A project with no package has no types; the default package should be on unless the researcher opts out.

## Upsides

- A package is a data file (fields, types, suggestions, formats), so the same format gives "export my types" and sharing between researchers.
- Formats hold English words ("census", "dwelling", "accessed"), so a package is the natural way to ship another language.
- The style-specific content lives in one labelled place, which helps an IP review.
