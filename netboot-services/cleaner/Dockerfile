FROM golang:1.21.1-alpine AS build
WORKDIR /app
COPY . .
RUN go test -v ./... -count=1
RUN CGO_ENABLED=0 go build -o cleaner

FROM alpine:3.18
WORKDIR /app
COPY --from=build /app/cleaner .
ENTRYPOINT ["/app/cleaner"]
