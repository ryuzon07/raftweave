FROM node:22-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci

COPY . ./
RUN mkdir -p /app/.angular && chown -R 65532:65532 /app && chmod -R 777 /app/.angular

EXPOSE 4200
CMD ["npm", "start", "--", "--host", "0.0.0.0"]
