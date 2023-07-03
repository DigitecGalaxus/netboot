FROM alpine:3.14.0 as bootfiles
RUN apk add --no-cache bash gcc binutils make perl xz-libs xz-dev mtools git libc-dev
WORKDIR /build
# Clone iPXE
RUN git clone git://git.ipxe.org/ipxe.git
RUN sed -i 's/#undef\tDOWNLOAD_PROTO_HTTPS/#define\ DOWNLOAD_PROTO_HTTPS/' ipxe/src/config/general.h
# Include custom logic
COPY custom.ipxe ipxe/src/custom.ipxe
# Run the builds
# Note: We're using snponly to retain the original UEFI-drivers and thus improving reliability and reducing the size of the bootloader.
RUN cd ipxe/src && make bin/undionly.kpxe EMBED=custom.ipxe && \
make bin-i386-efi/snponly.efi EMBED=custom.ipxe && \
make bin-x86_64-efi/snponly.efi EMBED=custom.ipxe

FROM alpine:3.14.0

RUN apk add --no-cache tftp-hpa curl
RUN mkdir -p /srv/tftp/ipxe

COPY  --from=bootfiles /build/ipxe/src/bin/undionly.kpxe /srv/tftp/undionly.kpxe
COPY  --from=bootfiles /build/ipxe/src/bin-i386-efi/snponly.efi /srv/tftp/ipxe32.efi
COPY  --from=bootfiles /build/ipxe/src/bin-x86_64-efi/snponly.efi /srv/tftp/ipxe64.efi

RUN adduser -D tftp -H
EXPOSE 69/udp

ENTRYPOINT ["/usr/sbin/in.tftpd"]
CMD ["--foreground", "--listen", "--secure", "/srv/tftp", "--blocksize", "1200", "--user", "tftp"]
