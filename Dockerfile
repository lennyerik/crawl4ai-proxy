FROM golang:alpine AS builder
COPY . /root
WORKDIR /root
RUN go build

FROM alpine
COPY --from=builder /root/crawl-proxy /root
CMD ["/root/crawl-proxy"]
