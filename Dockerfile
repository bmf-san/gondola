# Used by GoReleaser (dockers_v2): the pre-built binary for each target
# platform is provided in the build context. Do not build the binary here.
FROM gcr.io/distroless/static-debian12
ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/gondola /gondola
ENTRYPOINT ["/gondola", "-config", "/etc/gondola/config.yaml"]