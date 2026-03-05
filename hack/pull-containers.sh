#!/bin/bash
# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

IMAGES=(
  "us-central1-docker.pkg.dev/ptone-misc/public-docker/scion-gemini:latest"
  "us-central1-docker.pkg.dev/ptone-misc/public-docker/scion-claude:latest"
  "us-central1-docker.pkg.dev/ptone-misc/public-docker/scion-opencode:latest"
  "us-central1-docker.pkg.dev/ptone-misc/public-docker/scion-codex:latest"
)

detect_runtime() {
  if command -v container &>/dev/null && [[ "$(uname)" == "Darwin" ]]; then
    echo "container"
  elif command -v docker &>/dev/null; then
    echo "docker"
  elif command -v podman &>/dev/null; then
    echo "podman"
  else
    echo ""
  fi
}

RUNTIME="${1:-}"

if [[ -n "$RUNTIME" ]]; then
  case "$RUNTIME" in
    container|docker|podman)
      ;;
    *)
      echo "Error: unsupported runtime '$RUNTIME'. Use one of: container, docker, podman"
      exit 1
      ;;
  esac
else
  RUNTIME="$(detect_runtime)"
  if [[ -z "$RUNTIME" ]]; then
    echo "Error: no container runtime found. Install docker, podman, or container (macOS)."
    exit 1
  fi
fi

echo "Using runtime: $RUNTIME"
echo ""

for image in "${IMAGES[@]}"; do
  echo "Pulling: $image"
  "$RUNTIME" image pull "$image"
  echo ""
done

echo "Pruning unused images..."
"$RUNTIME" image prune -f
echo "Done."
