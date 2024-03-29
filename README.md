# Barman Cloud Exporter

Prometheus exporter for Barman Cloud WAL archiving and Backups

Requirements:

* barman-cli-cloud package (see www.pgbarman.org)
* PostgreSQL 10 or higher

## Building and running

### Build

    make build

### Running

##### Single exporter mode

Only single exporter mode is available in this application. 

On the prometheus side you can set a scrape config as follows

        - job_name: ...

## Using Docker

You can deploy this exporter using the Docker image.

For example:

```bash
docker network create my-pg-network
docker pull TBD

docker run -d \
  -p 61092:61092 \
  --network my-pg-network  \
  TBD
```

## Using systemd

See examples/systemd/README.md