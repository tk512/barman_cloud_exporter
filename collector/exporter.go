package collector

import (
	"errors"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"time"
)

const (
	namespace = "barman_cloud"
)

var errKeyNotFound = errors.New("key not found")

// Exporter collects metrics from barman cloud output
type Exporter struct {
	logger   log.Logger
	scrapers []Scraper

	info                  *prometheus.Desc
	up                    *prometheus.Desc
	scrapeDurationSeconds *prometheus.Desc
}

// New returns an initialized exporter.
func New(scrapers []Scraper, logger log.Logger) *Exporter {
	return &Exporter{
		logger:   logger,
		scrapers: scrapers,
		info: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "info"),
			"Cloud backup info",
			[]string{"cloud_provider"},
			nil,
		),
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Indicates whether the cloud backup successful",
			nil,
			nil,
		),
		scrapeDurationSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "scrape_duration_seconds"),
			"Scrape duration in seconds",
			nil,
			nil,
		),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.info
	ch <- e.up
	ch <- e.scrapeDurationSeconds
}

// Collect fetches the statistics from the scrapers, and
// delivers them as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	var up = 1.0
	scrapeTime := time.Now()

	var wg sync.WaitGroup
	defer wg.Wait()

	for _, scraper := range e.scrapers {
		wg.Add(1)
		go func(scraper Scraper) {
			defer wg.Done()
			if err := scraper.Scrape(ch, log.With(e.logger, "scraper", scraper.Name())); err != nil {
				_ = level.Error(e.logger).Log("msg", "Error from scraper",
					"scraper", scraper.Name(), "err", err)
				up = 0
			}
		}(scraper)
	}

	ch <- prometheus.MustNewConstMetric(e.info, prometheus.GaugeValue, 1.0, "cloud provider placeholder")
	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, up)
	ch <- prometheus.MustNewConstMetric(e.scrapeDurationSeconds, prometheus.GaugeValue,
		time.Since(scrapeTime).Seconds())
}
