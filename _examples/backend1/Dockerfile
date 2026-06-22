# Build stage.
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/backend .

# Runtime stage.
FROM gcr.io/distroless/static-debian12
COPY --from=build /out/backend /backend
EXPOSE 8081
ENTRYPOINT ["/backend"]