# Caching Server Fetcher

This folder contains the Dockerfile for a utility, that provides details of the caching servers to the syncer and monitoring containers. It queries an internal endpoint (`networksServiceURL`), and if it's reachable it dynamically reads the list of caching servers from there. If it's not reachable, it uses the file [caching_server_list.json](caching_server_list.json) as fallback.

## Usage

The file [caching_server_list.json](caching_server_list.json) should be populated with all caching servers. If no caching server is used, simply create a file with this content:

```json
{
  "cachingServers": []
}
```

If you use caching servers, populate it with the IPs and a description (site) of them as shown in the example file.
