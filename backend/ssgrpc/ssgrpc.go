/*
Package ssgrpc provides a ss.MultihomeBackend interface that uses grpc with profobufs for synchronizing broadcasts and roomcasts between multiple Sacrificial Socket instances.
*/
package ssgrpc

import (
	"encoding/json"
	ss "github.com/raz-varren/sacrificial-socket"
	"github.com/raz-varren/sacrificial-socket/backend/ssgrpc/transport"
	"github.com/raz-varren/sacrificial-socket/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"strings"
	"sync"
	"time"
)

//GRPCMHB... yep that's what I'm calling it. All you need to know is that GRPCMHB
//satisfies the ss.MultihomeBackend interface
type GRPCMHB struct {
	peerList          []string
	peers             map[string]*propagateClient
	keyFile, certFile string
	sharedKey         []byte
	gServer           *grpc.Server
	pServer           *propagateServer
	serverHostPort    string
	insecure          bool

	l *sync.RWMutex
}

//NewBackend returns a GRPCMHB that will use TLS and HMAC-SHA256 signed JWTs to connect and authenticate
//to the peers in peerList
func NewBackend(tlsKeyFile, tlsCertFile, grpcHostPort string, sharedKey []byte, peerList []string) *GRPCMHB {
	return &GRPCMHB{
		peerList:       peerList,
		peers:          make(map[string]*propagateClient),
		l:              &sync.RWMutex{},
		keyFile:        tlsKeyFile,
		certFile:       tlsCertFile,
		sharedKey:      sharedKey,
		serverHostPort: grpcHostPort,
		insecure:       false,
	}
}

//NewInsecureBackend returns a GRPCMHB that will use no encryption or authentication to connect to the
//peers in peerList
//
//It is highly discouraged to use this for production systems, as all data will be sent in clear
//text and no authentication will be done on peer connections
func NewInsecureBackend(grpcHostPort string, peerList []string) *GRPCMHB {
	return &GRPCMHB{
		peerList:       peerList,
		peers:          make(map[string]*propagateClient),
		l:              &sync.RWMutex{},
		serverHostPort: grpcHostPort,
		insecure:       true,
	}
}

func (g *GRPCMHB) constructClient(peer string) {
	var host, cn string

	g.l.RLock()
	certFile := g.certFile
	sharedKey := g.sharedKey
	insecure := g.insecure
	g.l.RUnlock()

	hcn := strings.Split(peer, "@")
	if len(hcn) == 2 {
		cn = hcn[0]
		host = hcn[1]
	} else {
		hp := strings.Split(hcn[0], ":")
		cn = hp[0]
		host = hcn[0]
	}

	dialOpts := []grpc.DialOption{grpc.WithBlock()}

	if insecure {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	} else {
		tlsCred, err := credentials.NewClientTLSFromFile(certFile, cn)
		if err != nil {
			log.Err.Fatalln(err)
		}
		rpcCred := &perRPCCreds{l: &sync.RWMutex{}, sharedKey: sharedKey}

		dialOpts = append(dialOpts, grpc.WithTransportCredentials(tlsCred))
		dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(rpcCred))
	}

	conn, err := grpc.Dial(host, dialOpts...)
	if err != nil {
		log.Err.Fatalln(err)
	}
	g.l.Lock()
	g.peers[host] = &propagateClient{conn: conn, client: transport.NewPropagateClient(conn)}
	g.l.Unlock()
	log.Info.Println("grpc connected to:", host)
}

//Init sets up the grpc server and creates connections to all grpc peers
func (g *GRPCMHB) Init() {
	g.l.Lock()
	defer g.l.Unlock()
	lis, err := net.Listen("tcp", g.serverHostPort)
	if err != nil {
		log.Err.Fatalln(err)
	}

	var opts []grpc.ServerOption
	if !g.insecure {
		srvCreds, err := credentials.NewServerTLSFromFile(g.certFile, g.keyFile)
		if err != nil {
			log.Err.Fatalln(err)
		}
		opts = append(opts, grpc.Creds(srvCreds))
	}

	serv := grpc.NewServer(opts...)

	g.gServer = serv
	g.pServer = &propagateServer{sharedKey: g.sharedKey, l: &sync.RWMutex{}, insecure: g.insecure}

	transport.RegisterPropagateServer(g.gServer, g.pServer)

	go g.gServer.Serve(lis)

	for _, host := range g.peerList {
		go g.constructClient(host)
	}
}

//Shutdown stops the grpc service and closes any peer connections
func (g *GRPCMHB) Shutdown() {
	g.l.Lock()
	defer g.l.Unlock()
	g.gServer.Stop()
	for _, peer := range g.peers {
		peer.conn.Close()
	}
}

//BroadcastToBackend propagates the broadcast to all active peer connections
func (g *GRPCMHB) BroadcastToBackend(b *ss.BroadcastMsg) {
	data, dataType := getDataType(b.Data)
	bCast := &transport.Broadcast{
		Timestamp: timestamp(),
		Event:     b.EventName,
		Data:      data,
		DataType:  dataType,
	}

	g.l.RLock()
	defer g.l.RUnlock()

	for _, peer := range g.peers {
		_, err := peer.client.DoBroadcast(context.Background(), bCast)
		if err != nil {
			log.Err.Println(err)
			continue
		}
		//log.Info.Printf("round trip time: %.3f\n", roundTrip(res.Timestamp))
	}
}

//RoomcastToBackend propagates the roomcast to all active peer connections
func (g *GRPCMHB) RoomcastToBackend(r *ss.RoomMsg) {
	data, dataType := getDataType(r.Data)
	rCast := &transport.Roomcast{
		Timestamp: timestamp(),
		Room:      r.RoomName,
		Event:     r.EventName,
		Data:      data,
		DataType:  dataType,
	}

	g.l.RLock()
	defer g.l.RUnlock()

	for _, peer := range g.peers {
		_, err := peer.client.DoRoomcast(context.Background(), rCast)
		if err != nil {
			log.Err.Println(err)
			continue
		}
		//log.Info.Printf("round trip time: %.3f\n", roundTrip(res.Timestamp))
	}
}

//BroadcastFromBackend listens on the local grpc service for calls from remote peers and
//propagates broadcasts to locally connected websockets
func (g *GRPCMHB) BroadcastFromBackend(b chan<- *ss.BroadcastMsg) {
	g.pServer.l.Lock()
	defer g.pServer.l.Unlock()
	g.pServer.bChan = b
}

//RoomcastFromBackend listens on the local grpc service for calls from remote peers and
//propagates roomcasts to locally connected websockets
func (g *GRPCMHB) RoomcastFromBackend(r chan<- *ss.RoomMsg) {
	g.pServer.l.Lock()
	defer g.pServer.l.Unlock()
	g.pServer.rChan = r
}

func getDataType(in interface{}) ([]byte, transport.DataType) {
	switch i := in.(type) {
	case string:
		return []byte(i), transport.DataType_STR
	case []byte:
		return i, transport.DataType_BIN
	default:
		j, err := json.Marshal(i)
		if err != nil {
			log.Err.Println(err)
			return []byte{}, transport.DataType_STR
		}
		return j, transport.DataType_JSON
	}
}

func roundTrip(timestamp uint64) float64 {
	return (float64(time.Now().UnixNano()) - float64(timestamp)) / 1000000000
}

func timestamp() uint64 {
	return uint64(time.Now().UnixNano())
}
