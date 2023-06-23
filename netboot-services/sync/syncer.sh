#!/bin/bash

#import environment variables

if [ -z "$SYNC_SAS_TOKEN" ]; then
  echo "SAS_TOKEN is not set"
  exit 1
fi

while true
do
  if [ "$SYNC_DEV" ]; then
    azcopy sync "https://thinclientsimgstore.blob.core.windows.net/dev/$SYNC_SAS_TOKEN" "/home/syncer/dev/" --delete-destination=true
  fi

  azcopy sync "https://thinclientsimgstore.blob.core.windows.net/prod/$SYNC_SAS_TOKEN" "/home/syncer/prod/" --delete-destination=true

  #Sync kernels
  azcopy sync "https://thinclientsimgstore.blob.core.windows.net/kernels/$SYNC_SAS_TOKEN" "/home/syncer/kernels/" --delete-destination=true

  #random sleep within 5 minutes
  sleep $(( ( RANDOM % 300 )  + 1 ))
  echo "Syncing run processed on $(date)"
done
