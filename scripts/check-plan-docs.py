#!/usr/bin/env python3
"""Fail when a PR plan or a Claude Design brief breaks its skill's shape.

The mechanical half of review-plan (.agents/skills/review-plan). Agents
review judgment; this script checks what does not need judgment, so a
reviewer never spends a round on a missing heading or a wrong path.

PR plans (docs/deployment-plan/<spike>/prs/S#-NN-<slug>.md), per plan-pr:
  - every header field and section heading in the template is present
  - Status uses the shared status words
  - the Implementer block matches the template word for word
  - no planner comments (<!-- -->) are left
  - no soft words (consider, maybe, TBD, e.g., ...) in Scope, Decisions,
    Design, Steps, or Definition of done
  - every repo path in backticks exists, unless its line says (new)
    (not checked once the plan is Landed)
  - every relative link resolves
  - every AC is covered by a test or the manual check; every test cites
    an AC or a Decision; every STEP has a Check
  - a Planned plan has no open questions and links a review log

Briefs (docs/deployment-plan/<spike>/design/S#-D#-<slug>.md), per
add-design-brief:
  - header fields present
  - the Paste block matches add-design-brief word for word
  - the UI building-block inventory exists, its Status words are known,
    and its Home paths exist

Usage:
  python3 scripts/check-plan-docs.py [FILE ...]

With no FILE, checks every plan in prs/ and every brief in design/ (not
design/archive/) of each open spike. Exit status 1 when any error is found.
Warnings do not fail.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
PLAN_TEMPLATE = ROOT / ".agents" / "skills" / "plan-pr" / "template.md"
BRIEF_SKILL = ROOT / ".agents" / "skills" / "add-design-brief" / "SKILL.md"
DEPLOYMENT_PLAN = ROOT / "docs" / "deployment-plan"

PLAN_NAME_RE = re.compile(r"^S\d+-\d+-[a-z0-9-]+\.md$")
BRIEF_NAME_RE = re.compile(r"^S\d+-D\d+-[a-z0-9-]+\.md$")

PLAN_FIELDS = [
    "Status", "Spike / slice", "Map row", "Brief", "Board", "Depends on",
    "Blocks", "Branch", "Version bump", "Skills to load", "Rules that bind",
    "Context", "Planned", "Review",
]
PLAN_SECTIONS = [
    "### Implementer — read this first",
    "## 1. Objective",
    "## 2. User stories",
    "## 3. Scope",
    "## 4. Current state",
    "## 5. Decisions",
    "## 6. Design",
    "### Codebase fit",
    "## 7. Steps",
    "## 8. Tests",
    "## 9. Definition of done",
    "## 10. Manual check",
    "## 11. Risks and stop conditions",
    "## 12. Open questions",
    "## 13. PR",
]
STATUS_RE = re.compile(
    r"^(Draft|Planned|Blocked — .+|In progress|Landed in .+|Descoped.*)$"
)
# Sections where a soft word hides a decision.
STRICT_SECTIONS = ("## 3.", "## 5.", "## 6.", "## 7.", "## 9.")
SOFT_RE = re.compile(
    r"\b(consider|maybe|ideally|TBD|TODO|probably|as appropriate|"
    r"or similar|if needed)\b|\be\.g\.|\betc\.",
    re.IGNORECASE,
)

BRIEF_FIELDS = ["Kind", "Spike", "Implements later as", "Depends on", "Related"]
INVENTORY_STATUS = {"Ship", "Extend", "Rethink", "Remove", "New"}

PATH_ROOTS = ("core/", "api/", "macos/", "docs/", "scripts/", ".agents/",
              ".cursor/", ".github/")
# Brief inventories and many plans name UI paths from macos/App/. Paths are
# also tried from the document's folder and its spike folder.
PATH_FALLBACKS = (ROOT, ROOT / "macos" / "App")
PLACEHOLDER_RE = re.compile(r"[…<>*{}#|]|NNNNNN|S#|\.\.\.")
PATHLIKE_RE = re.compile(r"^[\w.\-/]+(\.\w+|/)$")
LINK_RE = re.compile(r"\]\(([^)\s#]+)(?:#[^)]*)?\)")
CODE_SPAN_RE = re.compile(r"`([^`]+)`")


class Report:
    def __init__(self, path: Path) -> None:
        self.path = path
        self.errors: list[str] = []
        self.warnings: list[str] = []

    def error(self, line: int, msg: str) -> None:
        self.errors.append(f"{self._where(line)}: error: {msg}")

    def warn(self, line: int, msg: str) -> None:
        self.warnings.append(f"{self._where(line)}: warning: {msg}")

    def _where(self, line: int) -> str:
        return f"{display(self.path)}:{line}" if line else display(self.path)


def display(path: Path) -> str:
    return str(path.relative_to(ROOT)) if path.is_relative_to(ROOT) else str(path)


def normalize(text: str) -> str:
    return re.sub(r"\s+", " ", text).strip()


def prose_lines(lines: list[str]):
    """Yield (lineno, text) outside fenced code and HTML comments."""
    fenced = False
    comment = False
    for i, line in enumerate(lines, 1):
        stripped = line.strip()
        if stripped.startswith("```"):
            fenced = not fenced
            continue
        if fenced:
            continue
        if comment:
            comment = "-->" not in line
            continue
        if "<!--" in line:
            comment = "-->" not in line.split("<!--", 1)[1]
            continue
        yield i, line


def section_of(lines: list[str]) -> list[str]:
    """Heading ("## n. Title") in force at each line, by index."""
    current = ""
    out = []
    for line in lines:
        if line.startswith("## "):
            current = line
        out.append(current)
    return out


def header_fields(lines: list[str]) -> dict[str, tuple[int, str]]:
    fields = {}
    for i, line in enumerate(lines, 1):
        m = re.match(r"^\*\*([^*:]+):\*\*\s*(.*)$", line)
        if m:
            fields.setdefault(m.group(1).strip(), (i, m.group(2).strip()))
        if line.startswith("### ") or line.startswith("## "):
            break
    return fields


def block_between(text: str, start: str, stop_re: str) -> str | None:
    i = text.find(start)
    if i < 0:
        return None
    rest = text[i:]
    m = re.search(stop_re, rest[len(start):], re.MULTILINE)
    return rest[: len(start) + m.start()] if m else rest


def check_paths(report: Report, lines: list[str]) -> None:
    for i, line in prose_lines(lines):
        if "(new)" in line:
            continue
        for token in CODE_SPAN_RE.findall(line):
            token = token.strip()
            if " " in token or "/" not in token or PLACEHOLDER_RE.search(token):
                continue
            if not PATHLIKE_RE.match(token):
                continue
            bases = (*PATH_FALLBACKS, report.path.parent, report.path.parent.parent)
            if any((base / token).exists() for base in bases):
                continue
            if token.startswith(PATH_ROOTS):
                report.error(i, f"path does not exist: `{token}` (mark the line (new) if it is created by this PR)")
            else:
                report.warn(i, f"path not found from the repo root or macos/App/: `{token}`")


def check_links(report: Report, lines: list[str]) -> None:
    base = report.path.parent
    for i, line in prose_lines(lines):
        for target in LINK_RE.findall(line):
            if target.startswith(("http:", "https:", "mailto:")) or PLACEHOLDER_RE.search(target):
                continue
            if not (base / target).resolve().exists():
                report.error(i, f"broken link: {target}")


def check_plan(report: Report, text: str) -> None:
    lines = text.splitlines()

    fields = header_fields(lines)
    for name in PLAN_FIELDS:
        if name not in fields:
            report.error(0, f"missing header field **{name}:**")
    status_line, status = fields.get("Status", (0, ""))
    if status and not STATUS_RE.match(status):
        report.error(status_line, f"Status is not a status word: {status!r} (see docs/deployment-plan/README.md)")

    for heading in PLAN_SECTIONS:
        if not any(line.strip() == heading for line in lines):
            report.error(0, f"missing heading: {heading}")

    template = PLAN_TEMPLATE.read_text(encoding="utf-8")
    want = block_between(template, "### Implementer — read this first", r"^---\s*$")
    got = block_between(text, "### Implementer — read this first", r"^---\s*$")
    if want and got and normalize(want) != normalize(got):
        line = next((i for i, l in enumerate(lines, 1) if l.startswith("### Implementer")), 0)
        report.error(line, "Implementer block differs from .agents/skills/plan-pr/template.md; copy it verbatim")

    for i, line in enumerate(lines, 1):
        if "<!--" in line:
            report.error(i, "planner comment left in the plan; delete <!-- --> guidance")
            break

    sections = section_of(lines)
    for i, line in prose_lines(lines):
        if not sections[i - 1].startswith(STRICT_SECTIONS):
            continue
        bare = CODE_SPAN_RE.sub("", line)
        m = SOFT_RE.search(bare)
        if m:
            report.error(i, f"soft word {m.group(0)!r} hides a decision; write the decision or a stop condition")
        if "…" in bare:
            report.warn(i, "'…' left in a binding section; a template placeholder?")

    def body(prefix: str) -> str:
        return "\n".join(l for l, s in zip(lines, sections) if s.startswith(prefix))

    stories, tests, manual = body("## 2."), body("## 8."), body("## 10.")
    for ac in sorted(set(re.findall(r"\bAC-\d+\.\d+\b", stories))):
        if not re.search(rf"\b{re.escape(ac)}\b", tests + manual):
            report.error(0, f"{ac} has no test in §8 and no manual check in §10")
    for row in tests.splitlines():
        m = re.match(r"^\|\s*(T-\d+)\s*\|", row)
        if m and not re.search(r"\b(AC-\d+\.\d+|D-\d+)\b", row):
            report.error(0, f"{m.group(1)} does not cite an AC or a Decision")

    steps = body("## 7.")
    for chunk in re.split(r"(?=^\*\*STEP-\d+)", steps, flags=re.MULTILINE):
        m = re.match(r"^\*\*(STEP-\d+)", chunk)
        if m and "Check:" not in chunk:
            report.error(0, f"{m.group(1)} has no Check:")

    if status == "Planned":
        open_q = body("## 12.").split("\n", 1)[-1].strip()
        if not open_q.startswith("None"):
            report.error(0, "status Planned with open questions in §12")
        _, review = fields.get("Review", (0, ""))
        links = LINK_RE.findall(review)
        if not links or not (report.path.parent / links[0]).exists():
            report.error(fields.get("Review", (0, ""))[0], "status Planned needs **Review:** linking the review log (review-plan)")

    map_doc = report.path.parent.parent / "deployment-plan.md"
    if map_doc.exists() and report.path.name not in map_doc.read_text(encoding="utf-8"):
        report.warn(0, f"{display(map_doc)} does not link this plan in the PR map")

    if not status.startswith("Landed in"):
        check_paths(report, lines)
    check_links(report, lines)


def check_brief(report: Report, text: str) -> None:
    lines = text.splitlines()

    fields = header_fields(lines)
    for name in BRIEF_FIELDS:
        if name not in fields:
            report.error(0, f"missing header field **{name}:**")

    skill = BRIEF_SKILL.read_text(encoding="utf-8")
    m = re.search(r"## Paste block\s*```markdown\n(.*?)\n```", skill, re.DOTALL)
    if m and normalize(m.group(1)) not in normalize(text):
        report.error(0, "Paste block differs from .agents/skills/add-design-brief/SKILL.md; copy it verbatim")

    inv = next((i for i, l in enumerate(lines) if "UI building-block inventory" in l and l.startswith("#")), None)
    if inv is None:
        report.error(0, "missing the UI building-block inventory section")
    else:
        for i in range(inv + 1, len(lines)):
            line = lines[i]
            if line.startswith("#"):
                break
            cells = [c.strip() for c in line.strip().strip("|").split("|")]
            if len(cells) < 5 or cells[0] in ("Building block", "") or set(cells[0]) <= {"-", " "}:
                continue
            m = re.match(r"[*\s]*(\w+)", cells[2])
            status = m.group(1) if m else cells[2]
            if status not in INVENTORY_STATUS:
                report.error(i + 1, f"inventory Status {status!r} is not one of {sorted(INVENTORY_STATUS)}")
            if status == "New":
                report.warn(i + 1, f"inventory marks {cells[0]!r} New; the brief must name its second call site (add-design-brief)")

    check_paths(report, lines)
    check_links(report, lines)


def default_targets() -> list[Path]:
    out: list[Path] = []
    for spike in sorted(DEPLOYMENT_PLAN.glob("spike-*")):
        out += sorted(p for p in (spike / "prs").glob("*.md") if PLAN_NAME_RE.match(p.name))
        out += sorted(p for p in (spike / "design").glob("*.md") if BRIEF_NAME_RE.match(p.name))
    return out


def main(argv: list[str]) -> int:
    targets = [Path(a).resolve() for a in argv] or default_targets()
    errors = warnings = 0
    for path in targets:
        report = Report(path)
        if not path.exists():
            report.error(0, "no such file")
        elif PLAN_NAME_RE.match(path.name) and path.parent.name == "prs":
            check_plan(report, path.read_text(encoding="utf-8"))
        elif BRIEF_NAME_RE.match(path.name):
            check_brief(report, path.read_text(encoding="utf-8"))
        else:
            report.warn(0, "not a PR plan (prs/S#-NN-slug.md) or a brief (S#-D#-slug.md); skipped")
        for msg in report.errors + report.warnings:
            print(msg)
        errors += len(report.errors)
        warnings += len(report.warnings)
    print(f"check-plan-docs: {len(targets)} file(s), {errors} error(s), {warnings} warning(s)")
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
