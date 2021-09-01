package moniker

type Span string

const (
	Query    Span = "Query"
	Ping          = "Ping"
	Prepare       = "Prepare"
	Exec          = "Exec"
	Begin         = "Begin"
	Reset         = "Reset"
	Close         = "Close"
	Commit        = "Commit"
	Rollback      = "Rollback"
)

func (n Span) String() string { return string(n) }
