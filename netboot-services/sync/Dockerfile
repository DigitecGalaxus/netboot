FROM alpine:3.18.2
RUN apk add --no-cache bash nano libc6-compat
RUN mkdir -p /tmp
WORKDIR /tmp
RUN wget https://aka.ms/downloadazcopy-v10-linux
RUN tar -xvf downloadazcopy-v10-linux
RUN mv azcopy_linux_amd64_*/azcopy /usr/local/bin/
RUN chmod +x /usr/local/bin/azcopy

RUN adduser -S syncer -u 1000 -s /bin/ash -h /home/syncer
WORKDIR /home/syncer
USER syncer
COPY syncer.sh /scripts/syncer.sh
CMD ["/scripts/syncer.sh"]
