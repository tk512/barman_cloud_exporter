ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
LABEL maintainer=""

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/barman_cloud_exporter /bin/barman_cloud_exporter

EXPOSE      61092
USER        nobody
ENTRYPOINT  [ "/bin/barman_cloud_exporter" ]
