# syntax=docker/dockerfile:1
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

# Core Base Image - Foundation tools only (Go, Git, unix tools, GCS fuse)
# No scion-specific tooling or user setup.
#
# The install steps live in ../lib/install-core-toolchain.sh and the closing
# assertion in ../lib/verify-base-contract.sh. thick-prep runs the same two
# scripts. That is the point: these two bases are the only things scion-base is
# ever built on, they used to be independent recipes, and they drifted without
# anyone noticing (see ../lib/verify-base-contract.sh for the casualties). One
# recipe cannot drift from itself.
#
# Each step is still its own RUN, so layer caching survives the move out of
# this file.
#
# NOTE: the build context is image-build/, not image-build/core-base/, so that
# ../lib is reachable. See image-build/scripts/lib/targets.sh step_context_dir.

# GitHub CLI version — bump this variable to upgrade gh across all images.
ARG GH_VERSION=2.100.0

# Go: a floor and a pin. GO_MIN_VERSION is what scion needs (go.mod);
# GO_VERSION is what gets installed when the base does not already meet the
# floor. A base that ships something newer keeps it — these images serve many
# consumers besides scion, so we do not downgrade one to suit the other.
ARG GO_MIN_VERSION=1.26.1
ARG GO_VERSION=1.26.6

# Git is vendored from the Chainguard git image rather than compiled from
# source. Chainguard tracks upstream closely (currently 2.55.0) and rebuilds on
# CVE disclosure, which removes a multi-minute `make` from every core-base
# build.
#
# Worth stating plainly: Debian trixie's own apt git is 2.47.3, which already
# clears scion's 2.47.0 floor. core-base does not *need* the vendored git —
# the thick chain does (Ubuntu noble caps at 2.43.0). Vendoring here is an
# anti-drift choice, so both chains get byte-identical git, not a necessity.
#
# NOTE: `:latest` is a floating tag. CI should override this with a digest
# (`--build-arg GIT_IMAGE=chainguard/git@sha256:...`) so builds are
# reproducible, and refresh that digest on a schedule — a vendored binary stops
# receiving Chainguard's rebuild cadence the moment it is copied out.
ARG GIT_IMAGE=chainguard/git:latest

# Bun runtime source
FROM oven/bun:latest AS bun-source

# Git source (vendored binary + helpers)
FROM ${GIT_IMAGE} AS git-source

# --- Stage the vendored git and localize its libpcre2 ------------------------
# See ../lib/stage-vendored-git.sh for why the libpcre2 override is a RUNPATH
# on five binaries rather than an image-wide LD_LIBRARY_PATH. patchelf and
# binutils are confined to this stage and never reach the final image.
FROM node:24-trixie-slim AS git-patched
RUN apt-get update && apt-get install -y --no-install-recommends \
    patchelf \
    binutils \
    && rm -rf /var/lib/apt/lists/*
COPY lib/stage-vendored-git.sh /tmp/stage-vendored-git.sh
COPY --from=git-source /usr/bin/git          /staged/usr/bin/git
COPY --from=git-source /usr/libexec/git-core /staged/usr/libexec/git-core
COPY --from=git-source /usr/share/git-core   /staged/usr/share/git-core
COPY --from=git-source /usr/lib/libpcre2-8.so.0 /staged/usr/local/lib/scion-git/libpcre2-8.so.0
RUN bash /tmp/stage-vendored-git.sh /staged

FROM node:24-trixie-slim
USER root
ENV DEBIAN_FRONTEND=noninteractive

# The toolchain scripts stay in the finished image rather than being deleted:
# `docker run <image> /opt/scion-image-build/lib/verify-base-contract.sh`
# re-checks the contract against a running container, which is how you tell a
# broken base apart from a broken agent. They are deliberately NOT added to
# PATH — build tooling should not shadow anything at runtime.
COPY lib/ /opt/scion-image-build/lib/
RUN chmod +x /opt/scion-image-build/lib/*.sh

# Free uid 1000 (named `node` on this base) for the scion user that scion-base
# creates there.
RUN /opt/scion-image-build/lib/install-core-toolchain.sh free-uid-1000

RUN /opt/scion-image-build/lib/install-core-toolchain.sh apt-common
RUN /opt/scion-image-build/lib/install-core-toolchain.sh chromium
RUN /opt/scion-image-build/lib/install-core-toolchain.sh gcsfuse
RUN /opt/scion-image-build/lib/install-core-toolchain.sh gcloud

# --- Vendored git (Chainguard) ---------------------------------------------
# Three constraints, all learned the hard way:
#
# 1. Helpers MUST land in /usr/libexec/git-core, not /usr/local/libexec.
#    That path is compiled into the Chainguard binary as GIT_EXEC_PATH. Putting
#    them anywhere else yields the very misleading "git: 'remote-https' is not
#    a git command" / "fatal: remote helper 'https' aborted session", which
#    reads like missing HTTPS support rather than a path error.
#
# 2. The binary MUST land at /usr/bin/git specifically. Chainguard ships
#    /usr/libexec/git-core/git as a relative symlink to ../../bin/git, and git
#    re-execs itself through that path for internal subcommands (auto
#    maintenance, etc). Install the binary anywhere else (e.g.
#    /usr/local/bin/git) on a base that already has a distro git and the
#    symlink silently resolves to the OLD binary. The failure is baffling:
#    `git --version` reports 2.55.0 while a subcommand dies with
#    "error: unknown option \`detach'" followed by
#    "fatal: unknown repository extension found: relativeworktrees".
#
# 3. This COPY must stay AFTER every apt-get above. Several packages pull in
#    the distro git as a dependency; landing our binary last is what makes it
#    win — and per (2) it must overwrite /usr/bin/git, not sit beside it.
#
# Chainguard builds against a newer glibc, so the base image needs glibc >= 2.38
# (trixie ships 2.41). This is a real constraint on future base changes:
# bookworm (2.36) would fail at exec with "GLIBC_2.38 not found".
COPY --from=git-patched /staged/usr/bin/git          /usr/bin/git
COPY --from=git-patched /staged/usr/libexec/git-core /usr/libexec/git-core
COPY --from=git-patched /staged/usr/share/git-core   /usr/share/git-core
COPY --from=git-patched /staged/usr/local/lib/scion-git /usr/local/lib/scion-git

ARG GO_MIN_VERSION
ARG GO_VERSION
RUN /opt/scion-image-build/lib/install-core-toolchain.sh go "${GO_MIN_VERSION}" "${GO_VERSION}"

ARG GH_VERSION
RUN /opt/scion-image-build/lib/install-core-toolchain.sh gh "${GH_VERSION}"

# Pinned rather than tracking dl.k8s.io/release/stable.txt, so a rebuild of an
# old image produces the same kubectl it originally shipped. The step falls back
# to stable when no version is passed, which is what thick-prep does.
ARG KUBECTL_VERSION=v1.36.3
RUN /opt/scion-image-build/lib/install-core-toolchain.sh kubectl "${KUBECTL_VERSION}"
RUN /opt/scion-image-build/lib/install-core-toolchain.sh golangci-lint

# --- Infrastructure tooling --------------------------------------------------
# Everything below is core-base only: it is not part of the base contract that
# ../lib/verify-base-contract.sh asserts, and thick-prep does not install it.
# That is why these stay as RUN steps here instead of moving into
# ../lib/install-core-toolchain.sh — a tool only one chain installs cannot drift
# between the two chains, and the script is reserved for what both must share.

# Install OpenTofu
ARG TOFU_VERSION=1.12.5
RUN ARCH=$(dpkg --print-architecture) && \
    curl -fsSL "https://github.com/opentofu/opentofu/releases/download/v${TOFU_VERSION}/tofu_${TOFU_VERSION}_${ARCH}.deb" \
        -o /tmp/tofu.deb \
    && dpkg -i /tmp/tofu.deb \
    && rm /tmp/tofu.deb

# Install Helm
ARG HELM_VERSION=v4.2.4
RUN ARCH=$(dpkg --print-architecture) && \
    curl -fsSL "https://get.helm.sh/helm-${HELM_VERSION}-linux-${ARCH}.tar.gz" \
    | tar -xz --strip-components=1 -C /usr/local/bin linux-${ARCH}/helm

# Install Helmfile
ARG HELMFILE_VERSION=1.7.4
RUN ARCH=$(dpkg --print-architecture) && \
    curl -fsSL "https://github.com/helmfile/helmfile/releases/download/v${HELMFILE_VERSION}/helmfile_${HELMFILE_VERSION}_linux_${ARCH}.tar.gz" \
    | tar -xz -C /usr/local/bin helmfile

# Install Helm diff plugin (required by helmfile)
# Install into a shared location so all users can find it
ENV HELM_DATA_HOME=/usr/local/share/helm
ARG HELM_DIFF_VERSION=3.15.11
# --verify=false is required for URL-based plugin installs; Helm's verification
# mechanism only works for plugins served from a signed repository index.
RUN helm plugin install https://github.com/databus23/helm-diff \
    --version "${HELM_DIFF_VERSION}" --verify=false
RUN chmod -R a+rwX /usr/local/share/helm

# Install single-binary tools (talosctl, butane, yq, terragrunt)
ARG TALOS_VERSION=v1.13.8
ARG BUTANE_VERSION=v0.29.0
ARG YQ_VERSION=4.53.3
ARG TERRAGRUNT_VERSION=1.1.3
RUN ARCH=$(dpkg --print-architecture) && \
    BUTANE_ARCH=$([ "${ARCH}" = "arm64" ] && echo "aarch64-unknown-linux-gnu" || echo "x86_64-unknown-linux-gnu") && \
    curl -fsSL "https://github.com/siderolabs/talos/releases/download/${TALOS_VERSION}/talosctl-linux-${ARCH}" \
        -o /usr/local/bin/talosctl \
    && curl -fsSL "https://github.com/coreos/butane/releases/download/${BUTANE_VERSION}/butane-${BUTANE_ARCH}" \
        -o /usr/local/bin/butane \
    && curl -fsSL "https://github.com/mikefarah/yq/releases/download/v${YQ_VERSION}/yq_linux_${ARCH}" \
        -o /usr/local/bin/yq \
    && curl -fsSL "https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_${ARCH}" \
        -o /usr/local/bin/terragrunt \
    && chmod +x /usr/local/bin/talosctl /usr/local/bin/butane \
               /usr/local/bin/yq /usr/local/bin/terragrunt

# Install AWS CLI v2
RUN ARCH=$(dpkg --print-architecture) && \
    AWS_ARCH=$([ "${ARCH}" = "arm64" ] && echo "aarch64" || echo "x86_64") && \
    curl -fsSL "https://awscli.amazonaws.com/awscli-exe-linux-${AWS_ARCH}.zip" -o /tmp/awscliv2.zip \
    && unzip -q /tmp/awscliv2.zip -d /tmp \
    && /tmp/aws/install \
    && rm -rf /tmp/awscliv2.zip /tmp/aws

# Install SOPS (Secrets OPerationS)
ARG SOPS_VERSION=3.13.3
RUN ARCH=$(dpkg --print-architecture) && \
    curl -fsSL "https://github.com/getsops/sops/releases/download/v${SOPS_VERSION}/sops-v${SOPS_VERSION}.linux.${ARCH}" \
        -o /usr/local/bin/sops && \
    chmod +x /usr/local/bin/sops

# Install Kubeshark CLI (for remote MCP access via kubeshark mcp serve)
ARG KUBESHARK_VERSION=53.4.0
RUN ARCH=$(dpkg --print-architecture) && \
    curl -fsSL "https://github.com/kubeshark/kubeshark/releases/download/v${KUBESHARK_VERSION}/kubeshark_linux_${ARCH}" \
        -o /usr/local/bin/kubeshark && \
    chmod +x /usr/local/bin/kubeshark

# Install stern (multi-pod log tailing)
ARG STERN_VERSION=1.34.0
RUN ARCH=$(dpkg --print-architecture) && \
    curl -fsSL "https://github.com/stern/stern/releases/download/v${STERN_VERSION}/stern_${STERN_VERSION}_linux_${ARCH}.tar.gz" \
    | tar -xz -C /usr/local/bin stern && \
    chmod +x /usr/local/bin/stern


# Copy Bun from bun-source
COPY --from=bun-source /usr/local/bin/bun /usr/local/bin/bun

RUN /opt/scion-image-build/lib/install-core-toolchain.sh npm-global
ENV NPM_CONFIG_PREFIX=/usr/local/share/npm-global

# Optional npm mirror. Corporate networks commonly block registry.npmjs.org;
# point this at an internal proxy (Artifactory, Nexus, Verdaccio) to build
# behind one. Defaults to the public registry, so behaviour is unchanged when
# unset. Inherited by scion-base, every harness image, and agents at runtime.
ARG NPM_REGISTRY=https://registry.npmjs.org/
ENV NPM_CONFIG_REGISTRY=${NPM_REGISTRY}

# Optional Python package index, for the same reason as NPM_REGISTRY above:
# corporate networks commonly block pypi.org. Point this at an internal proxy
# (Artifactory, Nexus, devpi). Defaults to the public index, so behaviour is
# unchanged when unset. Inherited by scion-base, every harness image, and
# agents at runtime. Credentials do NOT belong here - a value baked into ENV
# persists in the image; use the pipconf build secret instead.
ARG PIP_INDEX_URL=https://pypi.org/simple
ENV PIP_INDEX_URL=${PIP_INDEX_URL}

ENV GOPATH=/root/go
ENV PATH=/usr/local/share/npm-global/bin:/usr/local/go/bin:/root/go/bin:/usr/local/gcloud/google-cloud-sdk/bin:$PATH
ENV LC_ALL=C.UTF-8
ENV EDITOR=nano
ENV VISUAL=nano

# yamllint and PyYAML. Deliberately after the ENV block above so the install
# honours PIP_INDEX_URL — run any earlier and the pip mirror escape hatch is
# silently bypassed.
RUN pip3 install --no-cache-dir --break-system-packages yamllint pyyaml

# Chrome DevTools MCP server, the playwright CLI and the Prometheus MCP server,
# installed globally.
RUN --mount=type=secret,id=npmrc,target=/root/.npmrc,required=false \
    /opt/scion-image-build/lib/install-core-toolchain.sh npm-tools

# Fail the build, not the agent, when this image does not satisfy what
# scion-base needs from the layer beneath it. thick-prep ends with the exact
# same line.
RUN GO_MIN_VERSION="${GO_MIN_VERSION}" \
    /opt/scion-image-build/lib/verify-base-contract.sh

# Default to bash - no scion-specific entrypoint
CMD ["bash"]
