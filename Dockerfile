# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.20-alpine3.18 as build

WORKDIR $GOPATH/src/app
COPY . .

ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-w" -a -o  .


FROM alpine:3.18
ENV GOPATH="/go/src"
WORKDIR /run

COPY --from=build $GOPATH/app/rest-gateway .
COPY --from=build $GOPATH/app/conf/* conf/
EXPOSE 3000

ENTRYPOINT ["/run/rest-gateway"]
