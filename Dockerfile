FROM alpine:latest AS certs
RUN apk add --no-cache ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/hmcl-token-manager /app/hmcl-token-manager
WORKDIR /app
ENTRYPOINT ["/app/hmcl-token-manager"]
