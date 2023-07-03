# Syncer

This folder contains the Dockerfile for the synchronization logic between the netboot server and our Azure Blob Storage. If you have another storage provider, you can easily adapt the Dockerfile / script to your needs.

We provide a way to determine, if a netboot server should  sync `dev` and/or `prod` images. This is done by checking the `SYNC_DEV` or `SYNC_PROD` environment variable. If the variable is set to `true`, the syncer will sync the images. If the variable is set to `false`, the syncer will not sync the images.

Note: Adjust the `SYNC_BLOB_URL` variable to your needs.
