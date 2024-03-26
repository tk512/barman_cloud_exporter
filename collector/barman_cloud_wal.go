package collector

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
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

type cloudWalRecord struct {
	timestamp   int64
	bucketName  string
	walName     string
	walSize     int
	success     int
	walDuration int
}

type BarmanCloudWal struct {
	WalLogFile string
}

// Name of the Scraper
func (w *BarmanCloudWal) Name() string {
	return "barman_cloud_wal"
}

// Help describes the role of the Scraper.
func (w *BarmanCloudWal) Help() string {
	return "Collect from Barman Cloud WAL archive result"
}

// Scrape collects data from result files and sends it over channel as prometheus metric.
func (w *BarmanCloudWal) Scrape(ch chan<- prometheus.Metric, logger log.Logger) error {
	var buf []byte

	buf, _ = ReadTailOfFile(w.WalLogFile, tailBufSize)
	reader := NewTsvReader(bytes.NewBuffer(buf))

	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	// Log invalid TSV line as a warning
	logInvalidTsv := func(lines []string) {
		_ = level.Warn(logger).Log("msg", "Could not parse input wal TSV on line", lines)
	}

	var cloudWalRecords []cloudWalRecord
	for _, record := range records {
		if len(record) <= 1 {
			break
		}

		timestamp, err := strconv.Atoi(record[0])
		if err != nil {
			logInvalidTsv(record)
			continue
		}

		walSize, err := strconv.Atoi(record[3])
		if err != nil {
			logInvalidTsv(record)
			continue
		}

		success, err := strconv.Atoi(record[4])
		if err != nil {
			logInvalidTsv(record)
			continue
		}

		walDuration, err := strconv.Atoi(record[5])
		if err != nil {
			logInvalidTsv(record)
			continue
		}

		cloudWalRecords = append(cloudWalRecords, cloudWalRecord{
			timestamp:   int64(timestamp),
			bucketName:  record[1],
			walName:     record[2],
			walSize:     walSize,
			success:     success,
			walDuration: walDuration,
		})

	}

	if len(cloudWalRecords) == 0 {
		return errors.New("No values in wal TSV file: " + w.WalLogFile)
	}

	// Find time closest to an hour ago and iterate wals, to see if any failed
	failuresLastHr := 0
	now := time.Now()
	then := now.Add(time.Hour)

	for _, r := range cloudWalRecords {
		failuresLastHr = 1
		if r.timestamp >= then.Unix() {
			fmt.Println(r)
		}
	}
	// TODO

	// Take last record in slice for the latest metrics
	latestWalRecord := cloudWalRecords[len(cloudWalRecords)-1]
	bucketName := latestWalRecord.bucketName

	ch <- prometheus.MustNewConstMetric(walLatestBytes, prometheus.GaugeValue,
		float64(latestWalRecord.walSize), bucketName)
	ch <- prometheus.MustNewConstMetric(walLatestProcessedTimestamp,
		prometheus.GaugeValue, float64(latestWalRecord.timestamp), bucketName)
	ch <- prometheus.MustNewConstMetric(walLatestProcessedDuration,
		prometheus.GaugeValue, float64(latestWalRecord.walDuration), bucketName)
	ch <- prometheus.MustNewConstMetric(walFailedLastHourTotal,
		prometheus.GaugeValue, float64(failuresLastHr), bucketName)

	return nil
}
