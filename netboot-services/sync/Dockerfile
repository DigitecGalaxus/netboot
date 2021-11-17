FROM alpine:3.14.0
RUN apk add --no-cache bash inotify-tools openssh-client rsync jq curl
RUN mkdir /ssh
WORKDIR /syncing/
COPY syncer.sh /scripts/syncer.sh
CMD ["/scripts/syncer.sh"]
