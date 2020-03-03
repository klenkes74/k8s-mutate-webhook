FROM golang:1.12-alpine AS build 
ENV GO111MODULE on
ENV CGO_ENABLED 0
ENV GOOS linux

RUN apk add git make openssl

WORKDIR /go/src/github.com/klenkes74/k8s-mutate-webhook
ADD . .
RUN make test app
RUN ls -ral /go/src/github.com/klenkes74/k8s-mutate-webhook


FROM alpine
RUN apk --no-cache add ca-certificates && mkdir -p /app
WORKDIR /app

COPY --from=build /go/src/github.com/klenkes74/k8s-mutate-webhook/mutateme .
COPY --from=build /go/src/github.com/klenkes74/k8s-mutate-webhook/ssl ssl

USER 1001

CMD ["/app/mutateme"]
