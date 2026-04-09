# builder image
FROM golang:alpine AS builder

WORKDIR /build

COPY ./go.mod ./
COPY ./go.sum ./

COPY ./ ./

RUN go mod download

RUN go build -v -o rankguessr .

# final image
FROM alpine

WORKDIR /build

COPY --from=builder /build/rankguessr /build/rankguessr

CMD ["./rankguessr", "start"]