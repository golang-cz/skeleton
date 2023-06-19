# -----------------------------------------------------------------
# Builder
# -----------------------------------------------------------------
ARG BUILDER_IMAGE

FROM ${BUILDER_IMAGE} as builder

ARG VERSION
ARG COMMIT

WORKDIR /skeleton
ADD ./ ./

RUN apk add --update make bash git
RUN make vendor
RUN mkdir /out && \
    GOBIN=/out VERSION=${VERSION} COMMIT=${COMMIT} make build

# -----------------------------------------------------------------
# Runner
# -----------------------------------------------------------------
FROM alpine:3.16

ENV TZ=UTC

RUN apk add --no-cache --update ca-certificates

COPY --from=builder /out/api /usr/bin/
COPY --from=builder /out/goose /usr/bin/

EXPOSE 8877

CMD ["api", "-config=/etc/dev.toml"]

