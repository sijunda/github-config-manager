#!/usr/bin/env bash
set -euo pipefail

# Verifies that GoReleaser produced the release assets expected by the installers
# and that checksums.txt contains valid SHA-256 entries for each archive.

dist_dir="${1:-dist}"
checksums_file="${dist_dir}/checksums.txt"

fail() {
  echo "ERROR: $*" >&2
  exit 1
}

if [[ ! -d "$dist_dir" ]]; then
  fail "dist directory not found: $dist_dir"
fi
if [[ ! -f "$checksums_file" ]]; then
  fail "checksums file not found: $checksums_file"
fi

if command -v sha256sum >/dev/null 2>&1; then
  sha256() { sha256sum "$1" | awk '{print $1}'; }
elif command -v shasum >/dev/null 2>&1; then
  sha256() { shasum -a 256 "$1" | awk '{print $1}'; }
else
  fail "sha256sum or shasum is required"
fi

verify_checksum() {
  local asset="$1"
  local filename expected actual
  filename=$(basename "$asset")
  expected=$(awk -v target="$filename" '$2 == target {print $1}' "$checksums_file")
  [[ -n "$expected" ]] || fail "missing checksum entry for $filename"

  actual=$(sha256 "$asset")
  [[ "$actual" == "$expected" ]] || fail "checksum mismatch for $filename: expected $expected, got $actual"
}

verify_archive_contains_binary() {
  local asset="$1"
  case "$asset" in
    *.tar.gz)
      tar -tzf "$asset" | grep -Eq '(^|/)gcm$' || fail "archive does not contain gcm binary: $(basename "$asset")"
      ;;
    *.zip)
      command -v unzip >/dev/null 2>&1 || fail "unzip is required to inspect zip archives"
      unzip -Z1 "$asset" | grep -Eq '(^|/)gcm.exe$' || fail "archive does not contain gcm.exe binary: $(basename "$asset")"
      ;;
    *)
      fail "unsupported archive type: $asset"
      ;;
  esac
}

require_one_asset() {
  local pattern="$1"
  local matches=()
  shopt -s nullglob
  matches=("${dist_dir}"/${pattern})
  shopt -u nullglob

  if [[ ${#matches[@]} -ne 1 ]]; then
    fail "expected exactly one asset matching ${pattern}, found ${#matches[@]}"
  fi

  verify_checksum "${matches[0]}"
  verify_archive_contains_binary "${matches[0]}"
  echo "verified $(basename "${matches[0]}")"
}

require_one_asset 'gcm_*_darwin_amd64.tar.gz'
require_one_asset 'gcm_*_darwin_arm64.tar.gz'
require_one_asset 'gcm_*_linux_amd64.tar.gz'
require_one_asset 'gcm_*_linux_arm64.tar.gz'
require_one_asset 'gcm_*_linux_arm.tar.gz'
require_one_asset 'gcm_*_windows_amd64.zip'

echo "release artifacts verified"
