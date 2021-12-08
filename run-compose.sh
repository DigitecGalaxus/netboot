#!/bin/bash
set -e

if [[ $# -lt 1 ]]; then
        echo "Error: Please specify the argument 'netbootServerIP' and optionally 'devCachingServerIP','pemFilePath','netbootServicesPullToken'. Exiting..."
        exit 2
else
        # The IP of the netboot server, where clients can reach it
        export netbootServerIP="$1"
        # Optional: The file path to the private key file to ssh/scp to the caching servers
        # If no value is provided, no caching servers will be used
        pemFilePath="$2"
        # Optional: The caching server for the dev/ folder
        # If no value is provided, the dev folder will not be synced
        export devCachingServerIP="$3"
        # Optional: The pull token to pull the docker images
        # If no value is provided, the images will be built locally
        netbootServicesPullToken="$4"
fi

# Skip pulling, if no credentials were provided and build it locally instead
if [[ -n "$netbootServicesPullToken" ]]; then
        echo "Logging into Docker registry and pulling latest images"
        docker login --username netboot-services-pull-token --password "$netbootServicesPullToken" anymodconrst001dg.azurecr.io
        docker-compose pull
else
        docker image build -t anymodconrst001dg.azurecr.io/planetexpress/netboot-tftp:latest ./netboot-services/tftp/
        docker image build -t anymodconrst001dg.azurecr.io/planetexpress/netboot-http:latest ./netboot-services/http/
        docker image build -t anymodconrst001dg.azurecr.io/planetexpress/netboot-sync:latest ./netboot-services/sync/
        docker image build -t anymodconrst001dg.azurecr.io/planetexpress/netboot-cleaner:latest ./netboot-services/cleaner/
        docker image build -t anymodconrst001dg.azurecr.io/planetexpress/netboot-monitoring:latest ./netboot-services/monitoring/
        docker image build -t anymodconrst001dg.azurecr.io/planetexpress/caching-server-fetcher:latest ./netboot-services/cachingServerFetcher/
        docker image build -t anymodconrst001dg.azurecr.io/planetexpress/netboot-ipxe-menu-generator:latest ./netboot-services/ipxeMenuGenerator/
fi

set -u

# Create the folders for the volume mounts to make sure they persist restarts and deployments
# These folders will be mounted into the different docker images
mkdir -p "$HOME"/netboot/config/menus/
mkdir -p "$HOME"/netboot/assets/prod/
mkdir -p "$HOME"/netboot/assets/dev/
mkdir -p "$HOME"/netboot/assets/kernels/

# Initial creation of latest-kernel-version.json
if [[ ! -f "$HOME"/netboot/assets/kernels/latest-kernel-version.json ]]; then
        echo '{ "version": "5.8.0-43-generic" }' >"$HOME"/netboot/assets/kernels/latest-kernel-version.json
fi

cp ./netboot-services/cachingServerFetcher/caching_server_list.json "$HOME"/netboot/caching_server_list.json

if [[ -n "$pemFilePath" ]]; then
        # Copy the key to where the syncer expects it
        cp -f "$pemFilePath" "$HOME"/netboot/caching-server.pem
        chmod 600 "$HOME"/netboot/caching-server.pem
fi

# create empty caching_server_list.json to be mountable by docker
touch "$HOME"/netboot/caching_server_list.json

# Copy any custom menus into the folder
cp ./custom-menus/* "$HOME"/netboot/config/menus/

# Start the services
docker-compose up -d
