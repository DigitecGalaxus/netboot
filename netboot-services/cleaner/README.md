# Cleaner

This folder contains the Dockerfile for a small container that has a specific folder structure mounted to the `/cleaning` folder. Check the structure in [http/README.md](../http/README.md).

The container continously cleans this folders `prod`, `dev` subfolders, to ensure that the oldest files are deleted from the respective folder. For each folder, a static maximum size is defined, when the size of this folder is exceeded, the oldest file is deleted. Also the maximum count of images is defined, when this count is exceeded, the oldest file is deleted. The check is done every 5 minutes.

To locally test the container, run the following command:

```bash
docker run --rm -d -v $(pwd):/cleaning  --name netboot-cleaner dgpublicimagesprod.azurecr.io/planetexpress/netboot-cleaner
```
