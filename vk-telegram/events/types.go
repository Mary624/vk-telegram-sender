package events

type Processor interface {
	Process(e Event) error
}

type Fetcher interface {
	Fetch(limit int) ([]Event, error)
}

type Event struct {
	Text   string
	ChatID int
	Meta   interface{}
}
