FROM golang:1.13 as builder
RUN apt-get update && \
    apt-get install -y upx
WORKDIR /go/src/github.com/funbox/bacula_exporter
ENV GOPATH=/go/
COPY . .
RUN make clean && make bacula_exporter && make compress

FROM scratch
WORKDIR /
COPY --from=builder /go/src/github.com/funbox/bacula_exporter/bacula_exporter /app/bacula_exporter
EXPOSE 33407
CMD ["/app/bacula_exporter", "--config", "/config/bacula_exporter.knf"]
