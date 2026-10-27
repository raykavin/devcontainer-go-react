# GO_VERSION, NODE_VERSION and TZ are injected by docker-compose.dev.yml
ARG GO_VERSION=1.27
ARG NODE_VERSION=24

# Node.js binaries come from the official prebuilt image on Docker Hub.
# Both images share the same Debian base (trixie), so the binaries are compatible.
FROM node:${NODE_VERSION}-trixie-slim AS node

FROM mcr.microsoft.com/devcontainers/go:2-${GO_VERSION}-trixie AS development

ARG TZ=America/Belem

ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=${TZ}

# The yarn apt repository shipped with this image has a broken signature on Trixie
RUN rm -f /etc/apt/sources.list.d/yarn.list

RUN apt-get update && apt-get install -y \
    fish \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=node /usr/local/lib/node_modules /usr/local/lib/node_modules

# npm, npx and corepack are Node scripts; recreate the symlinks the node image uses.
# The nvm bundled with the Go image is removed so its Node never shadows this one.
RUN ln -sf /usr/local/lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm \
    && ln -sf /usr/local/lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx \
    && ln -sf /usr/local/lib/node_modules/corepack/dist/corepack.js /usr/local/bin/corepack \
    && rm -rf /usr/local/share/nvm \
    && node --version && npm --version

RUN ln -snf /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone

# fish becomes the default shell in post-create.sh, after the image is built.
# Changing it here breaks tools that run `su vscode -c "..."` with a POSIX script.

USER vscode

CMD ["fish"]
