# Node major must match `.nvmrc` (use NVM locally: `nvm install` / `nvm use`).
FROM node:20-alpine

WORKDIR /app

COPY package.json package-lock.json* ./
RUN if [ -f package-lock.json ]; then npm ci --omit=dev; else npm install --omit=dev; fi

COPY src ./src

ENV NODE_ENV=production
USER node

EXPOSE 8080

CMD ["node", "src/index.js"]
