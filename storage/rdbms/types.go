package rdbms

// ////////////////////////////////////////////////////////////////////////////////// //

type BaculaJob struct {
	Name      string `db:"name"`
	Level     string `db:"level"`
	JobId     uint32 `db:"jobid"`
	Status    string `db:"jobstatus"`
	SchedTime uint32 `db:"schedtime"`
	StartTime uint32 `db:"starttime"`
	EndTime   uint32 `db:"endtime"`
	JobBytes  uint64 `db:"jobbytes"`
	JobFiles  uint64 `db:"jobfiles"`
}

type BaculaJobSummary struct {
	Name          string `db:"name"`
	Level         string `db:"level"`
	TotalJobBytes uint64 `db:"totaljobbytes"`
	TotalJobFiles uint64 `db:"totaljobfiles"`
}

type BaculaSummary struct {
	ScheduledJobs uint32 `db:"scheduledjobs"`
}

// ////////////////////////////////////////////////////////////////////////////////// //
