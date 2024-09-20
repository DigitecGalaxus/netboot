FROM golang:1.23.1-alpine AS build
WORKDIR /app
COPY . .
RUN go test -v ./... -count=1
RUN CGO_ENABLED=0 go build -o cleaner

FROM alpine:3.20.3
WORKDIR /app
COPY --from=build /app/cleaner .
ENTRYPOINT ["/app/cleaner"]
