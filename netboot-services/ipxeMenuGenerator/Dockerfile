FROM golang:1.20.5-alpine AS build
WORKDIR /work
COPY . /work
RUN go test -v ./... -count=1
RUN GCO_ENABLED=0 && go build -o menubuilder

FROM alpine:3.18
COPY --from=build /work/menubuilder /usr/local/bin/menubuilder
COPY *.j2 /work/
WORKDIR /work
ENTRYPOINT /usr/local/bin/menubuilder
