package events

type Event interface {
	isEvent()
}

type ConnectEvent struct {
	URL string
}

func (ConnectEvent) isEvent() {}

type MessageEvent struct {
	Msg string
}

func (MessageEvent) isEvent() {}

type LogEvent struct {
	Content string
}

func (LogEvent) isEvent() {}

type IncomingEvent struct {
	Content string
}

func (IncomingEvent) isEvent() {}

type ConnectedEvent struct{}

func (ConnectedEvent) isEvent() {}

// HandshakeDone signals success with res for log (dedicated, no string parse).
type HandshakeDone struct {
	Res string
}

func (HandshakeDone) isEvent() {}
