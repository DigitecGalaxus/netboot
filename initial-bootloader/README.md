# Introduction

This folder contains all necessary files to build a custom bootloader with custom logic.

## How it works

The DHCP server usually has some additional configuration for network booting. In this configuration you can specify the IP of your network boot server and the UEFI 32/64bit files as well as the default bios filename (undionly.kpxe in our case). When a client in the DHCP servers network boots, the DHCP server returns this information to the client, which will then contact the network boot server and request the files generated in this docker build. The [custom.ipxe](custom.ipxe) file will then leave the further netbooting logic to the menu.ipxe.

This bootloader basically redirects everything to the netboot-server (the file menu.ipxe served by the [netboot-tftp](../netboot-services/tftp/) container) to process the logic there:

1. Get a dynamic IP address (retry indefinitely; this helps to make sure the clients boot, even if the network connection is not ready yet, as the network hardware might be reinitialized before the ipxe menu flow).
2. Load the initial menu (menu.ipxe) from the netboot server. The menu.ipxe is placed onto the netboot server by the [thin-client repository](https://dev.azure.com/DigitecGalaxus/SystemEngineering/_git/ThinClient?path=%2Fpromote%2Fgenerate-new-ipxe-menus.sh)
TODO: replace this link once thin-client was opensourced

## Getting Started to build a new version

In order to build a new bootloader, please consult the official building documentation:

- [Source Code & Make](https://ipxe.org/download)
- [Embedded Scripts](https://ipxe.org/embed)

The included [dockerbuild.sh](dockerbuild.sh) automates the entire build. Simply make the necessary changes to custom.ipxe and build the bootloader by executing [dockerbuild.sh](dockerbuild.sh). The only requirement is a Docker daemon and internet access.
The following procedure can be used as a guideline:

1. Execute the script `./dockerbuild.sh`.
2. Copy the compiled bootloaders via scp to the netboot server: /$HOME/netboot/config/menus/

## Usage

1. Adjust the custom.ipxe to match your environment
2. Place the files undionly.kpxe, ipxe32.efi, ipxe64.efi onto your netboot server.
3. Configure netbooting on your DHCP server to point to your netboot server and the respective files.
