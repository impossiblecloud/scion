# Assumes a local copy of gemini-cli-sandbox


FROM gemini-cli-sandbox
USER root

RUN apt-get update && apt-get install -y --no-install-recommends \
  tmux \
  curl \
  git \
  ca-certificates \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

# Install Go
ARG GO_VERSION=1.25.4
RUN ARCH=$(dpkg --print-architecture) && \
    case "${ARCH}" in \
      amd64) GO_ARCH='linux-amd64' ;; \
      arm64) GO_ARCH='linux-arm64' ;; \
      *) echo "Unsupported architecture: ${ARCH}"; exit 1 ;; \
    esac && \
    curl -L "https://go.dev/dl/go${GO_VERSION}.${GO_ARCH}.tar.gz" -o go.tar.gz && \
    tar -C /usr/local -xzf go.tar.gz && \
    rm go.tar.gz

ENV PATH=/usr/local/go/bin:$PATH

USER node