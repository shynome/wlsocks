package lens

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type Lens struct {
	Endpoint string
	User     string
	Pass     string

	HTTPClient *http.Client
}

func New(Endpoint string, User string, Pass string) *Lens {
	hclient := &http.Client{
		Timeout: 5 * time.Second,
	}
	return &Lens{
		Endpoint: Endpoint,
		User:     User,
		Pass:     Pass,

		HTTPClient: hclient,
	}
}

func (l *Lens) Dial(topic string, id string, input []byte) (err error) {
	defer err2.Return(&err)
	link := try.To1(WithTopic(l.Endpoint, topic))
	req := try.To1(http.NewRequest(http.MethodDelete, link, bytes.NewReader(input)))
	req.SetBasicAuth(l.User, l.Pass)
	req.Header.Set("X-Event-Id", id)
	resp := try.To1(l.HTTPClient.Do(req))
	try.To(CheckResp(resp))
	return
}

func (l *Lens) Call(topic string, input []byte) (output []byte, err error) {
	defer err2.Return(&err)
	endpoint := try.To1(WithTopic(l.Endpoint, topic))
	req := try.To1(http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(input)))
	req.SetBasicAuth(l.User, "")
	resp := try.To1(l.HTTPClient.Do(req))
	try.To(CheckResp(resp))
	output = try.To1(io.ReadAll(resp.Body))
	return
}

func WithTopic(endpoint string, topic string) (rEndpoint string, err error) {
	defer err2.Return(&err)

	if topic == "" {
		return endpoint, nil
	}

	u := try.To1(url.Parse(endpoint))
	q := u.Query()
	q.Set("t", topic)
	u.RawQuery = q.Encode()

	rEndpoint = u.String()

	return
}

func CheckResp(resp *http.Response) (err error) {
	defer err2.Return(&err)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return
	}
	defer resp.Body.Close()
	errText := try.To1(io.ReadAll(resp.Body))
	err = fmt.Errorf("server err. code: %v. content: %s", resp.StatusCode, errText)
	return
}
