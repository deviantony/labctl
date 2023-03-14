# go-workbench Dockerfile template v1.0.1
FROM alpine:3.6 as base
RUN apk add -U --no-cache ca-certificates

FROM scratch
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY dist /

ENTRYPOINT [ "/labctl" ]