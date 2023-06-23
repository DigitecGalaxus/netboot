FROM golang:1.20.5 AS build
WORKDIR /work
COPY . /work
RUN GCO_ENABLED=0 && go build --ldflags '-linkmode external -extldflags "-static"' -o menubuilder


FROM alpine:latest
COPY --from=build /work/menubuilder /usr/local/bin/menubuilder
COPY *.j2 /work/
WORKDIR /work
ENTRYPOINT /usr/local/bin/menubuilder
