FROM gcr.io/distroless/static-debian12
COPY gondola /
ENTRYPOINT ["./gondola", "-config", "config.yaml"]