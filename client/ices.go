package client

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/pion/webrtc/v3"
)

func ReadIcesConfig(link string) (ices []webrtc.ICEServer, err error) {
	defer err2.Return(&err)

	if link == "" {
		return
	}

	var r io.Reader

	if link == "-" {
		r = os.Stdin
	} else if strings.HasPrefix(link, "http") {
		resp := try.To1(http.Get(link))
		defer resp.Body.Close()
		r = resp.Body
		return
	} else {
		f := try.To1(os.Open(link))
		defer f.Close()
		r = f
	}

	try.To(json.NewDecoder(r).Decode(&ices))

	return
}
