package collector

import (
	"errors"
	"fmt"
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
	address  string
	logger   log.Logger
	scrapers []Scraper

	up                    *prometheus.Desc
	uptime                *prometheus.Desc
	time                  *prometheus.Desc
	version               *prometheus.Desc
	rusageUser            *prometheus.Desc
	rusageSystem          *prometheus.Desc
	bytesRead             *prometheus.Desc
	bytesWritten          *prometheus.Desc
	currentConnections    *prometheus.Desc
	maxConnections        *prometheus.Desc
	connectionsTotal      *prometheus.Desc
	rejectedConnections   *prometheus.Desc
	connsYieldedTotal     *prometheus.Desc
	listenerDisabledTotal *prometheus.Desc
	currentBytes          *prometheus.Desc
	limitBytes            *prometheus.Desc
	commands              *prometheus.Desc
	items                 *prometheus.Desc
	itemsTotal            *prometheus.Desc
	evictions             *prometheus.Desc
	reclaimed             *prometheus.Desc
	itemStoreTooLarge     *prometheus.Desc
	itemStoreNoMemory     *prometheus.Desc
}

// New returns an initialized exporter.
func New(scrapers []Scraper, logger log.Logger) *Exporter {
	return &Exporter{
		logger:   logger,
		scrapers: scrapers,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could the memcached server be reached.",
			nil,
			nil,
		),
		uptime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "uptime_seconds"),
			"Number of seconds since the server started.",
			nil,
			nil,
		),
		time: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "time_seconds"),
			"current UNIX time according to the server.",
			nil,
			nil,
		),
		version: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "version"),
			"The version of this memcached server.",
			[]string{"version"},
			nil,
		),
		rusageUser: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_user_cpu_seconds_total"),
			"Accumulated user time for this process.",
			nil,
			nil,
		),
		rusageSystem: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "process_system_cpu_seconds_total"),
			"Accumulated system time for this process.",
			nil,
			nil,
		),
		bytesRead: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "read_bytes_total"),
			"Total number of bytes read by this server from network.",
			nil,
			nil,
		),
		bytesWritten: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "written_bytes_total"),
			"Total number of bytes sent by this server to network.",
			[]string{"foo", "bar"},
			nil,
		),
		currentConnections: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "current_connections"),
			"Current number of open connections.",
			nil,
			nil,
		),
		maxConnections: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "max_connections"),
			"Maximum number of clients allowed.",
			nil,
			nil,
		),
		connectionsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "connections_total"),
			"Total number of connections opened since the server started running.",
			nil,
			nil,
		),
		rejectedConnections: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "connections_rejected_total"),
			"Total number of connections rejected due to hitting the memcached's -c limit in maxconns_fast mode.",
			nil,
			nil,
		),
		connsYieldedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "connections_yielded_total"),
			"Total number of connections yielded running due to hitting the memcached's -R limit.",
			nil,
			nil,
		),
		listenerDisabledTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "connections_listener_disabled_total"),
			"Number of times that memcached has hit its connections limit and disabled its listener.",
			nil,
			nil,
		),
		currentBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "current_bytes"),
			"Current number of bytes used to store items.",
			nil,
			nil,
		),
		limitBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "limit_bytes"),
			"Number of bytes this server is allowed to use for storage.",
			nil,
			nil,
		),
		commands: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "commands_total"),
			"Total number of all requests broken down by command (get, set, etc.) and status.",
			[]string{"command", "status"},
			nil,
		),
		items: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "current_items"),
			"Current number of items stored by this instance.",
			nil,
			nil,
		),
		itemsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "items_total"),
			"Total number of items stored during the life of this instance.",
			nil,
			nil,
		),
		evictions: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "items_evicted_total"),
			"Total number of valid items removed from cache to free memory for new items.",
			nil,
			nil,
		),
		reclaimed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "items_reclaimed_total"),
			"Total number of times an entry was stored using memory from an expired entry.",
			nil,
			nil,
		),
		itemStoreTooLarge: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "item_too_large_total"),
			"The number of times an item exceeded the max-item-size when being stored.",
			nil,
			nil,
		),
		itemStoreNoMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "item_no_memory_total"),
			"The number of times an item could not be stored due to no more memory.",
			nil,
			nil,
		),
	}
}

// Describe describes all the metrics exported by the memcached exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	ch <- e.uptime
	ch <- e.time
	ch <- e.version
	ch <- e.rusageUser
	ch <- e.rusageSystem
	ch <- e.bytesRead
	ch <- e.bytesWritten
	ch <- e.currentConnections
	ch <- e.maxConnections
	ch <- e.connectionsTotal
	ch <- e.rejectedConnections
	ch <- e.connsYieldedTotal
	ch <- e.listenerDisabledTotal
	ch <- e.currentBytes
	ch <- e.limitBytes
	ch <- e.commands
	ch <- e.items
	ch <- e.itemsTotal
	ch <- e.evictions
	ch <- e.reclaimed
	ch <- e.itemStoreTooLarge
	ch <- e.itemStoreNoMemory
}

// Collect fetches the statistics from the configured memcached server, and
// delivers them as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	var up = 0.0
	//scrapeTime := time.Now()

	var wg sync.WaitGroup
	defer wg.Wait()

	for _, scraper := range e.scrapers {
		wg.Add(1)
		go func(scraper Scraper) {
			defer wg.Done()
			label := "collect." + scraper.Name()
			scrapeTime := time.Now()
			collectorSuccess := 1.0
			if err := scraper.Scrape(ch, log.With(e.logger, "scraper", scraper.Name())); err != nil {
				_ = level.Error(e.logger).Log("msg", "Error from scraper", "scraper", scraper.Name(), "err", err)
				collectorSuccess = 0.0
			}
			fmt.Println(label)
			fmt.Println(collectorSuccess)
			fmt.Println(scrapeTime)
			//ch <- prometheus.MustNewConstMetric(..ScrapeCollectorSuccess, prometheus.GaugeValue, collectorSuccess, label)
			//ch <- prometheus.MustNewConstMetric(...ScrapeDurationSeconds, prometheus.GaugeValue, time.Since(scrapeTime).Seconds(), label)
		}(scraper)
	}

	// TODO check if any errors happened
	up = 1.0

	//
	//ch <- prometheus.MustNewConstMetric(e.bytesWritten, prometheus.GaugeValue, float64(10001), "a", "b")
	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, up)

	//ch <- prometheus.MustNewConstMetric(...ScrapeDurationSeconds, prometheus.GaugeValue,
	//	time.Since(scrapeTime).Seconds(), "TBD")

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
