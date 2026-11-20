FROM node:22-alpine AS build

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .

ARG API_BASE=/api
ENV VITE_API_BASE=$API_BASE

RUN npm run build

FROM nginxinc/nginx-unprivileged:1.30-alpine

COPY --from=build --chown=101:101 /app/dist /usr/share/nginx/html

COPY --chown=101:101 nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]
