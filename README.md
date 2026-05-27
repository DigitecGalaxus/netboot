# Introduction

This repo contains all necessary files to easily maintain (and provision) the netboot services:

- tftp: Exposes the initial bootloader as well as the menus for iPXE to work.
- http: Exposes the assets (Filesystems) via HTTP for iPXE to boot.
- cleaner: Takes care of cleaning the assets folder so it won't grow too big.
- monitoring: Monitors the Protocol Endpoints (TFTP / HTTP) and writes them to an Influx DB
- ipxe-menu-generator: Generates IPXE menus for caching servers

## Prerequisites

To run the components of this repository, the following is required:

- A Linux virtual machine with docker and docker-compose (both runnable without sudo permissions), which is accessible from other clients in the network on ports 80 (http) and 69 (tftp)
- A way to upload your squashFS images to the netboot server (e.g. scp)
- Correct directory structure on the netboot server (see [Directory structure](#directory-structure))
- (Optional) A Datadog Instance, if you want to use the monitoring included here

### Directory structure

Directory tree on `~/` of the netboot server:

```tree
├── cleaner.env
├── docker-compose.yaml
├── ipxe-menu-generator.env
├── monitoring.env
└── netboot
    ├── assets
    │   ├── ProbasClient
...
    │   ├── dev
    │   │   ├── 26-04-24-noissue-vector-himmelblau-logs-7e44bb3
    │   │   │   ├── dg-thinclient.squashfs
    │   │   │   ├── initrd
    │   │   │   └── vmlinuz
...
    │   │   └── 26-05-20-SUP-51095-name-mapping-script-4cbfeb2
    │   │       ├── dg-thinclient.squashfs
    │   │       ├── initrd
    │   │       └── vmlinuz
    │   ├── easyscan
    │   │   └── easyscan.zip
    │   ├── healthcheck.json
    │   ├── prod
    │   │   ├── 26-05-19-master-7896901
    │   │   │   ├── dg-thinclient.squashfs
    │   │   │   ├── initrd
    │   │   │   └── vmlinuz
    │   │   ├── 26-05-19-master-c56b550
    │   │   │   ├── dg-thinclient.squashfs
    │   │   │   ├── initrd
    │   │   │   └── vmlinuz
...
    │   │   └── 26-05-26-master-b32afa3
    │   │       ├── dg-thinclient.squashfs
    │   │       ├── initrd
    │   │       └── vmlinuz
    │   └── zero-day-config.py
    └── config
        └── menus
            ├── advancedmenu.ipxe
            ├── menu.ipxe
            └── netinfo.ipxe
```

## How it works

We provision the services using the [docker-compose.yaml](/docker-compose.yaml) file. This requires the docker images to be present on the host. Those can be either pulled from our public registry or built manually. Set the correct environment variables in the `.env` files. Bring your stack up with `docker compose up -d`.

```bash
docker image build -t dgpublicimagesprod.azurecr.io/planetexpress/netboot-tftp:latest ./netboot-services/tftp/
docker image build -t dgpublicimagesprod.azurecr.io/planetexpress/netboot-http:latest ./netboot-services/http/
docker image build -t dgpublicimagesprod.azurecr.io/planetexpress/netboot-cleaner:latest ./netboot-services/cleaner/
docker image build -t dgpublicimagesprod.azurecr.io/planetexpress/netboot-monitoring:latest ./netboot-services/monitoring/
docker image build -t dgpublicimagesprod.azurecr.io/planetexpress/netboot-ipxe-menu-generator:latest ./netboot-services/ipxeMenuGenerator/
```

## Usage

1. Create the directory structure on the netboot server (see [Directory structure](#directory-structure))
2. Upload your image folders (`dg-thinclient.squashfs` + `initrd` + `vmlinuz`) to the netboot server
3. Build or pull the docker images
4. Set the correct environment variables in the `.env` files
5. Bring your stack up with `docker compose up -d`
6. For the monitoring part, check the [README](netboot-services/monitoring/README.md)

## Contribute

No matter how small, we value every contribution! If you wish to contribute,

1. Please create an issue first - this way, we can discuss the feature and flesh out the nitty-gritty details
2. Fork the repository, implement the feature and submit a pull request
3. Your feature will be added once the pull request is merged
