package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/pion/webrtc/v3"
	"github.com/shynome/wl"
	"github.com/shynome/wl/ortc"
)

func runClient() {
	lens := Lens{
		Endpoint: args.lens,
		User:     args.user,
	}

	t := wl.NewTransport()

	settingEngine := webrtc.SettingEngine{}
	settingEngine.DetachDataChannels()
	api := webrtc.NewAPI(webrtc.WithSettingEngine(settingEngine))

	ices := try.To1(readConfig(args.ices))
	config := webrtc.Configuration{
		ICEServers: ices,
	}
	pc := try.To1(api.NewPeerConnection(config))

	offer := try.To1(ortc.CreateOffer(pc))

	var info ConnectInfo = ConnectInfo{
		Offer: offer,
		Ices:  ices,
	}
	infoBytes := try.To1(json.Marshal(info))
	rofferBytes := try.To1(lens.Call(LinkTopic, infoBytes))
	var roffer ortc.Signal
	try.To(json.Unmarshal(rofferBytes, &roffer))
	try.To(ortc.Handshake(pc, roffer))

	fmt.Println("server connected")

	const id = "client"
	session := try.To1(wl.NewClientSession(pc))
	t.Set(id, session)

	pc.OnConnectionStateChange(func(pcs webrtc.PeerConnectionState) {
		if pcs == webrtc.PeerConnectionStateDisconnected {
			os.Exit(1)
		}
	})

	addr := &net.TCPAddr{IP: net.ParseIP(args.addr), Port: args.port}
	l := try.To1(net.ListenTCP("tcp", addr))

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("accept err:", err)
			break
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

			wconn := try.To1(t.NewConn(id))
			defer wconn.Close()

			go io.Copy(wconn, conn)
			io.Copy(conn, wconn)
		}(conn)
	}
}

func readConfig(link string) (ices []webrtc.ICEServer, err error) {
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
