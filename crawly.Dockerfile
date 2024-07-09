FROM golang:alpine AS builder

WORKDIR /build

ADD go.mod .

COPY . .

RUN go build  ./cmd/crawly/crawly.go

FROM alpine

WORKDIR /build

COPY --from=builder /build/crawly /build/crawly

CMD ["./crawly"]