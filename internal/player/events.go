package player

type EventKind int

const (
	EventUnknown EventKind = iota
	EventTimePos
	EventDuration
	EventPause
	EventVolume
	EventMediaTitle
	EventEOF
	EventStartFile
	EventFileLoaded
	EventIdle
)

type Event struct {
	Kind   EventKind
	Float  float64
	Int    int64
	Bool   bool
	String string
}
