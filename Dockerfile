# builder image
FROM golang:alpine AS builder

WORKDIR /build

COPY ./go.mod ./
COPY ./go.sum ./

COPY ./ ./

RUN go mod download

RUN go build -v -o guessr cmd/guessr

# final image
FROM alpine

WORKDIR /build

COPY --from=builder /build/guessr /build/guessr

CMD ["./guessr", "start"]
