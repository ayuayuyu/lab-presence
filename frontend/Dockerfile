FROM node:22-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
ARG NEXT_PUBLIC_GOOGLE_CLIENT_ID
ENV NEXT_PUBLIC_GOOGLE_CLIENT_ID=${NEXT_PUBLIC_GOOGLE_CLIENT_ID}
ARG NEXT_PUBLIC_ADMIN_EMAILS
ENV NEXT_PUBLIC_ADMIN_EMAILS=${NEXT_PUBLIC_ADMIN_EMAILS}
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/out /usr/share/nginx/html
# nginx.conf はdocker-compose側でマウント
EXPOSE 80
