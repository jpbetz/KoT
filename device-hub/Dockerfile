# Stage 1 - build react.js static single page webapp
FROM node:13.0.1-alpine as react-build
COPY webapp /
RUN yarn
RUN yarn build

# Stage 2 - build golang service
FROM golang:1.12 as go-build
COPY service /
WORKDIR /
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -o app

# Stage 2 - create release image
FROM gcr.io/distroless/base
COPY --from=react-build /build /static
COPY --from=go-build /app /app
EXPOSE 8080
CMD ["/app"]