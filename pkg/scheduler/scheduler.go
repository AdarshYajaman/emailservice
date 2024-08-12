package scheduler

import (
	"citi.com/179563_genesis_mail/pkg/service"

	"github.com/robfig/cron/v3"
)

// StartCRONScheduler which runs nightly at midnight to load new configurations from jobs table, and in turn removes existing and creates secondary jobs
func StartCRONScheduler(jobRefreshSchedule string) {

	primaryCron := cron.New()
	sch := scheduler{}
	primaryCron.AddJob(jobRefreshSchedule, sch)
	primaryCron.Start()

	//called first time only as part of application startup
	sch.Run()
}

type scheduler struct {
}

// Run creates a secondary cron based on expressions defined in jobs table
func (s scheduler) Run() {
	service.StartSecondaryCron()
}
