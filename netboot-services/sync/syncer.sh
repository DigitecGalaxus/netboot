#!/bin/bash

# Sync one file to one host
function sync() {
  hostToSync="$1"
  pathToSync="$2"
  fileToSync="$3"

  echo "$(date): Started to sync $pathToSync$fileToSync to $hostToSync."
  # Syncing that specific file except if it's a hidden file (the `.` is for hidden files, the `*` is a wildcard).
  # Ignoring variable used but not assigned, cachingServerUsername and cachingServerAssetsDirectory should be set as environment variable
  # shellcheck disable=SC2154
  rsync -avh --exclude=".*" -e "ssh -o 'StrictHostKeyChecking=no' -i /ssh/caching-server.pem" "$pathToSync$fileToSync" "$cachingServerUsername@$hostToSync:$cachingServerAssetsDirectory/$pathToSync"
  echo "$(date): Sync of $pathToSync$fileToSync to $hostToSync finished."
}

# We always have these two files along side each other. Only once the -kernel.json file is present, we can start to sync both files to the caching servers. This is due to partial file uploads of the rather large .squashfs file. The boot via caching server only boots from the caching server, if the json file is present. Therefore it's important that we sync the json file after the squashfs file.
# 21-04-09-gnome-master-f1d734e.squashfs
# 21-04-10-gnome-master-f1d734e-kernel.json
function syncKernelJsonThenSquashfs {
  hostToSync="$1"
  pathToSync="$2"
  kernelJsonFileToSync="$3"
  squashfsFileToSync="$4"

  sync "$hostToSync" "$pathToSync" "$squashfsFileToSync"
  sync "$hostToSync" "$pathToSync" "$kernelJsonFileToSync"
}

function syncFileDeletion {
  hostToSync="$1"
  pathToSync="$2"
  echo "$(date): Started to delete files for path $pathToSync on $hostToSync."
  # Syncing with --delete flag and without specific file. This way, deleted files get removed.
  # Making sure this rsync only deletes files: https://serverfault.com/questions/275493/using-rsync-to-only-delete-extraneous-files/713577#713577
  rsync -avh --exclude=".*" --delete --existing --ignore-existing -e "ssh -o 'StrictHostKeyChecking=no' -i /ssh/caching-server.pem" "$pathToSync" "$cachingServerUsername@$hostToSync:$cachingServerAssetsDirectory/$pathToSync"
  echo "$(date): Done with deleting files for path $pathToSync on $hostToSync."
}

if [[ -d /ssh/caching-server.pem ]]; then
  echo "caching server file path is a directory, did you specify this correctly? Exiting..."
  exit 1
fi

set -u

# Starting the inotify-loop and triggering syncs to all caching servers.
echo "Starting filesystem change monitoring based on Inotify"
# Inotify outputs one line for each file change and outputs path, action and file. So whenever a file changes, is deleted or a new file appears, we're notified
inotifywait -m -r -e close_write,delete prod/ kernels/ dev/ caching-server/ |
  while read -r path action file; do

    allCachingServers=$(jq -r '.cachingServers[].cachingServerIP' </ssh/caching_server_list.json)

    echo "The file '$file' appeared in directory '$path' via '$action'"
    # When we have a new file in the dev/ folder, we only want to synchronize it to a single caching server, as there we have development VMs on Proxmox for testing.
    if [ "$path" = "dev/" ]; then
      # This variable is passed as environment variable
      cachingServers=("$devCachingServerIP")
    else
      cachingServers=("$allCachingServers")
    fi

    while IFS= read -r cachingServer; do
      # This service also provides additional information for each network - here we want a descriptive name for better logs
      site=$(jq -r '.cachingServers[] | select(.cachingServerIP == "'"$cachingServer"'").site' </ssh/caching_server_list.json)
      echo "Going to sync to caching server with IP $cachingServer on site $site"
      if [[ "$action" == "DELETE" ]]; then
        syncFileDeletion "$cachingServer" "$path" &
        continue
      fi
      # From now on, we only handle the action "CLOSE_WRITE,CLOSE"
      if [[ "$file" == "netboot-caching-server.squashfs" ]]; then
        echo "$(date): Syncing $file."
        # We can synchronize those in the background in parallel, to avoid having one slow connection block the other transfers
        sync "$cachingServer" "$path" "$file" &
        continue
      fi
      if [[ "$file" == *.squashfs ]]; then
        echo "$(date): Not syncing $file specifically, will be synced with the -kernel.json file."
        continue
      fi
      # Some standard files we can just simply sync to all caching servers, as they are small and they have no dependency on other files
      if [[ "$file" == "latest-kernel-version.json" ]] || [[ "$file" == "initrd" ]] || [[ "$file" == "vmlinuz" ]]; then
        # We can synchronize those in the background in parallel, to avoid having one slow connection block the other transfers
        sync "$cachingServer" "$path" "$file" &
        continue
      fi
      if [[ "$file" == *-kernel.json ]]; then
        # Determine the corresponding .squashfs filename
        squashfsFileName="${file%%-kernel.json}.squashfs"
        if [[ -f "$path$squashfsFileName" ]]; then
          # We can synchronize those in the background in parallel, to avoid having one slow connection block the other transfers
          syncKernelJsonThenSquashfs "$cachingServer" "$path" "$file" "$squashfsFileName" &
        else
          echo "$(date): Error: could not find matching squashfs file for $path$file, not syncing it."
        fi
      fi
    done <<<"${cachingServers[@]}" # end caching for loop
  done                             # end while loop
