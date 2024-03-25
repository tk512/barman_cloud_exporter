package collector

import (
	"errors"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"strings"
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

func (e *Exporter) parseAndNewMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, valueType prometheus.ValueType, stats map[string]string, key string, labelValues ...string) error {
	return e.extractValueAndNewMetric(ch, desc, valueType, parse, stats, key, labelValues...)
}

func (e *Exporter) parseBoolAndNewMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, valueType prometheus.ValueType, stats map[string]string, key string, labelValues ...string) error {
	return e.extractValueAndNewMetric(ch, desc, valueType, parseBool, stats, key, labelValues...)
}

func (e *Exporter) parseTimevalAndNewMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, valueType prometheus.ValueType, stats map[string]string, key string, labelValues ...string) error {
	return e.extractValueAndNewMetric(ch, desc, valueType, parseTimeval, stats, key, labelValues...)
}

func (e *Exporter) extractValueAndNewMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, valueType prometheus.ValueType, f func(map[string]string, string, log.Logger) (float64, error), stats map[string]string, key string, labelValues ...string) error {
	v, err := f(stats, key, e.logger)
	if err == errKeyNotFound {
		return nil
	}
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(desc, valueType, v, labelValues...)
	return nil
}

func parse(stats map[string]string, key string, logger log.Logger) (float64, error) {
	value, ok := stats[key]
	if !ok {
		level.Debug(logger).Log("msg", "Key not found", "key", key)
		return 0, errKeyNotFound
	}

	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		level.Error(logger).Log("msg", "Failed to parse", "key", key, "value", value, "err", err)
		return 0, err
	}
	return v, nil
}

func parseBool(stats map[string]string, key string, logger log.Logger) (float64, error) {
	value, ok := stats[key]
	if !ok {
		level.Debug(logger).Log("msg", "Key not found", "key", key)
		return 0, errKeyNotFound
	}

	switch value {
	case "yes":
		return 1, nil
	case "no":
		return 0, nil
	default:
		level.Error(logger).Log("msg", "Failed to parse", "key", key, "value", value)
		return 0, errors.New("failed parse a bool value")
	}
}

func parseTimeval(stats map[string]string, key string, logger log.Logger) (float64, error) {
	value, ok := stats[key]
	if !ok {
		level.Debug(logger).Log("msg", "Key not found", "key", key)
		return 0, errKeyNotFound
	}
	values := strings.Split(value, ".")

	if len(values) != 2 {
		level.Error(logger).Log("msg", "Failed to parse", "key", key, "value", value)
		return 0, errors.New("failed parse a timeval value")
	}

	seconds, err := strconv.ParseFloat(values[0], 64)
	if err != nil {
		level.Error(logger).Log("msg", "Failed to parse", "key", key, "value", value, "err", err)
		return 0, errors.New("failed parse a timeval value")
	}

	microseconds, err := strconv.ParseFloat(values[1], 64)
	if err != nil {
		level.Error(logger).Log("msg", "Failed to parse", "key", key, "value", value, "err", err)
		return 0, errors.New("failed parse a timeval value")
	}

	return (seconds + microseconds/(1000.0*1000.0)), nil
}

func sum(stats map[string]string, keys ...string) (float64, error) {
	s := 0.
	for _, key := range keys {
		if _, ok := stats[key]; !ok {
			return 0, errKeyNotFound
		}
		v, err := strconv.ParseFloat(stats[key], 64)
		if err != nil {
			return 0, err
		}
		s += v
	}
	return s, nil
}

func firstError(errors ...error) error {
	for _, v := range errors {
		if v != nil {
			return v
		}
	}
	return nil
}
