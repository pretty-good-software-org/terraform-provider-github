#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
readonly SCRIPT_DIR
readonly BUILD_SCRIPT="${SCRIPT_DIR}/build-opentofu-provider-oci.sh"
readonly TEST_VERSION="1.2.3-test.1"

test_root="$(mktemp -d)"
trap 'rm -rf "$test_root"' EXIT

fail() {
  echo "FAIL: $1" >&2
  exit 1
}

create_archive() {
  local dist_dir="$1"
  local platform="$2"
  local archive_name="terraform-provider-github_${TEST_VERSION}_${platform}.zip"
  local staging_dir="${test_root}/staging-${platform}"

  mkdir -p "$staging_dir"
  printf '#!/usr/bin/env sh\nexit 0\n' > "${staging_dir}/terraform-provider-github_v${TEST_VERSION}"
  chmod +x "${staging_dir}/terraform-provider-github_v${TEST_VERSION}"
  (
    cd "$staging_dir"
    zip -q "${dist_dir}/${archive_name}" "terraform-provider-github_v${TEST_VERSION}"
  )
}

build_valid_layout() {
  local dist_dir="${test_root}/valid-dist"
  local layout_dir="${test_root}/valid-layout"

  mkdir -p "$dist_dir"
  create_archive "$dist_dir" "darwin_arm64"
  create_archive "$dist_dir" "linux_amd64"
  "$BUILD_SCRIPT" "$TEST_VERSION" "$dist_dir" "$layout_dir" >/dev/null

  printf '%s' "$layout_dir"
}

test_builds_expected_index_structure() {
  local layout_dir="$1"
  local actual
  local expected

  actual="$(oras manifest fetch --oci-layout "${layout_dir}:${TEST_VERSION}" | jq -S '{artifactType, mediaType, schemaVersion, manifests: [.manifests[] | {artifactType, mediaType, platform}]}')"
  expected="$(jq -S -n '{
    artifactType: "application/vnd.opentofu.provider",
    mediaType: "application/vnd.oci.image.index.v1+json",
    schemaVersion: 2,
    manifests: [
      {
        artifactType: "application/vnd.opentofu.provider-target",
        mediaType: "application/vnd.oci.image.manifest.v1+json",
        platform: {architecture: "arm64", os: "darwin"}
      },
      {
        artifactType: "application/vnd.opentofu.provider-target",
        mediaType: "application/vnd.oci.image.manifest.v1+json",
        platform: {architecture: "amd64", os: "linux"}
      }
    ]
  }')"

  [ "$actual" = "$expected" ] || fail "provider index structure differs from the expected OpenTofu OCI structure"
}

test_targets_contain_one_zip_layer() {
  local layout_dir="$1"
  local target_tag

  for target_tag in target-darwin_arm64 target-linux_amd64; do
    oras manifest fetch --oci-layout "${layout_dir}:${target_tag}" |
      jq -e '
        .artifactType == "application/vnd.opentofu.provider-target" and
        (.layers | length) == 1 and
        .layers[0].mediaType == "archive/zip" and
        (.layers[0].digest | startswith("sha256:")) and
        .layers[0].size > 0
      ' >/dev/null || fail "${target_tag} does not contain exactly one valid provider zip layer"
  done
}

test_rejects_empty_distribution() {
  local dist_dir="${test_root}/empty-dist"
  local layout_dir="${test_root}/empty-layout"

  mkdir -p "$dist_dir"
  if "$BUILD_SCRIPT" "$TEST_VERSION" "$dist_dir" "$layout_dir" >"${test_root}/empty.stdout" 2>"${test_root}/empty.stderr"; then
    fail "empty distribution was accepted"
  fi
  grep -Fq "no provider archives found" "${test_root}/empty.stderr" || fail "empty distribution error lacks context"
}

test_rejects_nonempty_layout() {
  local dist_dir="${test_root}/occupied-dist"
  local layout_dir="${test_root}/occupied-layout"

  mkdir -p "$dist_dir" "$layout_dir"
  create_archive "$dist_dir" "linux_arm64"
  printf 'occupied' > "${layout_dir}/existing"
  if "$BUILD_SCRIPT" "$TEST_VERSION" "$dist_dir" "$layout_dir" >"${test_root}/occupied.stdout" 2>"${test_root}/occupied.stderr"; then
    fail "nonempty OCI layout was accepted"
  fi
  grep -Fq "OCI layout directory must be empty" "${test_root}/occupied.stderr" || fail "nonempty layout error lacks context"
}

valid_layout="$(build_valid_layout)"
readonly valid_layout
test_builds_expected_index_structure "$valid_layout"
test_targets_contain_one_zip_layer "$valid_layout"
test_rejects_empty_distribution
test_rejects_nonempty_layout

echo "build-opentofu-provider-oci tests passed"
