# User experience lens

The question: **will a researcher using this get it right the first time, from
the keyboard or VoiceOver as easily as the mouse, in words that make sense, and
know what to do when it fails?** Copy, accessibility, workflow, and error
handling. Runs when the diff touches `macos/App/`, `Localizable.xcstrings`, or
an error code.

You review code, not pixels. You cannot open the app. Say what the code will
do; leave how it looks to the agreed board and the user's manual check.

## Read first

- The plan's **User stories**, **Design → macOS**, **Strings**, and **Manual check**.
- The brief and its **UI building-block inventory**, if the PR is gated
  (`design/archive/S#-D#-<slug>.md`). The inventory is binding.
- [`docs/macos-client-patterns.md`](../../../../docs/macos-client-patterns.md)
  §5 accessibility, §6 localization; [`docs/design-system-layers.md`](../../../../docs/design-system-layers.md).
- [`add-localized-string`](../../add-localized-string/SKILL.md), `.cursor/rules/macos-l10n.mdc`.
- The views around the change: a new control should behave like its neighbors.

## How to work

1. Read the user stories as tasks. Walk each one through the changed views:
   what the researcher sees, where focus lands, what each key does, what happens
   on success, on empty, and on failure.
2. Do it again with only the keyboard, then as VoiceOver would announce it.
3. Read every new or changed string in context.
4. Follow every error path from the Go error code to what the researcher reads.

## Look for

### Workflow

- The task takes more steps than the board or the neighbors' pattern.
- Focus: a new form or dialog that does not put focus in its first field; Tab
  order that skips or loops; a replaced editor that drops focus (S10-16 and
  S10-17 fixed several of these).
- Return, Escape, and Tab not following the repo rule: Return saves when the
  primary action is enabled, Escape cancels, a multiline field takes Return
  (S10-08).
- A destructive action without the shared confirm (`.pvConfirm`, the
  DeleteImpact recipe); a dirty sheet that closes without asking (S10-12).
- State the researcher loses: a draft, a scroll position, a selection, after
  Back, a sheet, or a save.
- Long content: a long title, a 70-observation citation, a long transcription.
  Truncation needs a way to read the whole value (S10-42).
- An empty state that does not say what to do next.
- **Every state has a way in.** For each view state or branch the diff adds,
  name the control that reaches it in the running app. A branch only a test
  reaches (a model method called directly) is not shipped.
- **An id held across a round trip.** For each id the view keeps between
  requests (a selection, a draft, a decided row), ask what happens when
  another window deletes or merges it meanwhile. The page names the item,
  recovers, and keeps the user's other work; it does not loop on an error.

### Copy

- Hard-coded English in a view or model instead of `L10n.*`; a raw catalog key
  string at a call site; a `Text("…")` that Xcode will extract as junk.
- Words that differ from the domain (Source, Artifact, Citation, Observation,
  Subject, Person, Event, Place) or from the same word elsewhere in the app.
- A button label that does not say what it does ("OK", "Submit"); a confirm that
  does not name the thing it affects.
- Plurals and arguments: a count without a plural rule; more than one argument
  without numbered placeholders (`%1$@`). The catalog check catches some of this.
- Stale strings the change no longer shows, left in the catalog.
- **Copy that describes a behavior.** When the diff changes a default or a
  behavior, search the strings for the concept in any phrasing (the state's
  name, the default's name, "starts as", "by default"), not only for the old
  sentence, and read each hit against the new behavior.

### Accessibility

- An icon-only button or toolbar item without an accessibility label (through
  `L10n`).
- A control that is or will be UI-tested without a stable dotted
  `.accessibilityIdentifier`.
- A custom control that re-solves a kit interaction and drops its contract:
  traits, hit area, keyboard activation, focus ring. Compose the `PV*` instead.
- An action reachable only by hover, drag, or a hidden gesture.
- Informative images without a label; decorative ones that VoiceOver reads.
- Graph and canvas: new cards or controls missing from the VoiceOver rotors.

### Errors

- A swallowed error: `_ =`, an empty `catch`, an ignored `lastRefreshError`.
- Generic copy ("Something went wrong") where a stable code and a specific
  `L10n.Errors` message fit. A new `apperr` code with no Swift mapping.
- Wrong placement: an alert where inline field validation fits; a toast for a
  form mistake; a modal for a transient refresh failure.
- An error that does not say which file, field, or item failed when the code
  knows.
- Internal or debug text shown to the researcher; localization by matching
  English Go error strings.
- **Errors where the researcher acted.** An error shown far from the control
  that caused it (the top of a long scroll when the button is in a footer) and
  not announced reads as "nothing happened". Show it by the action or move
  focus or scroll to it, and announce it.
- A form field whose visible label is a separate view, so VoiceOver reads its
  value with no name.

## Allow

Non-copy data (names, paths, refs, ids), the brand name, dotted identifiers,
generated protobuf, previews. Visual taste the board already decided.

## Severity in this lens

| Severity | Use when |
| --- | --- |
| **blocker** | A user story cannot be completed; a destructive action with no confirm; data loss on a common path; a primary flow unusable from the keyboard or VoiceOver; hard-coded English shipped |
| **major** | A swallowed or generic error on a real path; a focus or key-handling break; a missing label on a primary control; copy that names the wrong concept |
| **minor** | Wording, a missing identifier on a quiet control, an awkward but working flow |
| **nit** | Taste |

## Known traps

Classes that reached `main` past this lens and were found later, by an
app-health audit or a landing. Each line names the class, the shape it took,
and what to check. Add a line when an audit traces an item back to this
lens ([`review-pr`](../SKILL.md) "Close the loop").

- **A kit promise no code keeps.** `PVConfirm` put focus on Cancel and kept
  `.keyboardShortcut(.defaultAction)` on Delete, so Return deleted while the
  README promised it could not. When a flow leans on a kit promise (Return,
  Escape, dismissal, focus, an announcement), find the code or test that
  keeps it.
- **A dismissal that skips the sheet chrome.** A link or button inside a sheet
  that sets the presenting binding to nil skips the dirty-sheet question.
  Dismiss through the chrome.
- **A brief that waives a safety rule.** A board cannot decide that a dirty
  form closes without asking, that a destructive action skips its confirm, or
  that a failure is silent (S12-D3 GF-8). Raise it against the brief.
- **Silent results.** Toasts and inline error callouts reach VoiceOver only
  if the kit announces them. A write's success or failure must be heard.
- **`try?` on a load path.** A swallowed load error that leaves a valid-looking
  empty state (a blank new Citation) invites a duplicate. Check each `try?` the
  diff adds or moves.
- **A count without a plural.** A `%lld` noun with no plural variation reads
  "1 subjects".
- **The old scope's word.** Copy that still says "this Source" after the
  action moved to one Evidence graph, and VoiceOver labels that name the
  container but not which one (the canvas, the history menu).
- **A sheet branch with no opener.** A sheet drew fields for a row kind that
  no control could open it for; the model tests called the open method
  directly, and the plan's manual check was not run. Name the control.
- **A header that kept the old default.** A behavior change swept the copy
  for one phrasing; a sentence saying the same thing in other words stayed.
- **A deleted item that blocks the page.** After another window deleted an
  item the page held, every request failed with an error naming nothing, and
  only Cancel, losing every decision, got out.

## Good finding shape

The shape to copy. Names in the example are illustrative.

"The Sources search field clears on Escape but leaves focus in the field with
the list unfiltered and the status filter still set. S10-08's rule is that
Escape cancels the edit; here it should restore the query from before the edit.
`SourcesListSearchField.onExitCommand` at line 42."
