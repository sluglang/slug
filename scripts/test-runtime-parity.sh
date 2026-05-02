#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

report="$TMP_DIR/parity_report.txt"
: > "$report"

normalize_streams() {
  local stdout_file="$1"
  local stderr_file="$2"
  local out_file="$3"
  {
    cat "$stdout_file"
    cat "$stderr_file"
  } | sed -E 's/\x1B\[[0-9;]*[A-Za-z]//g; s/[[:space:]]+$//' > "$out_file"
}

check_file() {
  local root="$1"
  local file="$2"

  local tw_out tw_err vm_out vm_err tw_norm vm_norm
  tw_out="$(mktemp "$TMP_DIR/tw_out.XXXXXX")"
  tw_err="$(mktemp "$TMP_DIR/tw_err.XXXXXX")"
  vm_out="$(mktemp "$TMP_DIR/vm_out.XXXXXX")"
  vm_err="$(mktemp "$TMP_DIR/vm_err.XXXXXX")"
  tw_norm="$(mktemp "$TMP_DIR/tw_norm.XXXXXX")"
  vm_norm="$(mktemp "$TMP_DIR/vm_norm.XXXXXX")"

  local tw_ec vm_ec
  set +e
  SLUG_HOME="$ROOT_DIR" go run ./cmd/app/main.go --runtime treewalk -log-level error --root "$root" "$file" >"$tw_out" 2>"$tw_err"
  tw_ec=$?
  SLUG_HOME="$ROOT_DIR" go run ./cmd/app/main.go --runtime vm -log-level error --root "$root" "$file" >"$vm_out" 2>"$vm_err"
  vm_ec=$?
  set -e

  if [[ "$tw_ec" -ne "$vm_ec" ]]; then
    {
      echo "EXIT_MISMATCH|$root|$file|treewalk=$tw_ec|vm=$vm_ec"
      echo "--- treewalk stderr ---"
      sed -n '1,20p' "$tw_err"
      echo "--- vm stderr ---"
      sed -n '1,20p' "$vm_err"
    } >> "$report"
    return
  fi

  normalize_streams "$tw_out" "$tw_err" "$tw_norm"
  normalize_streams "$vm_out" "$vm_err" "$vm_norm"
  if ! diff -u "$tw_norm" "$vm_norm" > "$TMP_DIR/diff.tmp" 2>/dev/null; then
    {
      echo "OUTPUT_MISMATCH|$root|$file|treewalk=$tw_ec|vm=$vm_ec"
      echo "--- treewalk stderr ---"
      sed -n '1,20p' "$tw_err"
      echo "--- vm stderr ---"
      sed -n '1,20p' "$vm_err"
      echo "--- treewalk stdout ---"
      sed -n '1,20p' "$tw_out"
      echo "--- vm stdout ---"
      sed -n '1,20p' "$vm_out"
    } >> "$report"
  fi
}

echo "Running VM conformance suite..."
go test ./internal/runtime -run 'TestVMConformanceFixtures|TestVMKnownUnsupportedFixtures|TestVMConformanceErrorParityFixtures' -count=1

echo "Running runtime parity checks across tests/ (excluding vm-conformance)..."
while IFS= read -r f; do
  check_file "./tests" "$f"
done < <(find tests -name "*.slug" | sort | grep -v '^tests/vm-conformance/')

echo "Running runtime parity checks across test-suites/..."
while IFS= read -r f; do
  check_file "." "$f"
done < <(find test-suites -name "*.slug" | sort)

mismatches="$( (grep -E '^(EXIT_MISMATCH|OUTPUT_MISMATCH)\|' "$report" || true) | wc -l | tr -d ' ')"
echo "Runtime parity mismatches: $mismatches"
if [[ "$mismatches" -gt 0 ]]; then
  echo "== Runtime parity mismatches =="
  cat "$report"
  exit 1
fi

echo "Runtime parity passed."
