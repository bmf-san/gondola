FROM gcr.io/distroless/static-debian12
COPY main /
ENTRYPOINT ["./main", "-config", "config.yaml"]