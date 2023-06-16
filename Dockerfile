FROM golang:1.20-alpine3.18

#Builder
ARG BUILDER_IMAGE
FROM ${BUILDER_IMAGE} as builder

ARG VERSION
ARG COMMIT

WORKDIR /app

ADD ./ ./
RUN mkdir /out && \
    GOBIN=/out VERSION=${VERSION} COMMIT=${COMMIT} make -j8 dist
#
COPY go.mod .

RUN go mod download

COPY . .

RUN go build -o api cmd/api/main.go

EXPOSE 8877

CMD ["./api"] 
