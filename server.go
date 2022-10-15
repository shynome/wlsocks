package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/armon/go-socks5"
	"github.com/donovanhide/eventsource"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/pion/webrtc/v3"
	"github.com/shynome/wl"
	"github.com/shynome/wl/ortc"
	"github.com/shynome/wlsocks/lens"
)

func runServer() {

	l := wl.Listen()

	udpAddr := net.UDPAddr{IP: net.ParseIP(args.addr), Port: args.port}
	udpListener := try.To1(net.ListenUDP("udp", &udpAddr))

	settingEngine := webrtc.SettingEngine{}
	settingEngine.DetachDataChannels()
	settingEngine.SetICEUDPMux(webrtc.NewICEUDPMux(nil, udpListener))
	api := webrtc.NewAPI(webrtc.WithSettingEngine(settingEngine))

	lens := &Lens{
		Endpoint: args.lens,
		User:     args.user,
		Pass:     args.pass,
	}

	s := &Server{
		listener: l,
		wrtcApi:  api,
		lens:     lens,
	}
	stream := try.To1(s.openStream())
	go s.serveLoop(stream.Events)

	fmt.Println("server started")
	server := try.To1(socks5.New(&socks5.Config{}))
	try.To(server.Serve(l))
}

type Server struct {
	lens     *Lens
	wrtcApi  *webrtc.API
	listener *wl.Listener
}

func (s *Server) openStream() (stream *eventsource.Stream, err error) {
	defer err2.Return(&err)
	endpoint := try.To1(WithTopic(args.lens, lens.LinkTopic))
	req := try.To1(http.NewRequest(http.MethodGet, endpoint, nil))
	req.SetBasicAuth(args.user, args.pass)
	stream = try.To1(eventsource.SubscribeWithRequest("", req))
	return
}

func (s *Server) serveLoop(evs chan eventsource.Event) {
	for ev := range evs {
		go s.handleConnect(ev)
	}
}

func (s *Server) handleConnect(ev eventsource.Event) (err error) {
	defer err2.Return(&err)

	var (
		api = s.wrtcApi
	)

	var info lens.ConnectInfo
	try.To(json.Unmarshal([]byte(ev.Data()), &info))

	config := webrtc.Configuration{
		ICEServers: info.Ices,
	}
	var pc = try.To1(api.NewPeerConnection(config))

	var offer ortc.Signal = info.Offer
	roffer := try.To1(ortc.HandleConnect(pc, offer))
	rofferBytes := try.To1(json.Marshal(roffer))
	try.To(s.lens.Dial(lens.LinkTopic, ev.Id(), rofferBytes))

	peer := &wl.Peer{PC: pc}
	s.listener.Add(peer)
	pc.OnConnectionStateChange(func(pcs webrtc.PeerConnectionState) {
		if pcs == webrtc.PeerConnectionStateDisconnected {
			s.listener.Remove(peer)
			peer.Close()
		}
	})

	return
}
