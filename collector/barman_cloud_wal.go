package collector

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	wal = "wal"
)

// Metric descriptors.
var (
	walDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, wal, "test"),
		"WAL test.",
		[]string{"command"}, nil,
	)
)

type BarmanCloudWal struct{}

func (BarmanCloudWal) Name() string {
	return wal
}

func (BarmanCloudWal) Help() string {
	return "Collect from Barman Cloud WAL archive data"
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (BarmanCloudWal) Scrape(ch chan<- prometheus.Metric, logger log.Logger) error {
	fmt.Println("TODO get WAL data %v", ch)
	ch <- prometheus.MustNewConstMetric(walDesc, prometheus.GaugeValue, float64(111), "blaa")
	return nil
}

// check interface
var _ Scraper = BarmanCloudWal{}
