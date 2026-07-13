#!/usr/bin/env bash

set -euo pipefail

readonly PROVIDER_ARCHIVE_MEDIA_TYPE="archive/zip"
readonly PROVIDER_INDEX_ARTIFACT_TYPE="application/vnd.opentofu.provider"
readonly PROVIDER_TARGET_ARTIFACT_TYPE="application/vnd.opentofu.provider-target"
readonly PROVIDER_ARCHIVE_PREFIX="terraform-provider-github"
readonly PROVIDER_DESCRIPTION="OpenTofu OCI mirror for the GitHub provider"
readonly PROVIDER_LICENSE="MPL-2.0"
readonly PROVIDER_SOURCE="https://github.com/pretty-good-software-org/terraform-provider-github"

if [ "$#" -ne 3 ]; then
  echo "usage: $0 VERSION DIST_DIR OCI_LAYOUT_DIR" >&2
  exit 2
fi

readonly version="$1"
readonly dist_dir="$2"
readonly oci_layout_dir="$3"
readonly archive_prefix="${PROVIDER_ARCHIVE_PREFIX}_${version}_"

if [ ! -d "$dist_dir" ]; then
  echo "distribution directory does not exist: $dist_dir" >&2
  exit 1
fi

if [ -e "$oci_layout_dir" ] && [ -n "$(find "$oci_layout_dir" -mindepth 1 -maxdepth 1 -print -quit 2>/dev/null)" ]; then
  echo "OCI layout directory must be empty: $oci_layout_dir" >&2
  exit 1
fi

mkdir -p "$oci_layout_dir"

target_tags=()
while IFS= read -r archive; do
  archive_name="$(basename "$archive")"
  platform="${archive_name#"$archive_prefix"}"
  platform="${platform%.zip}"

  if [[ "$platform" != *_* ]]; then
    echo "provider archive has no OS/architecture suffix: $archive_name" >&2
    exit 1
  fi

  os="${platform%%_*}"
  architecture="${platform#*_}"
  if [ -z "$os" ] || [ -z "$architecture" ]; then
    echo "provider archive has an invalid OS/architecture suffix: $archive_name" >&2
    exit 1
  fi

  target_tag="target-${platform}"
  (
    cd "$dist_dir"
    oras push \
      --artifact-platform "${os}/${architecture}" \
      --artifact-type "$PROVIDER_TARGET_ARTIFACT_TYPE" \
      --oci-layout "${oci_layout_dir}:${target_tag}" \
      "${archive_name}:${PROVIDER_ARCHIVE_MEDIA_TYPE}"
  )
  target_tags+=("$target_tag")
done < <(find "$dist_dir" -maxdepth 1 -type f -name "${archive_prefix}*.zip" -print | sort)

if [ "${#target_tags[@]}" -eq 0 ]; then
  echo "no provider archives found in $dist_dir for version $version" >&2
  exit 1
fi

oras manifest index create \
  --annotation "org.opencontainers.image.description=${PROVIDER_DESCRIPTION}" \
  --annotation "org.opencontainers.image.licenses=${PROVIDER_LICENSE}" \
  --annotation "org.opencontainers.image.source=${PROVIDER_SOURCE}" \
  --artifact-type "$PROVIDER_INDEX_ARTIFACT_TYPE" \
  --oci-layout \
  "${oci_layout_dir}:${version}" \
  "${target_tags[@]}"

oras resolve --oci-layout "${oci_layout_dir}:${version}"
