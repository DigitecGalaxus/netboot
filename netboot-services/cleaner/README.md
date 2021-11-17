# Cleaner

This folder contains the Dockerfile for a small container that has a specific folder structure mounted to the `/cleaning` folder. Check the structure in [http/README.md](../http/README.md).

The container continously cleans this folders `prod`, `kernels`, `dev` subfolders, to ensure that the oldest files are deleted from the respective folder. For each folder, a static maximum size is defined, when the size of this folder is exceeded, the oldest file is deleted.
