FROM python:3.10.0rc2-alpine3.13
RUN pip3 install Jinja2 j2cli
RUN apk add --no-cache bash curl jq findutils
WORKDIR /work

COPY *.j2 /work/
COPY buildMainIpxeMenus.sh /work/buildMainIpxeMenus.sh
COPY buildIpxeMacMenus.sh /work/buildIpxeMacMenus.sh
COPY caching-server.ipxe.tmpl /work/caching-server.ipxe.tmpl
