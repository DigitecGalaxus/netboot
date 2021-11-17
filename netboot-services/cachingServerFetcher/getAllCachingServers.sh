#!/bin/bash
set -u -o pipefail

function getCachingServerIPAndSiteDescription {
    # this URL is passed as environment variable
    # shellcheck disable=SC2154
    networksWithCachingServer=$(curl -s --connect-timeout 2 "$networksServiceURL/networks" -H 'accept: application/json' | jq -rc '.[] | select(.netbootCachingServerIP != null)')

    while IFS= read -r network; do
        site=$(jq -r .description <<< "$network")
        gatewayIP=$(jq -r .gatewayIP <<< "$network")
        cachingServer=$(jq -r .netbootCachingServerIP <<< "$network")
        macAddress=$(curl -s --connect-timeout 5 "$networksServiceURL/gateways/dhcp/lease?DhcpIP=$cachingServer" -H 'accept: application/json' | jq -r .macAddress)
        jq -n -r --arg gatewayIP "$gatewayIP" --arg cachingServer "$cachingServer" --arg site "$site" --arg macAddress "$macAddress" '. | .cachingServerIP=$cachingServer | .site=$site | .macAddress=$macAddress | .gatewayIP=$gatewayIP' <<<'{}'
    done <<<"$networksWithCachingServer"
}

# This runs in an endless loop to ensure that we pick up changes during runtime
while true; do
    # Checking if we can get caching servers at all - if not, no connectivity to the REST endpoint is given and we don't update the server list
    cachingServers=$(curl --connect-timeout 2 -s -X 'GET' "$networksServiceURL/network/netbootCachingServerIPs" -H 'accept: application/json' | jq -r '.[]')

    if [[ "$cachingServers" == "" ]]; then
        echo "Warning: Using fallback static file for the caching servers"
    else
        cachingServerIPAndSiteDescription=$(getCachingServerIPAndSiteDescription)

        echo "$cachingServerIPAndSiteDescription" | jq -s '{cachingServers:[.[]]}' >caching_server_list.json
    fi
    sleep 120
done
