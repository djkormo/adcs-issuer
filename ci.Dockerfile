# syntax=docker/dockerfile:1.3

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot 
WORKDIR /
ARG BUILT_BINARY
COPY --chmod=555 ${BUILT_BINARY} /manager
USER nonroot:nonroot

ENTRYPOINT ["/manager"]
