FROM alpine:3.14.0 as bootfiles
RUN apk add --no-cache bash gcc binutils make perl xz-libs xz-dev mtools git libc-dev
WORKDIR /build
# Clone iPXE
RUN git clone git://git.ipxe.org/ipxe.git
# Include custom logic
COPY custom.ipxe ipxe/src/custom.ipxe
# Run the builds
# Note: We're using snponly to retain the original UEFI-drivers and thus improving reliability and reducing the size of the bootloader.
RUN cd ipxe/src && make bin/undionly.kpxe EMBED=custom.ipxe && \
make bin-i386-efi/snponly.efi EMBED=custom.ipxe && \
make bin-x86_64-efi/snponly.efi EMBED=custom.ipxe

FROM alpine:3.14.0

RUN apk add --no-cache tftp-hpa
RUN adduser -S tftp -s /bin/ash -h /home/tftp
WORKDIR /home/tftp
USER tftp

COPY --chown=tftp:nogroup --from=bootfiles /build/ipxe/src/bin/undionly.kpxe /home/tftp/undionly.kpxe
COPY --chown=tftp:nogroup --from=bootfiles /build/ipxe/src/bin-i386-efi/snponly.efi /home/tftp/ipxe32.efi
COPY --chown=tftp:nogroup --from=bootfiles /build/ipxe/src/bin-x86_64-efi/snponly.efi /home/tftp/ipxe64.efi


ENTRYPOINT ["/usr/sbin/in.tftpd"]
CMD ["--foreground", "--listen", "--secure", "/home/tftp", "--blocksize", "1200"]
