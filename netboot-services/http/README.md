# HTTP Fileserver

This folder contains the Dockerfile for an Nginx container that serves the files needed to netboot. This includes the kernel files, the initial ram disks (initrd) and the squashfs files.

## File Structure on the Netboot Server

The file structure below is used (and served via HTTP) on the netboot server. The dev folder contains untested squashfs files, the prod folder contains tested squashfs files and the kernels folder contains the kernel and initial ramdisks.

```txt
└── assets/
    ├── dev/
    │   ├── [flavor]-[branchname]-[date]-[shortsha].squashfs
    │   └── [flavor]-[branchname]-[date]-[shortsha]-kernel.json
    ├── prod/
    │   ├── [flavor]-[branchname]-[date]-[shortsha].squashfs
    │   └── [flavor]-[branchname]-[date]-[shortsha]-kernel.json
    └── kernels/
        ├── newest-kernel-version.json
        └── [kernel-version]/
            ├── vmlinuz
            └── initrd
```
