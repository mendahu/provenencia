# Executability lens (dry run)

The question: **can the model that will implement this plan carry it out
without making a decision the plan did not make?** Run this lens on that
model, not a stronger one. You are that implementer, rehearsing.

Do not write code. Do not edit files. Walk the plan.

## How to work

1. Read the plan top to bottom, then the Implementer block again.
2. Open every file under **Current state**. Confirm each row is still true at
   the **Planned** commit and at `main`. A row that is false is a blocking finding.
3. Load every skill under **Skills to load**.
4. Walk **Steps** in order. For each step, write down (to yourself) what you
   would do: which file, which function, what change. Then answer the questions
   below for that step.
5. Walk **Tests**: could you write each one from its row alone?
6. Walk **Definition of done**: could you show evidence for each box?

## Questions, per step

- **Choice.** Where would you have to choose — a name, a file, a type, a
  layer, an error code, a copy string, a behavior on empty or failure? Each
  choice the plan does not make is a finding. Quote the step and name the choice.
- **Find.** Is there a path, symbol, command, or key you could not find?
- **Check.** Can you run the step's check, and would it tell you the step is
  right (not only that it compiles)?
- **Order.** Does this step need something a later step creates?
- **Size.** Would you lose the thread here? A step that hides five changes is
  five steps.
- **Temptation.** What would you be tempted to add or tidy that the plan does
  not mention? If it matters, it needs an Out line or a Decision.
- **Contradiction.** Does the step disagree with a Decision, a working rule,
  a skill's "Do not", or the brief?
- **Imports.** Does the step move a type, call or interface into another
  package? Check the import graph (`go list -deps`); a cycle stops you mid-PR.
- **Locks.** Does the step take a lock, a decode slot or the catalog session
  that a caller may already hold? Say who holds it.
- **Edges.** For each loop, map, lookup or recursion the step touches, could
  you say from the plan what happens with none, one, many, a duplicate, a
  cycle, a value that contains the separator a key uses, and an input that was
  deleted or decided meanwhile? Each one you would have to guess is a finding.
- **Data.** Does a test in the step need the shipped tables or configuration
  to reach the new path? If the plan's fixture leaves them out, say so.

## Briefs

For a Claude Design brief, the "implementer" is Claude Design and then the PR
that builds the board. Ask: does each requirement say what to draw, with no
choice left that changes the inventory? Does each inventory row say which kit
piece to instance? Would a board drawn from this brief alone match the PR scope
in the map? Does each view say what its Wide pane (1600pt and up) does with the
extra space, with a `-W` requirement and a Wide frame, rather than leaving the
Regular layout jammed left?
