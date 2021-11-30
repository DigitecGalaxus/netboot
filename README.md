# Introduction

# blub

This repo contains all necessary files to easily maintain (and provision) the netboot services:

- tftp: Exposes the initial bootloader as well as the menus for iPXE to work.
- http: Exposes the assets (Filesystems) via HTTP for iPXE to boot.
- sync: Takes care of syncing the assets to the caching servers.
- cleaner: Takes care of cleaning the assets folder so it won't grow too big.
- monitoring: Monitors the Protocol Endpoints (TFTP / HTTP) and writes them to an Influx DB
- caching-server-fetcher: Provides a list of all caching servers
- ipxe-menu-generator: Generates IPXE menus for caching servers

[![Build Status](https://dev.azure.com/digitecgalaxus/SystemEngineering/_apis/build/status/Github/DigitecGalaxus.netboot?repoName=DigitecGalaxus%2Fnetboot&branchName=main)](https://dev.azure.com/digitecgalaxus/SystemEngineering/_build/latest?definitionId=1182&repoName=DigitecGalaxus%2Fnetboot&branchName=main)

## Prerequisites

To run the components of this repository, the following is required:

- A Linux virtual machine with docker and docker-compose (both runnable without sudo permissions), which is accessible from other clients in the network on ports 80 (http) and 69 (tftp)
- A configured private key to access this virtual machine via SSH. This SSH key will be needed in the thin-client repository (TODO: link) and the caching server repository (TODO: link) to upload squashfs files and modify IPXE menus.
- In case you want to use caching servers, a private key to access them. See the Usage section on how to use it. If you want to skip caching servers, you will need to modify some parts (syncer is not needed)
- (Optional) A Datadog Instance, if you want to use the monitoring included here

## How it works

We provision the five services using the [docker-compose.yaml](/docker-compose.yaml) file. This requires the docker images to be present on the host. The [run-compose.sh](run-compose.sh) contains the instructions to build them.

## Usage

1. Decide if you want to use caching servers or not.
    1. If you want to use caching servers, prepare the private key for the caching servers and specify it's full path as an argument to the run-compose.sh script. Additionally, check the cachinigServerFetcher and fill out the file [netboot-services/cachingServerFetcher/caching_server_list.json](netboot-services/cachingServerFetcher/caching_server_list.json) with the details.
    2. If you don't want to use caching servers, there are some components in this repository that are not necessary. From the docker-compose file, the netboot-syncer, the netboot-ipxe-menu-generator and the caching-server-fetcher can be removed. In this case, the pemFilePath does not need to be specified as argument to the [run-compose.sh](run-compose.sh) script.
2. On the netboot server, create the folders for the volume mounts and start the services with the [run-compose.sh](run-compose.sh) script. As first parameter, pass the IP of the server, where the containers are running to the script. This IP needs to be reachable from the clients that network-boot. Additionally, you can pass more optional parameters, refer to the script for more information.
3. For the monitoring part, check the [README](netboot-services/monitoring/README.md)

## Contribute

No matter how small, we value every contribution! If you wish to contribute,

1. Please create an issue first - this way, we can discuss the feature and flesh out the nitty-gritty details
2. Fork the repository, implement the feature and submit a pull request
3. Your feature will be added once the pull request is merged
