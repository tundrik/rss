FROM golang:alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build  ./cmd/crawly/crawly.go

FROM alpine

WORKDIR /build

COPY --from=builder /build/crawly /build/crawly

CMD ["./crawly"]