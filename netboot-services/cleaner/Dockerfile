FROM golang:1.20.5 AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o cleaner

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=build /app/cleaner .
ENTRYPOINT ["/app/cleaner"]
