# Systemd unit

## Installation of service unit
To install this systemd unit on e.g. Ubuntu Linux, first put the *.service file
in the /etc/systemd/service directory.

```
$ sudo systemctl daemon-reload
$ sudo systemctl enable barman-cloud-exporter
```

## Starting unit
The service is now enabled, and you can start it:
```
$ sudo systemctl start barman-cloud-exporter
```

## Status check and log
To check on its status:
```
$ sudo systemctl status barman-cloud-exporter
● barman-cloud-exporter.service - Barman Cloud Exporter for Prometheus
     Loaded: loaded (/etc/systemd/system/barman-cloud-exporter.service; enabled; >
     Active: active (running) since Tue 2024-03-26 18:16:02 UTC; 7s ago
       Docs: https://github.com/tk512/barman_cloud_exporter
   Main PID: 3681980 (barman_cloud_ex)
      Tasks: 4 (limit: 19009)
     Memory: 9.2M
        CPU: 16ms
     CGroup: /system.slice/barman-cloud-exporter.service
             └─3681980 /var/lib/postgresql/barman_cloud_exporter>
```

The exporter now listens on port 61092 on all interfaces by default.

## Metrics endpoint test
To test using curl:
```
$ curl http://<IP ADDRESS>:61092/metrics
```

This endpoint can now be used by Prometheus for gathering metrics. 