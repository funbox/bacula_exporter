package daemon

// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"net/http"

	"pkg.re/essentialkaos/ek.v12/log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// /////////////////////////////////////////////////////////////////////////////

type baculaMetrics struct {
	LatestJobFiles       *prometheus.Desc
	LatestJobBytes       *prometheus.Desc
	LatestJobSchedTime   *prometheus.Desc
	LatestJobStartTime   *prometheus.Desc
	LatestJobEndTime     *prometheus.Desc
	SummaryJobTotalFiles *prometheus.Desc
	SummaryJobTotalBytes *prometheus.Desc
}

// /////////////////////////////////////////////////////////////////////////////

// startHTTPServer start HTTP server
func startHTTPServer(ip, port, endpoint string) error {
	addr := ip + ":" + port

	log.Info("HTTP server is started on %s", addr)

	http.Handle(endpoint, promhttp.Handler())

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	return http.ListenAndServe(addr, nil)
}

// /////////////////////////////////////////////////////////////////////////////

// baculaCollector returns baculaMetrics struct for Prometheus
func baculaCollector() *baculaMetrics {
	return &baculaMetrics{
		LatestJobFiles: prometheus.NewDesc("bacula_latest_job_files_total",
			"Total files saved for server during latest backup for client combined",
			[]string{"name", "jobid", "level", "status"}, nil,
		),
		LatestJobBytes: prometheus.NewDesc("bacula_latest_job_bytes_total",
			"Total bytes saved for server during latest backup for client combined",
			[]string{"name", "jobid", "level", "status"}, nil,
		),
		LatestJobSchedTime: prometheus.NewDesc("bacula_latest_job_sched_time",
			"Timestamp when the latest job was scheduled",
			[]string{"name", "jobid", "level", "status"}, nil,
		),
		LatestJobStartTime: prometheus.NewDesc("bacula_latest_job_start_time",
			"Timestamp when the latest job was started",
			[]string{"name", "jobid", "level", "status"}, nil,
		),
		LatestJobEndTime: prometheus.NewDesc("bacula_latest_job_end_time",
			"Timestamp when the latest job was ended",
			[]string{"name", "jobid", "level", "status"}, nil,
		),
		SummaryJobTotalFiles: prometheus.NewDesc("bacula_summary_job_files_total",
			"Total files saved for server during all backups for client combined",
			[]string{"name", "level"}, nil,
		),
		SummaryJobTotalBytes: prometheus.NewDesc("bacula_summary_job_bytes_total",
			"Total bytes saved for server during all backups for client combined",
			[]string{"name", "level"}, nil,
		),
	}
}

// /////////////////////////////////////////////////////////////////////////////

// Describe implements Describe() method using by the Prometheus registry
// when describing metrics
func (collector *baculaMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.LatestJobFiles
	ch <- collector.LatestJobBytes
	ch <- collector.LatestJobSchedTime
	ch <- collector.LatestJobStartTime
	ch <- collector.LatestJobEndTime
	ch <- collector.SummaryJobTotalFiles
	ch <- collector.SummaryJobTotalBytes
}

// Collect implements Collect() method using by the Prometheus registry
// when collecting metrics
func (collector *baculaMetrics) Collect(ch chan<- prometheus.Metric) {
	latestJobs, err := env.DB.GetLatestJobs()

	if err != nil {
		log.Crit(err.Error())
		return
	}

	for _, job := range latestJobs {
		ch <- prometheus.MustNewConstMetric(
			collector.LatestJobFiles,
			prometheus.GaugeValue,
			float64(job.JobFiles),
			job.Name,
			job.JobId,
			job.Level,
			job.Status,
		)
		ch <- prometheus.MustNewConstMetric(
			collector.LatestJobBytes,
			prometheus.GaugeValue,
			float64(job.JobBytes),
			job.Name,
			job.JobId,
			job.Level,
			job.Status,
		)
		ch <- prometheus.MustNewConstMetric(
			collector.LatestJobSchedTime,
			prometheus.CounterValue,
			float64(job.SchedTime),
			job.Name,
			job.JobId,
			job.Level,
			job.Status,
		)
		ch <- prometheus.MustNewConstMetric(
			collector.LatestJobStartTime,
			prometheus.CounterValue,
			float64(job.StartTime),
			job.Name,
			job.JobId,
			job.Level,
			job.Status,
		)
		ch <- prometheus.MustNewConstMetric(
			collector.LatestJobEndTime,
			prometheus.CounterValue,
			float64(job.EndTime),
			job.Name,
			job.JobId,
			job.Level,
			job.Status,
		)
	}

	jobsSummary, err := env.DB.GetJobsSummary()

	if err != nil {
		log.Crit(err.Error())
		return
	}

	for _, job := range jobsSummary {
		ch <- prometheus.MustNewConstMetric(
			collector.SummaryJobTotalFiles,
			prometheus.GaugeValue,
			float64(job.TotalJobFiles),
			job.Name,
			job.Level,
		)
		ch <- prometheus.MustNewConstMetric(
			collector.SummaryJobTotalBytes,
			prometheus.GaugeValue,
			float64(job.TotalJobBytes),
			job.Name,
			job.Level,
		)
	}
}

// /////////////////////////////////////////////////////////////////////////////
