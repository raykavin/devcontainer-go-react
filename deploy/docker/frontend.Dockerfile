# Global build arguments
ARG NODE_BUILD_BASE_IMAGE=24-trixie-slim
ARG PROJECT_NAME="my-app"
ARG TZ=America/Belem

# Build stage
FROM node:${NODE_BUILD_BASE_IMAGE} AS build

ARG PROJECT_NAME
ARG TZ

ARG VITE_API_BASE_URL=https://api.example.com

ENV APP_NAME=${PROJECT_NAME}
ENV TZ=${TZ}
ENV CI=true
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

WORKDIR /build

RUN apt-get update \
 && apt-get install -y --no-install-recommends git tzdata \
 && rm -rf /var/lib/apt/lists/*

RUN ln -snf /usr/share/zoneinfo/${TZ} /etc/localtime \
 && echo ${TZ} > /etc/timezone

RUN corepack enable

COPY package.json pnpm-lock.yaml* ./

RUN pnpm install --frozen-lockfile

COPY . .

RUN pnpm run build

# Runtime stage
FROM nginx:alpine AS runtime

ARG PROJECT_NAME
ARG APP_USER_ID=1001
ARG APP_GROUP_ID=1001
ARG TZ=America/Belem

ENV APP_NAME=${PROJECT_NAME}
ENV TZ=${TZ}

RUN apk add --no-cache tzdata curl

RUN ln -snf /usr/share/zoneinfo/${TZ} /etc/localtime \
 && echo ${TZ} > /etc/timezone

COPY --from=build /build/dist /usr/share/nginx/html

RUN addgroup -g ${APP_GROUP_ID} ${APP_NAME} \
 && adduser -D -u ${APP_USER_ID} -G ${APP_NAME} -s /bin/sh ${APP_NAME} \
 && chown -R ${APP_NAME}:${APP_NAME} /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost/ || exit 1

CMD ["nginx", "-g", "daemon off;"]