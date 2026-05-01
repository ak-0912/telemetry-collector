FROM mcr.microsoft.com/devcontainers/go:1.25

RUN apt-get update \
    && apt-get install -y --no-install-recommends curl git make ca-certificates \
    && rm -rf /var/lib/apt/lists/*
