FROM alpine:3.14.0

RUN apk add --no-cache tftp-hpa

ENTRYPOINT ["/usr/sbin/in.tftpd"]
CMD ["--foreground", "--listen", "--secure", "/config/menus", "--blocksize", "1200"]
