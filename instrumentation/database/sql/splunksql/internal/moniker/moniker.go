package moniker

type Span string

const (
	Query    Span = "Query"
	Ping     Span = "Ping"
	Prepare  Span = "Prepare"
	Exec     Span = "Exec"
	Begin    Span = "Begin"
	Reset    Span = "Reset"
	Close    Span = "Close"
	Commit   Span = "Commit"
	Rollback Span = "Rollback"
	Rows     Span = "Rows"
)

func (n Span) String() string { return string(n) }

type Event string

const (
	Next Event = "Next"
)

func (n Event) String() string { return string(n) }
