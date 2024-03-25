package main

import (
	"barman_cloud_exporter/collector"
	"context"
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	var (
		webConfig   = webflag.AddFlags(kingpin.CommandLine, ":61092")
		metricsPath = kingpin.Flag("web.telemetry-path",
			"Path under which to expose metrics.").Default("/metrics").String()
	)

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Version(version.Print("barman_cloud_exporter"))
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	_ = level.Info(logger).Log("msg", "Starting barman_cloud_exporter", "version", version.Info())
	_ = level.Info(logger).Log("msg", "Build context", "context", version.BuildContext())

	prometheus.MustRegister(versioncollector.NewCollector("barman_cloud_exporter"))

	var newScrapers = []collector.Scraper{
		collector.BarmanCloudBackup{},
		collector.BarmanCloudWal{},
	}
	handlerFunc := newHandler(newScrapers, logger)
	http.Handle(*metricsPath, promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer, handlerFunc))

	if *metricsPath != "/" && *metricsPath != "" {
		landingConfig := web.LandingConfig{
			Name:        "barman_cloud_exporter",
			Description: "Prometheus Exporter for Barman Cloud WAL archiving and backup",
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{
					Address: *metricsPath,
					Text:    "Metrics",
				},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			_ = level.Error(logger).Log("err", err)
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	srv := &http.Server{}
	if err := web.ListenAndServe(srv, webConfig, logger); err != nil {
		_ = level.Error(logger).Log("msg", "Error running HTTP server", "err", err)
		os.Exit(1)
	}
}

func newHandler(scrapers []collector.Scraper, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = level.Debug(logger).Log("msg", "scraping barman_cloud")

		// TODO increase scrape count

		ctx := r.Context()
		// If a timeout is configured via the Prometheus header, add it to the context.
		if v := r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"); v != "" {
			timeoutSeconds, err := strconv.ParseFloat(v, 64)
			if err != nil {
				_ = level.Error(logger).Log("msg", "Failed to parse timeout from header", "err", err)
			} else {
				// XXX TODO
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSeconds*float64(time.Second)))
				defer cancel()
				// Overwrite request with timeout context.
				r = r.WithContext(ctx)
			}
		}

		e := collector.New(scrapers, logger)
		registry := prometheus.NewRegistry()
		if err := registry.Register(e); err != nil {
			_ = level.Error(logger).Log("msg", "Failed to register collector", "err", err)
			return
		}

		promhttp.HandlerFor(
			registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError},
		).ServeHTTP(w, r)
	}
}
