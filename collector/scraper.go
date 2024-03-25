package collector

import (
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

type Scraper interface {
	Name() string
	Help() string
	Scrape(ch chan<- prometheus.Metric, logger log.Logger) error
}
