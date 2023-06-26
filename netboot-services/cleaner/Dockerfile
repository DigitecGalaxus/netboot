FROM golang:1.20.5 AS build
WORKDIR /work
COPY . /work
RUN GCO_ENABLED=0 && go build --ldflags '-linkmode external -extldflags "-static"' -o cleaner


FROM alpine:latest
COPY --from=build /work/cleaner /usr/local/bin/cleaner
WORKDIR /work
ENTRYPOINT /usr/local/bin/cleaner
