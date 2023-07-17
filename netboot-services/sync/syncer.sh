#!/bin/bash

#import environment variables

if [ -z "$SYNC_SAS_TOKEN" ]; then
  echo "SAS_TOKEN is not set"
  exit 1
fi

if [ -z "$SYNC_BLOB_URL" ]; then
  echo "SYNC_BLOB_URL is not set"
  exit 1
fi

if [ -z "$SYNC_BANDWITDH_LIMIT_MBITS" ]; then
  echo "SYNC_BANDWITDH_LIMIT is not set"
  SYNC_BANDWITDH_LIMIT_MBITS=300
fi

while true
do
  if [ "$SYNC_DEV" = "true" ] || [ "$SYNC_DEV" = "True" ]; then
    azcopy cp --cap-mbps "$SYNC_BANDWITDH_LIMIT_MBITS" "$SYNC_BLOB_URL/dev/*$SYNC_SAS_TOKEN" "/home/syncer/dev/" --overwrite=ifSourceNewer --recursive
  fi

  if [ "$SYNC_PROD" = "true" ] || [ "$SYNC_PROD" = "True" ]; then
  azcopy cp --cap-mbps "$SYNC_BANDWITDH_LIMIT_MBITS" "$SYNC_BLOB_URL/prod/*$SYNC_SAS_TOKEN" "/home/syncer/prod/" --overwrite=ifSourceNewer --recursive
  fi

  #Sync kernels
  azcopy cp --cap-mbps "$SYNC_BANDWITDH_LIMIT_MBITS" "$SYNC_BLOB_URL/kernels/*$SYNC_SAS_TOKEN" "/home/syncer/kernels/" --overwrite=ifSourceNewer --recursive

  #random sleep within 5 minutes
  sleep $(( ( RANDOM % 300 )  + 1 ))
  echo "Syncing run processed on $(date)"
done
