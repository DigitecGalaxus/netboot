# Monitoring

We monitor the tftp service from the host it is running on to be able to use docker-compose networking. This has the benefit of not having to map the tftp ports: the initial connect happens via port 69, then it connects via UDP on a random port in the range 30000+. This is difficult to handle from within a docker container (e.g. within a Kubernetes cluster), as all ports above 30000 would need to be mapped to this container.

## Usage

To get the telegraf config to work, set the correct environment variables in the `.env` file. It is setup to log to Datadog only when the respective parameters are provided. If not, it will simply log to stdout with the `telegraf --test` option.

## TFTP Monitoring

The script [monitor_tftp.sh](monitor_tftp.sh) checks for the presence of the files, which are initially used to boot from the network. If those are not available, no client is able to boot. It also actually downloads those files (and discards them) to check for potential issues with downloading as well. As this runs on the same host as the TFTP server, this will actually not congest the network.

## HTTP Monitoring

The script [monitor_serveravailability.sh](monitor_serveravailability.sh) checks whether the files needed to boot are available and accessible on the netboot server's HTTP endpoint. It reads the advertised prod version from `healthcheck.json` and verifies that the corresponding `healthcheck.json` and prod image file can actually be downloaded.
