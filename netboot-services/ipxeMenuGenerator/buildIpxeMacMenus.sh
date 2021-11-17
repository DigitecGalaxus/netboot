#!/bin/bash
set -eu -o pipefail

# This function generates an IPXE menu for each caching server. As the caching servers themselves use network booting, the MAC address for each needs to be configured to boot into the caching-server.squashfs
function buildCachingServerIPXEMenus {
    netbootIP="$1"
    netbootMenusFolder="$2"

    # Setting the specific kernel version
    # There is a file on the netboot server, which describes the latest kernel-version. This version will be used to boot the caching server
    kernelVersion=$(curl -s --connect-timeout 2 "http://$netbootIP/kernels/latest-kernel-version.json" | jq -r .version)
    if [[ "$kernelVersion" == "" ]]; then
        echo "Error: could not determine latest kernel version"
        exit 1
    fi
    # Replace the kernelVersion and netbootIP in the ipxe file
    sed "s/kernelVersionPlaceholder/$kernelVersion/g" caching-server.ipxe.tmpl >caching-server.ipxe
    sed -i "s/netbootIPPlaceholder/$netbootIP/g" caching-server.ipxe

    cachingServerIPs=$(jq -r '.cachingServers[].cachingServerIP' < caching_server_list.json)

    for cachingServerIP in $cachingServerIPs; do
        echo "Copying MAC IPXE file for $cachingServerIP"
        macAddress=$(jq -r '.cachingServers[] | select(.cachingServerIP=="'"$cachingServerIP"'") | .macAddress' < caching_server_list.json)
        # We need a specific format of the mac address: no separators between the segments
        macAddress=${macAddress//:/}

        cp ./caching-server.ipxe "$netbootMenusFolder/MAC-$macAddress.ipxe"
    done
}

# This is passed as environment variable
# shellcheck disable=SC2154
netbootIP="$netbootServerIP"
netbootMenusFolder="/work/menus"

# This runs in an endless loop to ensure that we pick up changes during runtime
while true; do
    buildCachingServerIPXEMenus "$netbootIP" "$netbootMenusFolder"
    sleep 120
done
