#!/usr/bin/env python3
"""Fail when a skill or Cursor rule names code that no longer exists, or a kit
folder has no DesignSystem README row.

Skills and rules teach the next agent the patterns; a name they cite that the
code renamed or deleted trains it wrong (app-health audit 2026-10-27, A-6:
`eventTitlesBySource`, `PVEvidenceIcon`). This check reads every backticked
span in `.agents/skills/**/*.md` and `.cursor/rules/*.mdc`, keeps the parts
that look like code names (camelCase or PascalCase with an inner capital, or
snake_case, six characters or more), and reports each one that no Go, Swift,
proto, SQL, Python, shell or string-catalog file under core/, api/, macos/ or
scripts/ contains, either inside or as its file name. This file does not count.

A name a guide mentions on purpose (an anti-pattern to avoid, an illustrative
example, a SwiftUI or tool name) goes in scripts/check-guidance-names.allow
with the reason.

It cannot see a renamed short or generic name (`Effect.Source`): for those,
`rg` the old qualified name, as implement-pr's sweep says.

It also fails on a folder under macos/App/DesignSystem/Components/ or Recipes/
that macos/App/DesignSystem/README.md has no row for (app-health audit A-33):
a row names its folder as `Components/Name/` or `Recipes/Name/` in backticks.

    python3 scripts/check-guidance-names.py
"""
import os
import re
import sys

SELF = os.path.abspath(__file__)
ROOT = os.path.dirname(os.path.dirname(SELF))
CODE_DIRS = ("core", "api", "macos", "scripts")
CODE_EXT = (".go", ".swift", ".proto", ".sql", ".py", ".sh", ".xcstrings")
FILE_EXT = re.compile(r"\.(swift|go|md|mdc|py|sh|sql|proto|json|xcstrings|pbxproj|yml|plist)$")
SPAN = re.compile(r"`([^`\n]+)`")
NAMEISH = re.compile(r"^(?:[A-Za-z_][A-Za-z0-9_]*\.)*[A-Za-z_][A-Za-z0-9_]*(?:\(.*\))?$")
WORD = re.compile(r"[A-Za-z_][A-Za-z0-9_]*")
KIT = os.path.join(ROOT, "macos", "App", "DesignSystem")
KIT_LAYERS = ("Components", "Recipes")
APPLE = re.compile(r"^(NS|UI|CG|CF|PDF|AV|QL|VN|CA|PBX|XC)[A-Z]")


def guide_files():
    out = []
    for d, _, files in os.walk(os.path.join(ROOT, ".agents", "skills")):
        out += [os.path.join(d, f) for f in files if f.endswith(".md")]
    rules = os.path.join(ROOT, ".cursor", "rules")
    out += [os.path.join(rules, f) for f in os.listdir(rules) if f.endswith(".mdc")]
    return sorted(out)


def code_words():
    words = set()
    for top in CODE_DIRS:
        for d, _, files in os.walk(os.path.join(ROOT, top)):
            for f in files:
                path = os.path.join(d, f)
                if not f.endswith(CODE_EXT) or os.path.abspath(path) == SELF:
                    continue
                # A kit component is often named by its file (PVFormDialog.swift).
                words.add(os.path.splitext(f)[0])
                with open(path, encoding="utf-8", errors="ignore") as fh:
                    words.update(WORD.findall(fh.read()))
    return words


def allowed():
    names = set()
    path = os.path.join(ROOT, "scripts", "check-guidance-names.allow")
    with open(path, encoding="utf-8") as fh:
        for line in fh:
            name = line.split("#", 1)[0].strip()
            if name:
                names.add(name)
    return names


def looks_like_code(name):
    if len(name) < 6 or name.startswith("_") or APPLE.match(name):
        return False
    return bool(re.search(r"[a-z][A-Z]", name)) or ("_" in name.strip("_") and name.islower())


def kit_rows_missing():
    """One problem per kit folder the DesignSystem README has no row for."""
    with open(os.path.join(KIT, "README.md"), encoding="utf-8") as fh:
        readme = fh.read()
    problems = []
    for layer in KIT_LAYERS:
        base = os.path.join(KIT, layer)
        for name in sorted(os.listdir(base)):
            if os.path.isdir(os.path.join(base, name)) and f"`{layer}/{name}/" not in readme:
                problems.append(f"macos/App/DesignSystem/README.md: no row for {layer}/{name}/")
    return problems


def main():
    words, allow = code_words(), allowed()
    problems = []
    for path in guide_files():
        rel = os.path.relpath(path, ROOT)
        with open(path, encoding="utf-8") as fh:
            for lineno, line in enumerate(fh, 1):
                for span in SPAN.findall(line):
                    span = span.strip()
                    if FILE_EXT.search(span) or not NAMEISH.match(span):
                        continue
                    for name in WORD.findall(span.split("(", 1)[0]):
                        if looks_like_code(name) and name not in words and name not in allow:
                            problems.append(f"{rel}:{lineno}: `{span}` names {name}, which no code defines")
    problems += kit_rows_missing()
    if problems:
        print("\n".join(problems))
        print(f"check-guidance-names: {len(problems)} problem(s). Rename a stale name to the current code "
              "or add it to scripts/check-guidance-names.allow with the reason; "
              "give a kit folder a row in macos/App/DesignSystem/README.md.")
        return 1
    print("check-guidance-names: OK")
    return 0


if __name__ == "__main__":
    sys.exit(main())
