package gitexporter

import "github.com/harness/harness-migrate/internal/report"

const (
	ReportTypeWebhooks    = "webhook"
	ReportTypePRs         = "pull requests"
	ReportTypeBranchRules = "branch rules"
	ReportTypeUsers       = "users"
)

func publishReport(report map[string]*report.Report) {
	for _, v := range report {
		v.PublishReport()
	}
}
