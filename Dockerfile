FROM golang:1.26.7-alpine3.24 AS builder

RUN apk --update add ca-certificates make git
RUN echo 'polestar:*:65532:' > /tmp/group && \
    echo 'polestar:*:65532:65532:polestar:/:/polestar-exporter' > /tmp/passwd

WORKDIR /workspace
COPY . /workspace

RUN make build

FROM scratch

ARG IMAGE_VERSION="unknown"
ARG IMAGE_REVISION="unknown"

LABEL org.opencontainers.image.title="polestar-exporter" \
      org.opencontainers.image.description="Prometheus exporter for Polestar cars" \
      org.opencontainers.image.authors="Terje Sannum <terje@offpiste.org>" \
      org.opencontainers.image.url="https://github.com/terjesannum/polestar-exporter" \
      org.opencontainers.image.source="https://github.com/terjesannum/polestar-exporter" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      org.opencontainers.image.revision="${IMAGE_REVISION}"

WORKDIR /
EXPOSE 8080

COPY --from=builder /tmp/passwd /tmp/group /etc/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /workspace/polestar-exporter .

USER 65532:65532

ENTRYPOINT ["/polestar-exporter"]
