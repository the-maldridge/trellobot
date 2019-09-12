FROM golang:latest as stage0
RUN mkdir -p /go/trellobot
COPY ./ /go/trellobot
RUN cd /go/trellobot && go mod vendor && CGO_ENABLED=0 go build -o /trellobot main.go

FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch
COPY --from=stage0 /trellobot /
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
CMD ["/trellobot"]
EXPOSE 3000/tcp
USER 1000
