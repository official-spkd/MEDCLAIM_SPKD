# Frontend MedClaim — build statis Vite (base /app/) lalu sajikan via nginx.
FROM node:22-alpine AS build
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
# URL backend dibakar saat build (browser mengaksesnya dari host).
ARG VITE_API_BASE=http://localhost:7081
ENV VITE_API_BASE=$VITE_API_BASE VITE_MOCK=
RUN npx vite build

FROM nginx:1.27-alpine
COPY --from=build /src/dist /usr/share/nginx/html/app
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
