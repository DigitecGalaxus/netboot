# Syncer

This folder contains the Dockerfile for the synchronization logic between the main netboot server and its caching servers.

## Synchronization caveats

For the synchronization logic, a few considerations have to be made.

1. *Limited bandwitdh*: The size of the files that are synchronized as well as the number of targets have to be thought of.

    With the caching server setup, we would like to cache all files as quickly as possible. As we have different internet speeds at the caching server locations (some are quite slow), we don't want those to block the other locations. This is why we synchronize in parallel to all locations. This has worked fine so far.
2. *Partial files*: What happens, if a file is synchronized only partially?

    This is an issue for us - With a partial file (this can happen, e.g. when a file is initially uploaded to the main netboot server) a client might attempt to boot from it, which will obviously fail. We want to prevent synchronizing partial files to our caching hosts, as this would only worsen the problem. We noticed that the partial file upload is only an issue with rather large files (>100MB), so we introduced an additional descriptive json file for each squashfs file that is going to be synchronized. Only once the json file is uploaded, we start with the synchronization of the squashfs file to the caching server and once this is complete, we synchronize the json file as well. When booting from a caching server, we check for the presence of the json file, only once the json file is present we can boot this specific squashfs file from the caching server.

## Trigger a resync

It can happen that a full resync of a caching server is needed. For instance when you add a new one. This can be done (on the server hosting the Syncer Docker container) using the following command:

```bash
docker exec -it netboot-syncer rsync -avh --exclude=".*" -e "ssh -o 'StrictHostKeyChecking=no' -i /ssh/caching-server.pem" /syncing/ "$cachingServerUsername@<caching-server-IP-here>:$cachingServerAssetsDirectory/"
```
