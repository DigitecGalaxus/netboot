FROM alpine:3.14.0
RUN apk add --no-cache bash curl jq
WORKDIR /work

COPY getAllCachingServers.sh /work/getAllCachingServers.sh
COPY caching_server_list.json /work/caching_server_list.json
CMD ["/work/getAllCachingServers.sh"]
