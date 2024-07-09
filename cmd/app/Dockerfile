FROM golang:alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build  ./cmd/app/app.go

FROM alpine

WORKDIR /build

COPY --from=builder /build/app /build/app

CMD ["./app"]