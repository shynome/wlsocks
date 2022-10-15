package lens

import (
	"github.com/pion/webrtc/v3"
	"github.com/shynome/wl/ortc"
)

type ConnectInfo struct {
	Offer ortc.Signal
	Ices  []webrtc.ICEServer
}

const LinkTopic = "link"
