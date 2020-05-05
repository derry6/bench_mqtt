FROM golang:1.13-alpine3.10 AS builder 

COPY . /bench_mqtt
ENV GOPROXY 'https://goproxy.cn,direct'

WORKDIR /bench_mqtt

RUN go build -o bench_mqtt

FROM alpine:3.10 AS runner 

COPY --from=builder /bench_mqtt/bench_mqtt /bench_mqtt 
WORKDIR /
CMD ["/bench_mqtt"]