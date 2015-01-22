package manager

import "encoding/gob"

const (
	MESSAGE_TYPE_CONFIG = 0
)

func init() {
	gob.Register(ManagerMessage{})
}

// ManagerMessage is a message that is sent between managers to convey changes in configuration
type ManagerMessage struct {
	Type    int
	Payload interface{}
}
