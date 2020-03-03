FROM golang:1.12-alpine AS build 
ENV GO111MODULE on
ENV CGO_ENABLED 0
ENV GOOS linux

RUN apk add git make openssl

WORKDIR /go/src/github.com/klenkes74/k8s-mutate-webhook
ADD . .
RUN make test
RUN make app

FROM alpine
RUN apk --no-cache add ca-certificates && mkdir -p /app
WORKDIR /app

USER 1001

COPY --from=build /go/src/github.com/klenkes74/k8s-mutate-webhook/mutateme .
COPY --from=build /go/src/github.com/klenkes74/k8s-mutate-webhook/ssl ssl
CMD ["/app/mutateme"]
