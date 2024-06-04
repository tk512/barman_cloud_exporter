package collector

import (
	"bytes"
	"errors"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
)

const (
	backup = "backup"
)

// Metric descriptors
var (
	backupLatestBytes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, backup, "latest_bytes"),
		"Latest backup size in bytes",
		[]string{"bucket_name", "backup_id"}, nil,
	)
	backupLatestProcessedTimestamp = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, backup, "latest_timestamp_seconds"),
		"Latest backup performed at timestamp",
		[]string{"bucket_name", "backup_id"}, nil,
	)
	backupLatestProcessedDuration = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, backup, "latest_processed_duration"),
		"Latest backup performed duration in seconds",
		[]string{"bucket_name", "backup_id"}, nil,
	)
	backupSuccess = prometheus.NewDesc(prometheus.BuildFQName(namespace, backup, "lastest_success"),
		"Lasted backup status succesful (1) or not (0)",
		[]string{"bucket_name", "backup_id"}, nil)
)

type cloudBackupRecord struct {
	timestamp      int64
	bucketName     string
	success        int
	backupDuration int
	backupSize     int
	backupId       string
}

type BarmanCloudBackup struct {
	BackupLogFile string
}

// Name of the Scraper
func (b *BarmanCloudBackup) Name() string {
	return "barman_cloud_backup"
}

// Help describes the role of the Scraper.
func (b *BarmanCloudBackup) Help() string {
	return "Collect from Barman Cloud backup result"
}

// Scrape collects data from result files and sends it over channel as prometheus metric.
func (b *BarmanCloudBackup) Scrape(ch chan<- prometheus.Metric, logger log.Logger) error {
	var buf []byte

	buf, _ = ReadTailOfFile(b.BackupLogFile, tailBufSize)
	reader := NewTsvReader(bytes.NewBuffer(buf))

	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	// Log invalid TSV line as a warning
	logInvalidTsv := func(lines []string) {
		_ = level.Warn(logger).Log("msg", "Could not parse input backup TSV on line", "line", lines)
	}

	var cloudBackupRecords []cloudBackupRecord
	for _, record := range records {
		if len(record) <= 1 {
			break
		}

		timestamp, err := strconv.Atoi(record[0])
		if err != nil {
			logInvalidTsv(record)
			continue
		}

		exit_code, err := strconv.Atoi(record[2])
		if err != nil {
			logInvalidTsv(record)
			continue
		}

		success := 0
		if exit_code == 0 {
			success = 1
		}

		backupDuration, err := strconv.Atoi(record[3])
		if err != nil {
			logInvalidTsv(record)
			continue
		}

		backupSize, err := strconv.Atoi(record[4])
		if err != nil {
			logInvalidTsv(record)
			continue
		}

		cloudBackupRecords = append(cloudBackupRecords, cloudBackupRecord{
			timestamp:      int64(timestamp),
			bucketName:     record[1],
			success:        success,
			backupDuration: backupDuration,
			backupSize:     backupSize,
			backupId:       record[5],
		})

	}

	if len(cloudBackupRecords) == 0 {
		return errors.New("No values in backup TSV file: " + b.BackupLogFile)
	}

	// Take last record in slice for the latest metrics
	latestBackupRecord := cloudBackupRecords[len(cloudBackupRecords)-1]
	latestBackupId := latestBackupRecord.backupId
	bucketName := latestBackupRecord.bucketName

	// Backup size
	ch <- prometheus.MustNewConstMetric(backupLatestBytes, prometheus.GaugeValue,
		float64(latestBackupRecord.backupSize), bucketName, latestBackupId)

	// Latest processed timestamp
	ch <- prometheus.MustNewConstMetric(backupLatestProcessedTimestamp, prometheus.GaugeValue,
		float64(latestBackupRecord.timestamp), bucketName, latestBackupId)

	// Latest processed backup - duration in seconds
	ch <- prometheus.MustNewConstMetric(backupLatestProcessedDuration, prometheus.GaugeValue,
		float64(latestBackupRecord.backupDuration), bucketName, latestBackupId)

	// Lastest process backup successful (1) or not (0)
	ch <- prometheus.MustNewConstMetric(backupSuccess, prometheus.GaugeValue,
		float64(latestBackupRecord.success), bucketName, latestBackupId)

	return nil
}
