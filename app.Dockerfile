FROM golang:alpine AS builder

WORKDIR /build

ADD go.mod .

COPY . .

RUN go build  ./cmd/app/app.go

FROM alpine

WORKDIR /build

COPY --from=builder /build/app /build/app

CMD ["./app"]