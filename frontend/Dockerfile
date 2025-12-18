FROM node:22-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .

ARG API_BASE=http://localhost:8080
ENV NUXT_PUBLIC_API_BASE=$API_BASE

RUN npm run generate

FROM nginx:alpine

COPY --from=builder /app/.output/public /usr/share/nginx/html

COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]
