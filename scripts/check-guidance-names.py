#!/usr/bin/env python3
"""Fail when a skill or Cursor rule names code that no longer exists, a kit
folder has no DesignSystem README row, or an Xcode synchronized folder would
bundle a non-Swift file by accident (or lists an exception that is gone).

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

It reads macos/Provenencia.xcodeproj/project.pbxproj for the synchronized
folders Xcode builds from disk (PBXFileSystemSynchronizedRootGroup), keeping no
list of its own, and fails on: a file in one that is not Swift and not in its
membership exceptions of a target that builds the folder (Xcode would bundle it in the app), a membership
exception that is not on disk, a synchronized folder missing on disk, and a
file or folder directly under macos/App that the project neither synchronizes
nor groups (nothing would build it).

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


def _objects(text, isa):
    """(id, body) of each project object of the given isa; an object is one line or runs to a tab-indented `};`."""
    out, cur = [], None
    for line in text.split("\n"):
        if cur is None:
            m = re.match(r"\t\t([0-9A-Fa-f]{24})(?: /\*.*?\*/)? = \{", line)
            if not m:
                continue
            cur = (m.group(1), [line])
            if line.rstrip().endswith("};"):
                out.append((cur[0], line))
                cur = None
        else:
            cur[1].append(line)
            if line == "\t\t};":
                out.append((cur[0], "\n".join(cur[1])))
                cur = None
    return [(i, body) for i, body in out if re.search(r"\bisa = %s;" % isa, body)]


def _field(body, key):
    m = re.search(r"\b%s = (\"[^\"]*\"|[^;\s]+)(?: /\*.*?\*/)?;" % key, body)
    return m.group(1).strip('"') if m else None


def _list(body, key):
    """Entries of `key = ( ... );`, one per comma, without quotes, comments or blanks."""
    m = re.search(r"\b%s = \((.*?)\);" % key, body, re.S)
    if not m:
        return []
    items = (e.strip().strip('"') for e in re.sub(r"/\*.*?\*/", "", m.group(1), flags=re.S).split(","))
    return [e for e in items if e]


def _id_list(body, key):
    return [e for e in _list(body, key) if re.fullmatch(r"[0-9A-Fa-f]{24}", e)]


def synced_folders_unlisted():
    """Problems between the Xcode synchronized folders and the disk (see the docstring)."""
    with open(os.path.join(ROOT, "macos", "Provenencia.xcodeproj", "project.pbxproj"), encoding="utf-8") as fh:
        text = fh.read()
    groups = dict(_objects(text, "PBXGroup"))
    roots = dict(_objects(text, "PBXFileSystemSynchronizedRootGroup"))
    sets = dict(_objects(text, "PBXFileSystemSynchronizedBuildFileExceptionSet"))
    builders = {}
    for tid, body in _objects(text, "PBXNativeTarget"):
        for rid in _id_list(body, "fileSystemSynchronizedGroups"):
            builders.setdefault(rid, set()).add(tid)
    problems = []
    for rid, body in roots.items():
        path = _field(body, "path")
        parent = next((_field(g, "path") for g in groups.values() if rid in _id_list(g, "children")), None)
        rel = "/".join(["macos"] + ([parent] if parent else []) + [path])
        exceptions = set()
        for sid in _id_list(body, "exceptions"):
            if _field(sets.get(sid, ""), "target") in builders.get(rid, ()):
                exceptions.update(_list(sets[sid], "membershipExceptions"))
        base = os.path.join(ROOT, rel)
        if not os.path.isdir(base):
            problems.append(f"{rel}: synchronized folder missing on disk")
            continue
        found = set()
        for d, dirs, files in os.walk(base):
            dirs[:] = [x for x in dirs if not x.startswith(".")]
            for f in files:
                if f.startswith("."):
                    continue
                sub = os.path.relpath(os.path.join(d, f), base).replace(os.sep, "/")
                found.add(sub)
                if not f.endswith(".swift") and sub not in exceptions:
                    problems.append(f"{rel}/{sub}: not Swift and not in a membership exception of the {rel} synchronized folder")
        for entry in sorted(exceptions - found):
            problems.append(f"{rel}/{entry}: listed in a membership exception but not on disk")
    main_group = _field(text, "mainGroup")
    named = set()
    for child in _id_list(groups.get(main_group, ""), "children"):
        if child in groups and _field(groups[child], "path") == "App":
            for sub in _id_list(groups[child], "children"):
                body = groups.get(sub) or roots.get(sub) or ""
                named.add(_field(body, "path"))
    app = os.path.join(ROOT, "macos", "App")
    for name in sorted(os.listdir(app)):
        if not name.startswith(".") and name not in named:
            problems.append(f"macos/App/{name}: not a synchronized folder or group in the Xcode project")
    return sorted(problems)


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
    problems += synced_folders_unlisted()
    if problems:
        print("\n".join(problems))
        print(f"check-guidance-names: {len(problems)} problem(s). Rename a stale name to the current code "
              "or add it to scripts/check-guidance-names.allow with the reason; "
              "give a kit folder a row in macos/App/DesignSystem/README.md; "
              "list a non-Swift file in its synchronized folder's membership exceptions in "
              "macos/Provenencia.xcodeproj/project.pbxproj")
        return 1
    print("check-guidance-names: OK")
    return 0


if __name__ == "__main__":
    sys.exit(main())
