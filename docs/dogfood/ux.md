# UX notes

Scratch log of annoyances from dogfooding. Newest at the top of **Open**. Not a commitment and not a test plan.

## How to add

Copy the stub. Surface + what got in the way is enough. A wanted fix is optional.

```markdown
### Short title

- **Date:** YYYY-MM-DD
- **Where:** Evidence graph / citation composer / Sources list / …
- **Annoyance:** What you were doing and what felt wrong.
- **Wanted:** (optional) What would have been better.
```

## Open

<!-- add below this line -->

### Prefill the structured date modal from the transcription

- **Date:** 2026-10-07
- **Where:** Citation composer → date-type Observation → Structured date modal
- **Annoyance:** The date is usually already sitting in the Citation's transcription, but the date modal opens empty and I retype it.
- **Wanted:** When I open the date modal for a date-type Observation, prefill it with a date regex-matched from the transcription. Date formats are fairly limited, so the aim is to get 80–90% of cases right, not every one.
  - Needs a set of patterns for each supported language. That should be manageable.
  - Related: *Automatic structure from the Artifact* (field hints such as `NSDataDetector`) and *Structured date modal should open focused on Year*.
- **Several dates:** use the first match. Citations are usually narrowed to one specific part of a Source, so multiple dates in one transcription are rare. If someone cites a larger chunk, they fix the date by hand.

### Auto-fill name parts from the full name line

- **Date:** 2026-10-07
- **Where:** Citation composer → Observation name parts modal
- **Annoyance:** After typing the full name into the single full-form line, I still have to add each name part by hand.
- **Wanted:** Fill in the structured parts automatically from the full-form line. Split it on spaces and add one name part per word.
  - **Trigger:** run the split on blur, so it happens as soon as you tab out of the full-form line. Also add a **Split** button for people using the mouse.
  - **Type guessing:** it doesn't have to be perfect; handling 80–90% of names well is enough. Basic rules:
    - A comma means the surname comes first (`Smith, John Henry` → Surname, Given, Given).
    - Without a comma, the last word is the Surname and the rest are Given names.
    - Anything the rules can't place defaults to Given name.
  - I then fix the types that are wrong and I'm done. Pairs with *Name part type dropdown: shifting width and no type-to-cycle*, since that's the dropdown used for the fixes.
- **Still open:** what blur does when name parts already exist (e.g. leave them alone, or only replace parts the split created), so it never overwrites parts I edited by hand.

### Click outside a modal sheet to close it

- **Date:** 2026-10-07
- **Where:** Modal sheets (app-wide)
- **Annoyance:** Clicking outside an open modal sheet doesn't close it.
- **Wanted:** Clicking outside the sheet closes it.
- **Open question:** What happens to unsaved edits when the sheet closes this way: discard them, or ask first?

### Enter should save a Property title edit

- **Date:** 2026-10-07
- **Where:** Properties view
- **Annoyance:** Pressing Enter while editing a Property's title doesn't save it.
- **Wanted:** Enter saves the title edit.

### Rethink the Event icon

- **Date:** 2026-10-07
- **Where:** Event icon / mark (everywhere it's used)
- **Annoyance:** Events currently use an hourglass, which feels a bit odd for an Event.
- **Wanted:** A new mark, maybe a calendar or datebook. Worth doing alongside *Sidebar icons don't match Subject card icons* so the new icon gets used everywhere.

### Show sex / gender on the Person detail page

- **Date:** 2026-10-07
- **Where:** Person detail page
- **Annoyance:** The Person detail page doesn't show a Person's sex or gender visually, so you can't tell at a glance.
- **Wanted:** Use sex / gender to style the page, e.g. an accent color or a small flag/badge. It should also handle unknown or conflicting values.
  - **Seed a `gender` Property**, separate from the existing `sex_at_birth` (which is explicitly not gender identity; see [seeded vocabulary §3.7](../seeded-vocabulary.md)). It needs its own term set.
  - **Styling precedence:** use `gender` when a Person has it, and fall back to `sex_at_birth` when they don't.

### Sidebar icons don't match Subject card icons

- **Date:** 2026-10-06
- **Where:** Sidebar, Subject cards
- **Annoyance:** The icons for People, Places, and Events in the sidebar are different from the ones on their Subject cards.
- **Wanted:** Use the same icon for each type in both places.

### Open the Artifact file without expanding its row

- **Date:** 2026-10-06
- **Where:** Source detail page (Artifact list), Evidence graph
- **Annoyance:** To open a Source's Artifact file from the Source detail page, I have to expand that Artifact's row in the list to reach the open button.
- **Wanted:** Move the open button up into the collapsed row, or show a copy there, so the file opens in one click. Offer the same open action on the Evidence graph page.

### Onboarding defaults to a new user when opening an existing project

- **Date:** 2026-10-05
- **Where:** Onboarding → user selection (after choosing an existing project)
- **Annoyance:** After picking an existing project, the user selection on the next screen defaults to creating a new user.
- **Wanted:** Default to the oldest existing user in the list.

### Last updated field on list views

- **Date:** 2026-10-05
- **Where:** Sources list, People list, Events list, Places list
- **Annoyance:** None of the lists show when an item was last changed.
- **Wanted:** A **Last updated** field on the Sources list, and the same on the People, Events, and Places lists.

### Sources list sort and filter don't persist

- **Date:** 2026-10-05
- **Where:** Sources list
- **Annoyance:** The sort and filter options I pick in the Sources list don't stick.
- **Wanted:** The Sources list remembers its sort and filter settings and restores them when I come back.

### Box-select and move multiple Subject cards on the Evidence graph

- **Date:** 2026-10-04
- **Where:** Evidence graph
- **Annoyance:** Subject cards grow as you add to them and outgrow their starting positions. Keeping the layout tidy means moving them one at a time.
- **Wanted:** Hold a hotkey, then click and drag to draw a selection box. Every Subject card inside the box gets selected, and dragging any of them moves the whole group together.
  - A toolbar button also arms box-select, for people who don't know the hotkey. The button shows the hotkey (label or tooltip) so it's easy to discover.

### Structured date modal should open focused on Year

- **Date:** 2026-10-04
- **Where:** Structured date modal
- **Annoyance:** When the modal opens, focus isn't on the Year field. The year is what you usually type first, much more often than the modifiers above it.
- **Wanted:** The modal opens with focus on the Year field.

### Add Observation should focus the property search

- **Date:** 2026-10-04
- **Where:** Citation composer
- **Annoyance:** After clicking **Add Observation**, keyboard focus doesn't go to the new Observation. I have to click into the property search combo box before I can type.
- **Wanted:** Clicking **Add Observation** moves focus straight to the new Observation's property search combo box.

### Name part type dropdown: shifting width and no type-to-cycle

- **Date:** 2026-10-04
- **Where:** Citation composer → Observation name parts modal
- **Annoyance:** Two problems with the name part type dropdown:
  - **Width:** it sizes to the selected option, so it grows and shrinks as you change the selection. It looks jumpy and tense.
  - **Hotkeys:** with the dropdown focused, pressing S jumps to the first option starting with S (Surname prefix). Pressing S again stays there instead of moving to the next S option (Surname). You can't cycle through matches with the keyboard.
- **Wanted:** A fixed width (e.g. sized to the widest option). Pressing the same letter again cycles through every option that starts with it.

### Evidence graph misrenders when zoomed far out

- **Date:** 2026-10-04
- **Where:** Evidence graph
- **Annoyance:** Zooming far out breaks the rendering in two ways:
  - **Background grid:** parts of it turn solid black and the grid disappears in those areas.
  - **Lines between cards:** past a certain zoom level they start to disappear, but unevenly. Some vanish and others stay.
- **Wanted:** Zoomed-out views look refined and polished. The grid stays visible (or fades out cleanly), and lines behave the same way at every zoom level.

### Evidence graph button hit boxes are wrong

- **Date:** 2026-10-04
- **Where:** Evidence graph (toolbar, bridge cards)
- **Annoyance:** Button hit boxes still don't match the buttons. Two cases stand out:
  - **Toolbar (top left):** Add a place, Add a person, Add an event, Connect. Only the text or icon takes the click. The rest of the button does nothing.
  - **Bridge cards:** the button that opens the citation for the bridge relationship has a hit box shifted about one button-width to the left. Clicking the button does nothing; clicking the empty space to its left opens the citation.
- **Wanted:** The whole visible button accepts the click, and the hit box sits exactly on the button.

### Evidence graph viewport resets after visiting the composer

- **Date:** 2026-10-04
- **Where:** Evidence graph ↔ Citation composer
- **Annoyance:** I pan around the Evidence graph, zoom into a card, open the composer to add a property, then go back. The graph's view has been reset instead of staying where I left it, which is disorienting. This happens whether I use the back button or the breadcrumbs.
- **Wanted:** Coming back puts me at the exact same center coordinates and zoom level. Longer term, the last viewport for each Evidence graph (center + zoom) should be saved persistently per user, either in a dedicated table or by reusing wherever card coordinates are already stored.

### Omnibar ranks fuzzy name matches above exact title words

- **Date:** 2026-10-04
- **Where:** Omnibar search
- **Annoyance:** Searched `marriage` expecting a marriage certificate and a newspaper marriage announcement. Both have "marriage" in the title. They landed at the bottom of the results, under obituaries and memorials for someone named **Marion Margaret**. Looks like the shared leading "Mar…" on both names stacked up and beat the real hits. That feels like a bug: I typed the whole word, so an exact word match in a title should win.
- **Wanted:** A whole-word or exact title match always ranks above fuzzy or prefix matches. Partial matches on several tokens (Marion + Margaret) should not add up to more than one exact hit.

### Dense Evidence graphs (census-scale)

- **Date:** 2026-09-24
- **Where:** Evidence graph
- **Annoyance:** A real census page can put dozens of Subjects and hundreds of Observations on one canvas. Nothing is wrong with the model — it just gets hard to see the household you care about.
- **Wanted:** Layers or filters when a Source actually hurts. Not scheduled. Parked here until dogfood proves we need it.

### Automatic structure from the Artifact (beyond transcription)

- **Date:** 2026-09-23
- **Where:** Citation composer / Evidence graph
- **Annoyance:** Even after transcription is filled, the researcher still retypes names, dates, and roles onto cards. Spike 8 only dumps text into **transcription**.
- **Wanted:** Cheap field hints after there is a string (`NSDataDetector`, name/date patterns) and, later, guided extract. Not catalog writes.
- **Still out:** PDF OCR / Vision on page rasters; Live Text / VisionKit; Foundation Models (`macOS 26+`, raise-the-floor). Pipeline if we ever did it: string → suggestions the researcher accepts → real Citation + Observations. Experiment: “obituary → draft cards,” not “obituary → catalog writes.”

## Done

### Pulled into Spike 8 (2026-09-24)

- **Composer rethink** — Citation as the document; multi-subject rows; reuse / pinning; empty Save ([Spike 8](../deployment-plan/archive/spike-8/))
- Image **Auto Transcribe** (Vision → transcription)
- PDF **Find** + **select** + **paste transcription**
- Source page **Open Evidence graph** (was “graph starts blind”)
