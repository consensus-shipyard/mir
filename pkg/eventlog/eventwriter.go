package eventlog

type EventWriter interface {
	Write(record EventRecord) error
	Close() error
}
