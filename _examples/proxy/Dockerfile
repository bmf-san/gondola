# Runtime image for the gondola reverse proxy.
# The static binary is built on the host by `make up`
# (CGO_ENABLED=0 GOOS=linux go build -o gondola ./cmd).
FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY gondola /app/gondola
COPY config.yaml /app/config.yaml
COPY public /app/public
COPY certificates /app/certificates

EXPOSE 443
ENTRYPOINT ["/app/gondola", "-config", "config.yaml"]