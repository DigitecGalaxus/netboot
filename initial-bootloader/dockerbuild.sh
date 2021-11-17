#!/bin/bash
set -eu
docker image build -t temp_bootloader-image .
docker create --name extract temp_bootloader-image
docker cp extract:/ipxe/src/bin/undionly.kpxe ./undionly.kpxe
docker cp extract:/ipxe/src/bin-i386-efi/snponly.efi ./ipxe32.efi
docker cp extract:/ipxe/src/bin-x86_64-efi/snponly.efi ./ipxe64.efi
docker rm extract
docker rmi temp_bootloader-image