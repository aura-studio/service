package message

import (
	"github.com/aura-studio/encodingx"
	"github.com/aura-studio/service/route"
)

type Message struct {
	ID       uint64
	Route    route.Route
	Encoding encodingx.Encoding
	Data     []byte
}
