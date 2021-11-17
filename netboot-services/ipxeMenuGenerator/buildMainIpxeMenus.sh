#!/bin/bash
set -eo pipefail

# This script reads the contents of the assets directory and builds IPXE menus for the content.

# Outputs the image names and belonging kernel versions in json-like format
function getKernelVersionPerImageAndEnvironment {
    environment="$1"
    allImages="images.json"
    find "/assets/$environment" -name "*.json" -printf '%f\n' >"$allImages"
    for row in $(cat $allImages); do
        kernelVersion=$(jq -r .version < "/assets/$environment/$row")
        imageName="${row%-kernel.json}"
        jq -n -r --arg imageName "$imageName" --arg kernelFolderName "$kernelVersion" '{"imageName": $imageName, "kernelFolderName": $kernelFolderName}'
    done
    rm "$allImages"
}

# Builds the menu.ipxe based on the information provided by caching_server_list.json and files on the file system
function buildMenuIpxe {
        # The latest squashfs file is in this folder
        folderName="/assets/prod/"
        # Sort the list of .squashfs files by date and pick the newest filename
        mostRecentSquashfsFilename=$(ls -ltc "$folderName"*squashfs | head -1 | awk '{print $9}')
        # The name of the squashfs without file extension
        imageName="${mostRecentSquashfsFilename%.squashfs}"
        # Remove the folder name
        imageName="${imageName#$folderName}"
        # The kernel version that this squashfs was tested with
        kernelVersion=$(jq -r .version < "$folderName${imageName}-kernel.json")
        # Assemble all information in JSON format for templating
        kernelJSON=$(jq -n -r --arg imgName "$imageName" --arg kernFolderName "$kernelVersion" '{"imageName": $imgName, "kernelFolderName": $kernFolderName}')
        cachingServerJSON=$(cat caching_server_list.json)
        netbootServerJSON=$(jq -n -r --arg netbootServerIP "$netbootServerIP" '{"netbootServerIP": $netbootServerIP}')
        jq --argjson netbootServerIP "$netbootServerJSON" --argjson serverList "$cachingServerJSON" --argjson kernelList "$kernelJSON" -n '$serverList + $kernelList + $netbootServerIP' > menu-parameters.json
        # Use Jinja2 to create the final file from the template and parameters
        j2 -f json menu.ipxe.j2 menu-parameters.json > menu.ipxe
}

# Builds the advancedmenu.ipxe based on the files in the dev and prod folders
function buildAdvancedMenuIpxe {
        prodJSON=$(getKernelVersionPerImageAndEnvironment prod | jq -r -s  '{prod:[.[]]}')
        devJSON=$(getKernelVersionPerImageAndEnvironment dev | jq -r -s '{dev:[.[]]}')
        netbootServerJSON=$(jq -n -r --arg netbootServerIP "$netbootServerIP" '{"netbootServerIP": $netbootServerIP}')

        #Merge Both JSON's
        jq --argjson netbootServerIP "$netbootServerJSON" --argjson prod "$prodJSON" --argjson dev "$devJSON" -n '$prod + $dev + $netbootServerIP' > advanced-menu-parameters.json
        j2 -f json advancedmenu.ipxe.j2 advanced-menu-parameters.json > advancedmenu.ipxe
}

# This is passed as environment variable
# shellcheck disable=SC2154
netbootServerIP="$netbootServerIP"

set -u

while true; do
        buildMenuIpxe
        mv menu.ipxe /menus/menu.ipxe
        buildAdvancedMenuIpxe
        mv advancedmenu.ipxe /menus/advancedmenu.ipxe
        sleep 120
done
