FROM golang:1.23.1 AS build
RUN apt-get update && apt-get install -y python3 python3-venv
WORKDIR /work
COPY . /work

# Create and activate a virtual environment
RUN python3 -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

# Install Jinja2 in the virtual environment
RUN pip install Jinja2

# Run tests and build
RUN go test -v ./... -count=1
RUN CGO_ENABLED=0 go build -o menubuilder

FROM alpine:3.20.3
COPY --from=build /work/menubuilder /usr/local/bin/menubuilder
COPY *.j2 /work/
WORKDIR /work
ENTRYPOINT ["/usr/local/bin/menubuilder"]
