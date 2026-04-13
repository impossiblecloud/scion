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
# No scion-specific tooling or user setup

# Bun runtime source
FROM oven/bun:latest AS bun-source

FROM node:20-slim AS builder
ENV DEBIAN_FRONTEND=noninteractive

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    wget \
    curl \
    make \
    gcc \
    libssl-dev \
    libcurl4-gnutls-dev \
    libexpat1-dev \
    zlib1g-dev \
    gettext \
    ncurses-term \
    ca-certificates \
    gnupg \
    && rm -rf /var/lib/apt/lists/*

# Install Go
ARG GO_VERSION=1.26.2
RUN ARCH=$(dpkg --print-architecture) && \
    case "${ARCH}" in \
      amd64) GO_ARCH='linux-amd64' ;; \
      arm64) GO_ARCH='linux-arm64' ;; \
      *) echo "Unsupported architecture: ${ARCH}"; exit 1 ;; \
    esac && \
    curl -L "https://go.dev/dl/go${GO_VERSION}.${GO_ARCH}.tar.gz" -o go.tar.gz && \
    tar -C /usr/local -xzf go.tar.gz && \
    rm go.tar.gz

# Download GitHub CLI keyring
RUN mkdir -p -m 755 /etc/apt/keyrings && \
    wget -qO- https://cli.github.com/packages/githubcli-archive-keyring.gpg > /etc/apt/keyrings/githubcli-archive-keyring.gpg

# Download gcsfuse GPG key
RUN curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg

# Install Git from source
WORKDIR /opt
RUN wget https://www.kernel.org/pub/software/scm/git/git-2.52.0.tar.gz
RUN tar -xvf git-2.52.0.tar.gz
WORKDIR /opt/git-2.52.0/
RUN make prefix=/usr/local all
RUN make prefix=/usr/local install

# Final Stage
FROM node:20-slim
USER root

# Remove default node user to avoid UID 1000 collision with host user
RUN if getent passwd node; then userdel -r node; fi

COPY --from=builder /etc/apt/keyrings/githubcli-archive-keyring.gpg /etc/apt/keyrings/githubcli-archive-keyring.gpg
COPY --from=builder /usr/share/keyrings/cloud.google.gpg /usr/share/keyrings/cloud.google.gpg
RUN echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" > /etc/apt/sources.list.d/github-cli.list

# Setup gcsfuse repository
RUN . /etc/os-release && \
    echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt gcsfuse-${VERSION_CODENAME} main" | tee /etc/apt/sources.list.d/gcsfuse.list
RUN apt-get update && apt-get install -y --no-install-recommends \
    tmux \
    ca-certificates \
    libcurl4-gnutls-dev \
    libexpat1 \
    zlib1g \
    python3 \
    make \
    g++ \
    man-db \
    curl \
    wget \
    dnsutils \
    less \
    jq \
    bc \
    gh \
    unzip \
    rsync \
    ripgrep \
    procps \
    psmisc \
    lsof \
    socat \
    sudo \
    fzf \
    zsh \
    gnupg2 \
    iptables \
    ipset \
    iproute2 \
    aggregate \
    nano \
    vim \
    python3-venv \
    python3-pip \
    lsb-release \
    chromium \
    xz-utils \
    zip \
    openssh-client \
    gosu \
    tree \
    file \
    openssl \
    bash-completion \
    iputils-ping \
    netcat-openbsd \
    net-tools \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Setup GCS-FUSE
RUN GCSFUSE_REPO=gcsfuse-`lsb_release -c -s` && echo "deb [signed-by=/usr/share/keyrings/cloud.google.asc] https://packages.cloud.google.com/apt $GCSFUSE_REPO main" | sudo tee /etc/apt/sources.list.d/gcsfuse.list
RUN curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo tee /usr/share/keyrings/cloud.google.asc

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcsfuse \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Copy Go and Git from builder
COPY --from=builder /usr/local/go /usr/local/go
COPY --from=builder /usr/local/bin/git /usr/local/bin/git
COPY --from=builder /usr/local/libexec/git-core /usr/local/libexec/git-core
COPY --from=builder /usr/local/share/git-core /usr/local/share/git-core

# Copy Bun from bun-source
COPY --from=bun-source /usr/local/bin/bun /usr/local/bin/bun

# Install gcloud CLI
RUN echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && \
    apt-get update && apt-get install -y google-cloud-cli && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Install golangci-lint
RUN ARCH=$(dpkg --print-architecture) && \
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b /usr/local/bin

# Install yamllint
RUN pip3 install --no-cache-dir yamllint --break-system-packages

# Install OpenTofu
ARG TOFU_VERSION=1.11.6
RUN ARCH=$(dpkg --print-architecture) && \
    curl -fsSL "https://github.com/opentofu/opentofu/releases/download/v${TOFU_VERSION}/tofu_${TOFU_VERSION}_${ARCH}.deb" \
        -o /tmp/tofu.deb \
    && dpkg -i /tmp/tofu.deb \
    && rm /tmp/tofu.deb

# Install kubectl
ARG KUBECTL_VERSION=v1.35.3
RUN ARCH=$(dpkg --print-architecture) && \
    curl -fsSL "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${ARCH}/kubectl" \
        -o /usr/local/bin/kubectl \
    && chmod +x /usr/local/bin/kubectl

# Install Helm
ARG HELM_VERSION=v4.1.4
RUN ARCH=$(dpkg --print-architecture) && \
    curl -fsSL "https://get.helm.sh/helm-${HELM_VERSION}-linux-${ARCH}.tar.gz" \
    | tar -xz --strip-components=1 -C /usr/local/bin linux-${ARCH}/helm

# Install Helmfile
ARG HELMFILE_VERSION=1.4.4
RUN ARCH=$(dpkg --print-architecture) && \
    curl -fsSL "https://github.com/helmfile/helmfile/releases/download/v${HELMFILE_VERSION}/helmfile_${HELMFILE_VERSION}_linux_${ARCH}.tar.gz" \
    | tar -xz -C /usr/local/bin helmfile

# Install Helm diff plugin (required by helmfile)
# Install into a shared location so all users can find it
ENV HELM_DATA_HOME=/usr/local/share/helm
ARG HELM_DIFF_VERSION=3.10.0
# --verify=false is required for URL-based plugin installs; Helm's verification
# mechanism only works for plugins served from a signed repository index.
RUN helm plugin install https://github.com/databus23/helm-diff \
    --version "${HELM_DIFF_VERSION}" --verify=false
RUN chmod -R a+rX /usr/local/share/helm

# Install single-binary tools (talosctl, butane, yq, terragrunt)
ARG TALOS_VERSION=v1.12.6
ARG BUTANE_VERSION=v0.27.0
ARG YQ_VERSION=4.52.5
ARG TERRAGRUNT_VERSION=1.0.0
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

# Setup npm global directory
RUN mkdir -p /usr/local/share/npm-global
ENV NPM_CONFIG_PREFIX=/usr/local/share/npm-global
ENV GOPATH=/root/go
ENV PATH=/usr/local/share/npm-global/bin:/usr/local/go/bin:/root/go/bin:/usr/local/gcloud/google-cloud-sdk/bin:$PATH
ENV LC_ALL=C.UTF-8
ENV EDITOR=nano
ENV VISUAL=nano
ENV HELM_DATA_HOME=/usr/local/share/helm
ENV TF_PLUGIN_CACHE_DIR=/root/.terraform.d/plugin-cache

# Install Chrome DevTools MCP server globally
RUN npm install -g chrome-devtools-mcp

# Default to bash - no scion-specific entrypoint
CMD ["bash"]
