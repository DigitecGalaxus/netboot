FROM alpine:3.14.0
RUN apk add --no-cache bash nano libc6-compat
RUN wget https://aka.ms/downloadazcopy-v10-linux
RUN tar -xvf downloadazcopy-v10-linux
RUN mv azcopy_linux_amd64_*/azcopy /usr/local/bin/
RUN rm -rf azcopy_linux_amd64_* downloadazcopy-v10-linux

RUN adduser -S syncer -s /bin/ash -h /home/syncer
WORKDIR /home/syncer
USER syncer

COPY syncer.sh /scripts/syncer.sh
CMD ["/scripts/syncer.sh"]
