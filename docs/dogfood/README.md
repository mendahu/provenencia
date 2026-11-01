# Dogfood notes

Friction found while using Provenencia on **real research**, not a planned feature spike.

[`ideas/`](../ideas/) is for unscheduled product shapes. [`deployment-plan/`](../deployment-plan/) is for committed implementation slices. This folder is the in-between: “this was annoying while I was entering data.” Adding a note does not schedule it.

## Log

| Doc | Use |
| --- | --- |
| [**UX notes**](ux.md) | Living list of UI / workflow nits. Add here first. |
| [**UX feedback entries**](ux/) | One file per dogfood finding (`ux/YYYY-MM-DD-<slug>.md`), same stub as `ux.md` plus a **Type** line (Bug / Friction / Idea). Use when an entry needs its own room or code notes. |

When a note is big enough to be its own product idea, move it to `ideas/` and delete it here. When you decide to fix a cluster, pull those rows into a spike or a small PR and delete them from this log.
