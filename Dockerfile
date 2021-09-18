# syntax=docker/dockerfile:1

FROM golang:1.16-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /build/main

FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /build/main /main

EXPOSE 8080

CMD [ "/main" ]