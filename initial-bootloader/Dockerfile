FROM alpine:3.14.0
# Install dependencies
RUN apk add --no-cache bash gcc binutils make perl xz-libs xz-dev mtools git libc-dev
# Clone iPXE
RUN git clone git://git.ipxe.org/ipxe.git
# Include custom logic
COPY custom.ipxe ipxe/src/custom.ipxe
# Run the builds
# Note: We're using snponly to retain the original UEFI-drivers and thus improving reliability and reducing the size of the bootloader.
RUN cd ipxe/src && make bin/undionly.kpxe EMBED=custom.ipxe && \
make bin-i386-efi/snponly.efi EMBED=custom.ipxe && \
make bin-x86_64-efi/snponly.efi EMBED=custom.ipxe
