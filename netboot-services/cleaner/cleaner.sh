#!/bin/bash
set -eu

# Defining function
function cleanFolder() {
  folderToClear="$1"
  maximumFolderSizeInGB="$2"
  # See: https://askubuntu.com/a/177939/1167561
  # $9 is the file name
  # $5 is the file size in bytes
  # "test -f" returns 0, if it is a file and not a directory. invert it to trigger the if condition
  # The below command lists all files in the folder by size and orders them by modification time (newest first). It accumulates the file sizes and removes all files that exceed the maximum folder size.
  # Ignoring this shellcheck, as this will work fine for this use case
  # shellcheck disable=SC2012
  cd "$folderToClear" && ls -ltc | awk '{ if (!system("test -f " $9)) { size += $5; if (size > '"$maximumFolderSizeInGB"'*2^30 ) { system("rm " $9); printf "%s will be deleted\n",$9 } } }'
}

# Triggering the function in a timed loop, so we don't need cronjobs
sleepTime="15"
echo "Starting $sleepTime seconds-cleaning-loop..."
while true; do
  # To understand the folder structure, refer to the README.md at netboot-services/http/README.md
  # Clean the folder prod, such that it is smaller than 5GBs in size after the cleanup.
  cleanFolder /cleaning/prod 5
  cleanFolder /cleaning/dev 5
  cleanFolder /cleaning/kernels 1
  sleep $sleepTime
done
