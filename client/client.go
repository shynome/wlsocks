package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/pion/webrtc/v3"
	"github.com/shynome/wl"
	"github.com/shynome/wl/ortc"
	"github.com/shynome/wlsocks/lens"
)

type Client struct {
	config    Config
	lens      *lens.Lens
	transport *wl.Transport
	ctx       context.Context
	id        string
}

type Config struct {
	Lens string
	User string
	Ices string
}

func New(config Config) (client *Client) {
	t := wl.NewTransport()
	lens := lens.New(config.Lens, config.User, "")
	client = &Client{
		id:        "client",
		lens:      lens,
		transport: t,
		ctx:       context.Background(),
	}
	return
}

func (c *Client) Connect() (err error) {
	defer err2.Return(&err)

	config := c.config

	settingEngine := webrtc.SettingEngine{}
	settingEngine.DetachDataChannels()
	api := webrtc.NewAPI(webrtc.WithSettingEngine(settingEngine))

	ices := try.To1(ReadIcesConfig(config.Ices))
	wconfig := webrtc.Configuration{
		ICEServers: ices,
	}
	pc := try.To1(api.NewPeerConnection(wconfig))

	offer := try.To1(ortc.CreateOffer(pc))

	var info lens.ConnectInfo = lens.ConnectInfo{
		Offer: offer,
		Ices:  ices,
	}
	infoBytes := try.To1(json.Marshal(info))
	rofferBytes := try.To1(c.lens.Call(lens.LinkTopic, infoBytes))
	var roffer ortc.Signal
	try.To(json.Unmarshal(rofferBytes, &roffer))
	try.To(ortc.Handshake(pc, roffer))

	session := try.To1(wl.NewClientSession(pc))
	c.transport.Set(c.id, session)

	var cancel context.CancelFunc
	c.ctx, cancel = context.WithCancel(context.Background())
	pc.OnConnectionStateChange(func(pcs webrtc.PeerConnectionState) {
		if pcs == webrtc.PeerConnectionStateDisconnected {
			cancel()
			c.transport.Set(c.id, nil)
			pc.Close()
		}
	})
	return
}

func (c *Client) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *Client) NewConn() (conn net.Conn, err error) {
	return c.transport.NewConn("client")
}

func (c *Client) Close() (err error) {
	return
}

func (c *Client) Serve(l net.Listener) (err error) {
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("accept err:", err)
			return err
		}
		go func(conn net.Conn) {
			var err error
			defer func() {
				if err != nil {
					fmt.Println("conn err:", err)
				}
			}()
			defer err2.Return(&err)
			defer conn.Close()

			wconn := try.To1(c.NewConn())
			defer wconn.Close()

			go io.Copy(wconn, conn)
			io.Copy(conn, wconn)
		}(conn)
	}
}
