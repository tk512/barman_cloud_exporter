package collector

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	backup = "backup"
)

// Metric descriptors.
var (
	backupSize = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, backup, "size_bytes"),
		"Combined size of all registered binlog files.",
		[]string{}, nil,
	)
)

type BarmanCloudBackup struct{}

func (BarmanCloudBackup) Name() string {
	return backup
}

// Help describes the role of the Scraper.
func (BarmanCloudBackup) Help() string {
	return "Collect from Barman Cloud base backup"
}

func (BarmanCloudBackup) Scrape(ch chan<- prometheus.Metric, logger log.Logger) error {
	fmt.Println("TODO get Backup data %v", ch)
	ch <- prometheus.MustNewConstMetric(backupSize, prometheus.GaugeValue, float64(2123355))
	return nil
}

// check interface
var _ Scraper = BarmanCloudBackup{}
