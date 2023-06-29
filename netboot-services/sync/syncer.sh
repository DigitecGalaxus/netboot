#!/bin/bash

#import environment variables

if [ -z "$SYNC_SAS_TOKEN" ]; then
  echo "SAS_TOKEN is not set"
  exit 1
fi

while true
do
  if [ "$SYNC_DEV" ]; then
    azcopy cp "https://thinclientsimgstore.blob.core.windows.net/dev/*$SYNC_SAS_TOKEN" "/home/syncer/dev/" --overwrite=ifSourceNewer --recursive
  fi
 
  azcopy cp "https://thinclientsimgstore.blob.core.windows.net/prod/*$SYNC_SAS_TOKEN" "/home/syncer/prod/" --overwrite=ifSourceNewer --recursive

  #Sync kernels
  azcopy cp "https://thinclientsimgstore.blob.core.windows.net/kernels/*$SYNC_SAS_TOKEN" "/home/syncer/kernels/" --overwrite=ifSourceNewer --recursive

  #random sleep within 5 minutes
  sleep $(( ( RANDOM % 300 )  + 1 ))
  echo "Syncing run processed on $(date)"
done
