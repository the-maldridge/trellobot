FROM golang:latest as stage0
RUN mkdir -p /go/trellobot
COPY ./ /go/trellobot
RUN cd /go/trellobot && go mod vendor && CGO_ENABLED=0 go build -o /trellobot main.go

FROM scratch
COPY --from=stage0 /trellobot /
CMD ["/trellobot"]
EXPOSE 3000/tcp
USER 1000
