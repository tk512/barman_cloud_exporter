package collector

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	backup = "backup"
)

// Metric descriptors
var (
	backupLatestBytes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, backup, "latest_bytes"),
		"Latest backup size in bytes",
		[]string{"bucket_name"}, nil,
	)
	backupLatestProcessedTimestamp = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, backup, "latest_timestamp_seconds"),
		"Latest backup performed at timestamp",
		[]string{"bucket_name"}, nil,
	)
	backupLatestProcessedDuration = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, backup, "latest_processed_duration"),
		"Latest backup performed duration in seconds",
		[]string{"bucket_name"}, nil,
	)
)

type BarmanCloudBackup struct{}

// Name of the Scraper
func (BarmanCloudBackup) Name() string {
	return "barman_cloud_backup"
}

// Help describes the role of the Scraper.
func (BarmanCloudBackup) Help() string {
	return "Collect from Barman Cloud backup result"
}

// Scrape collects data from result files and sends it over channel as prometheus metric.
func (BarmanCloudBackup) Scrape(ch chan<- prometheus.Metric, logger log.Logger) error {
	fmt.Println("TODO get Backup data %v", ch)
	bucketName := "helium-db-2a" // Sample
	ch <- prometheus.MustNewConstMetric(backupLatestBytes, prometheus.GaugeValue, float64(1), bucketName)
	ch <- prometheus.MustNewConstMetric(backupLatestProcessedTimestamp, prometheus.GaugeValue, float64(600), bucketName)
	ch <- prometheus.MustNewConstMetric(backupLatestProcessedDuration, prometheus.GaugeValue, float64(6), bucketName)
	return nil
}

var _ Scraper = BarmanCloudBackup{}
