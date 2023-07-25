FROM alpine:3.18 as bootfiles
RUN apk add --no-cache bash gcc binutils make perl xz-libs xz-dev mtools git libc-dev curl openssl coreutils
WORKDIR /build
# Clone iPXE
RUN git clone git://git.ipxe.org/ipxe.git
RUN sed -i 's/#undef\tDOWNLOAD_PROTO_HTTPS/#define\ DOWNLOAD_PROTO_HTTPS/' ipxe/src/config/general.h
RUN sed -i 's/\/\/#define PING_CMD/#define PING_CMD/' ipxe/src/config/general.h
RUN sed -i 's/\/\/#define NSLOOKUP_CMD/#define NSLOOKUP_CMD/' ipxe/src/config/general.h
# Workaround for fetching data from servers with a Let's Encrypt certificate.
# https://github.com/ipxe/ipxe/issues/606 - Certificate validation fails when the last certificate of the chain is signed by an expired CA (DST Root CA X3)
RUN export CERT=ca.pem,isrgrootx1.pem,lets-encrypt-r3.pem && export TRUST=ca.pem,isrgrootx1.pem,lets-encrypt-r3.pem
RUN curl -s http://ca.ipxe.org/ca.crt > ipxe/src/ca.pem 
RUN curl -s https://letsencrypt.org/certs/isrgrootx1.pem > ipxe/src/isrgrootx1.pem
RUN curl -s https://letsencrypt.org/certs/lets-encrypt-r3.pem > ipxe/src/lets-encrypt-r3.pem

# Include custom logic
COPY custom.ipxe ipxe/src/custom.ipxe
# Run the builds
# Note: We're using snponly to retain the original UEFI-drivers and thus improving reliability and reducing the size of the bootloader.
RUN cd ipxe/src && make bin/undionly.kpxe EMBED=custom.ipxe CERT=ca.pem,isrgrootx1.pem,lets-encrypt-r3.pem TRUST=ca.pem,isrgrootx1.pem,lets-encrypt-r3.pem && \
make bin-i386-efi/snponly.efi EMBED=custom.ipxe CERT=ca.pem,isrgrootx1.pem,lets-encrypt-r3.pem TRUST=ca.pem,isrgrootx1.pem,lets-encrypt-r3.pem  && \
make bin-x86_64-efi/snponly.efi EMBED=custom.ipxe CERT=ca.pem,isrgrootx1.pem,lets-encrypt-r3.pem TRUST=ca.pem,isrgrootx1.pem,lets-encrypt-r3.pem 

FROM alpine:3.18

RUN apk add --no-cache tftp-hpa curl
RUN mkdir -p /srv/tftp/ipxe

COPY  --from=bootfiles /build/ipxe/src/bin/undionly.kpxe /srv/tftp/undionly.kpxe
COPY  --from=bootfiles /build/ipxe/src/bin-i386-efi/snponly.efi /srv/tftp/ipxe32.efi
COPY  --from=bootfiles /build/ipxe/src/bin-x86_64-efi/snponly.efi /srv/tftp/ipxe64.efi

RUN adduser -D tftp -H
EXPOSE 69/udp

ENTRYPOINT ["/usr/sbin/in.tftpd"]
CMD ["--foreground", "--listen", "--secure", "/srv/tftp", "--blocksize", "1200", "--user", "tftp"]
