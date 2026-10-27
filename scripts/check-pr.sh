#!/usr/bin/env bash
# The deterministic gate for a PR branch: every check here answers yes or no.
# The implementer runs it before saying "done"; the orchestrator runs it again
# before any reviewer starts (.agents/skills/ship-pr). CI runs the same Go and
# Xcode commands on non-draft PRs.
#
#   scripts/check-pr.sh [--base <ref>]
#
# --base   what the branch is compared against for "changed files"
#          (default: merge-base with origin/main)
#
# Checks, in order:
#   diff-check    git diff --check: whitespace errors and conflict markers
#   gofmt         gofmt -l on the Go files this branch changes
#   go-vet        go vet -tags fts5 ./api/... ./core/...
#   go-test       go test -tags fts5 ./api/... ./core/...
#   localizable   scripts/check-localizable-xcstrings.py
#   guidance-names  scripts/check-guidance-names.py: every code name a skill or
#                 Cursor rule cites still exists, and every kit Components or Recipes folder has a DesignSystem README row
#   plan-docs     scripts/check-plan-docs.py on the plans and open briefs this branch changes
#   xcode-test    xcodebuild test (ProvenenciaTests), the CI command. macOS only:
#                 anywhere else it fails, because a PR is not checked without it.
#                 The Swift files this branch changes are touched first, so an
#                 incremental build recompiles them and prints their warnings.
#   swift-warnings  no Swift compiler warning in xcode-test's build output. main
#                 builds with none, so any warning here is the branch's.
#
# Exit 0 only when every check passes. Logs for a failed check are printed;
# all logs stay in the directory named on the last line.
set -uo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

base=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --base) base="$2"; shift 2 ;;
    -h|--help) sed -n '2,30p' "$0"; exit 0 ;;
    *) echo "check-pr: unknown argument $1" >&2; exit 2 ;;
  esac
done
if [[ -z "$base" ]]; then
  base="$(git merge-base HEAD origin/main 2>/dev/null || git merge-base HEAD main)"
fi

logs="$(mktemp -d "${TMPDIR:-/tmp}/check-pr.XXXXXX")"
failed=()

# Files changed against base, committed or not, that still exist.
changed() { git diff --name-only --diff-filter=ACMR "$base" -- "$@"; }

run() {
  local name="$1"; shift
  local start=$SECONDS
  if "$@" >"$logs/$name.log" 2>&1; then
    printf 'PASS  %-14s %4ss\n' "$name" $((SECONDS - start))
  else
    printf 'FAIL  %-14s %4ss\n' "$name" $((SECONDS - start))
    tail -n 40 "$logs/$name.log" | sed 's/^/      /'
    failed+=("$name")
  fi
}

skip() { printf 'SKIP  %-14s       %s\n' "$1" "$2"; }

gofmt_changed() {
  local files bad
  files="$(changed '*.go')"
  [[ -z "$files" ]] && return 0
  # shellcheck disable=SC2086
  bad="$(gofmt -l $files)"
  [[ -z "$bad" ]] && return 0
  echo "gofmt would rewrite:"; echo "$bad"; return 1
}

plan_docs_changed() {
  local docs
  docs="$(changed 'docs/deployment-plan/*/prs/*.md' 'docs/deployment-plan/*/design/*.md' \
    | grep -v '/design/archive/' \
    | grep -E '/(S[0-9]+-[0-9]+|S[0-9]+-D[0-9]+)-[a-z0-9-]+\.md$' || true)"
  [[ -z "$docs" ]] && { echo "no plans or briefs changed"; return 0; }
  # shellcheck disable=SC2086
  python3 scripts/check-plan-docs.py $docs
}

xcode_test() {
  local files
  files="$(changed 'macos/*.swift')"
  # shellcheck disable=SC2086
  [[ -n "$files" ]] && touch $files
  NSUnbufferedIO=YES xcodebuild test \
    -project macos/Provenencia.xcodeproj \
    -scheme Provenencia \
    -destination 'platform=macOS' \
    -only-testing:ProvenenciaTests \
    CODE_SIGNING_ALLOWED=NO
}

swift_warnings() {
  local warnings
  # A warning with no source location (a key path or Sendable check the
  # compiler raises module-wide) prints as "<unknown>:0: warning:"; one inside
  # a macro expansion (#expect, @Test) names an "@__swiftmacro_….swift" file.
  warnings="$(grep -E '^([^ ]+\.swift:[0-9]+:[0-9]+|<unknown>:0): warning:' "$logs/xcode-test.log" \
    | sed "s|^$ROOT/||" | sort -u)"
  [[ -z "$warnings" ]] && return 0
  echo "Swift warnings:"
  echo "$warnings"
  return 1
}

echo "check-pr: against $(git rev-parse --short "$base")"
run diff-check  git diff --check "$base"
run gofmt       gofmt_changed
run go-vet      env CGO_ENABLED=1 go vet -tags fts5 ./api/... ./core/...
run go-test     env CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...
run localizable python3 scripts/check-localizable-xcstrings.py
run guidance-names python3 scripts/check-guidance-names.py
run plan-docs   plan_docs_changed
if [[ "$(uname -s)" == "Darwin" ]]; then
  run xcode-test xcode_test
  run swift-warnings swift_warnings
else
  skip xcode-test "not macOS"
  skip swift-warnings "not macOS"
  failed+=("xcode-test (needs macOS)")
fi

if [[ ${#failed[@]} -eq 0 ]]; then
  echo "check-pr: PASS  (logs: $logs)"
  exit 0
fi
echo "check-pr: FAIL  ${failed[*]}  (logs: $logs)"
exit 1
