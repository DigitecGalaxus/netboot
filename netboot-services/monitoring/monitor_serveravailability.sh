#!/bin/bash
set -u

function getFilenameWithFilter {
    serverURL=$1
    serverFolderPath=$2
    curlFilter=$3

    # Try to find a filename in a folder on the HTTP server, filter with $3 and cut the quotes
    retrievedFileOnServerToCheck=$(curl -s "$serverURL/$serverFolderPath/" --connect-timeout 2 --max-time 3 | grep "$curlFilter" | cut -d '"' -f2 | tail -1)
    echo "$retrievedFileOnServerToCheck"
}

function requestFilesAndEchoInfluxOutput() {
    serverURL=$1
    serverFolderPath=$2
    fileName=$3

    # Assemble the full URL to the file
    completeURLtoCheck="$serverURL/$serverFolderPath/$fileName"
    # Download the file and get the status code of the download
    curlStatusCode=$(curl -s --connect-timeout 2 --max-time 3 -w "%{http_code}\n" "$completeURLtoCheck" -o "/tmp/$fileName")

    # Check if the downloaded file exists and has size greater than zero (-s) and we were actually able to determine a valid file name (-n)
    if [[ -s "/tmp/$fileName" ]] && [[ -n "$fileName" ]]; then
        httpFileIsAvailable="1"
        rm "/tmp/$fileName"
    else
        httpFileIsAvailable="0"
    fi

    formatInfluxData "$serverURL" "$fileName" "$httpFileIsAvailable" "$curlStatusCode"
}

function formatInfluxData() {
    serverURL=$1
    fileName=$2
    httpFileIsAvailable=$3
    curlStatusCode=$4

    echo files_available,url_base="$serverURL" file=\""$fileName"\",file_available="$httpFileIsAvailable",code="$curlStatusCode"
}

netbootServer="$1"
# This references an internal endpoint and might not be reachable.
latestKernelversion=$(curl --connect-timeout 2 -s "$netbootServer/kernels/latest-kernel-version.json" | jq -r .version)

if [[ "$latestKernelversion" == "" ]]; then
    echo "Error: could not determine latest kernel version" >>/dev/stderr
    exit 1
fi

# Get the latest filename with json in it's name
mostRecentKernelVersionJson=$(getFilenameWithFilter "$netbootServer" "prod" "json")

### Check HTTP functionality on Netboot Server with a small json file to avoid downloading the full squashfs. Also Test the Kernel Files and do execute the same on the cachingServers
requestFilesAndEchoInfluxOutput "$netbootServer" "kernels/$latestKernelversion" "vmlinuz"
requestFilesAndEchoInfluxOutput "$netbootServer" "prod" "$mostRecentKernelVersionJson"
