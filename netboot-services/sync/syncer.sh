#!/bin/bash

#import environment variables

if [ -z "$SAS_TOKEN" ]; then
  echo "SAS_TOKEN is not set"
  exit 1
fi

while true
do
  if [ "$SYNC_DEV" ]; then
    azcopy sync "https://thinclientsimgstore.blob.core.windows.net/dev/$SAS_TOKEN" "/home/syncer/"
  fi

  azcopy sync "https://thinclientsimgstore.blob.core.windows.net/prod/$SAS_TOKEN" "/home/syncer/"

  #random sleep every 5 minutes
  sleep $(( ( RANDOM % 300 )  + 1 ))
  echo "Syncing run processed on $(date)"
done
