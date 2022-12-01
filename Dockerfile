FROM golang:1.18-alpine3.15 as build

WORKDIR $GOPATH/src/app
COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -a -o  .


FROM alpine:3.15
ENV GOPATH="/go/src"
WORKDIR /run

COPY --from=build $GOPATH/app/http-proxy .
COPY --from=build $GOPATH/app/conf/* conf/
EXPOSE 3000

ENTRYPOINT ["/run/http-proxy"]