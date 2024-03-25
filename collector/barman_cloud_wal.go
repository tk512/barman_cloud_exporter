package collector

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	wal = "wal"
)

// Metric descriptors
var (
	walLatestBytes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, wal, "latest_bytes"),
		"Latest WAL size in bytes",
		[]string{"bucket_name"}, nil,
	)
	walLatestProcessedTimestamp = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, wal, "latest_timestamp_seconds"),
		"Latest WAL processed at timestamp",
		[]string{"bucket_name"}, nil,
	)
	walLatestProcessedDuration = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, wal, "latest_processed_duration"),
		"Latest WAL process duration in seconds",
		[]string{"bucket_name"}, nil,
	)
	walFailedLastHourTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, wal, "failed_last_hour_total"),
		"Number of WAL archives that failed in the last hour",
		[]string{"bucket_name"}, nil,
	)
)

type BarmanCloudWal struct{}

// Name of the Scraper
func (BarmanCloudWal) Name() string {
	return "barman_cloud_wal"
}

// Help describes the role of the Scraper.
func (BarmanCloudWal) Help() string {
	return "Collect from Barman Cloud WAL archive result"
}

// Scrape collects data from result files and sends it over channel as prometheus metric.
func (BarmanCloudWal) Scrape(ch chan<- prometheus.Metric, logger log.Logger) error {
	fmt.Println("TODO get WAL data %v", ch)

	// Iterate log file
	// 1711371873      helium-db-2a    000000010000369F00000009        16777216        0       7

	bucketName := "helium-db-2a" // Sample
	ch <- prometheus.MustNewConstMetric(walLatestBytes, prometheus.GaugeValue, float64(100), bucketName)
	ch <- prometheus.MustNewConstMetric(walLatestProcessedTimestamp, prometheus.GaugeValue, float64(200), bucketName)
	ch <- prometheus.MustNewConstMetric(walLatestProcessedDuration, prometheus.GaugeValue, float64(2), bucketName)
	ch <- prometheus.MustNewConstMetric(walFailedLastHourTotal, prometheus.GaugeValue, float64(0), bucketName)
	return nil
}

var _ Scraper = BarmanCloudWal{}
