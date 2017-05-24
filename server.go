/*
Package ss (Sacrificial-Socket) is a Go server library and pure JS client library for managing communication between websockets, that has an API similar to Socket.IO, but feels less... well, Javascripty. Socket.IO is great, but nowadays all modern browsers support websockets natively, so in most cases there is no need to have websocket simulation fallbacks like XHR long polling or Flash. Removing these allows Sacrificial-Socket to be lightweight and very performant.

Sacrificial-Socket supports rooms, roomcasts, broadcasts, and event emitting just like Socket.IO, but with one key difference. The data passed into event functions is not an interface{} that is implied to be a string or map[string]interface{}, but is always passed in as a []byte making it easier to unmarshal into your own JSON data structs, convert to a string, or keep as binary data without the need to check the data's type before processing it. It also means there aren't any unnecessary conversions to the data between the client and the server.

Sacrificial-Socket also has a MultihomeBackend interface for syncronizing broadcasts and roomcasts across multiple instances of Sacrificial-Socket running on multiple machines. Out of the box Sacrificial-Socket provides a MultihomeBackend interface for the popular noSQL database MongoDB, and one for the not so popular GRPC protocol, for syncronizing instances on multiple machines.
*/
package ss

import (
	"github.com/raz-varren/sacrificial-socket/log"
	"golang.org/x/net/websocket"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const ( //                        ASCII chars
	startOfHeaderByte uint8 = 1 //SOH
	startOfDataByte         = 2 //STX

	SupportedSubProtocol string = "sac-sock"
)

type event struct {
	eventName    string
	eventHandler func(*Socket, []byte)
}

//SocketServer manages the coordination between
//sockets, rooms, events and the socket hub
type SocketServer struct {
	hub              *socketHub
	events           map[string]*event
	onConnectFunc    func(*Socket)
	onDisconnectFunc func(*Socket)
	l                *sync.RWMutex
}

//NewServer creates a new instance of SocketServer
func NewServer() *SocketServer {
	s := &SocketServer{
		hub:    newHub(),
		events: make(map[string]*event),
		l:      &sync.RWMutex{},
	}

	return s
}

//EnableSignalShutdown listens for linux syscalls SIGHUP, SIGINT, SIGTERM, SIGQUIT, SIGKILL and
//calls the SocketServer.Shutdown() to perform a clean shutdown. true will be passed into complete
//after the Shutdown proccess is finished
func (serv *SocketServer) EnableSignalShutdown(complete chan<- bool) {
	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGKILL)

	go func() {
		<-c
		complete <- serv.Shutdown()
	}()
}

//Shutdown closes all active sockets and triggers the Shutdown()
//method on any MultihomeBackend that is currently set.
func (serv *SocketServer) Shutdown() bool {
	log.Info.Println("shutting down")
	//complete := serv.hub.shutdown()

	serv.hub.shutdownCh <- true
	socketList := <-serv.hub.socketList

	for _, s := range socketList {
		s.Close()
	}

	if serv.hub.multihomeEnabled {
		log.Info.Println("shutting down multihome backend")
		serv.hub.multihomeBackend.Shutdown()
		log.Info.Println("backend shutdown")
	}

	log.Info.Println("shutdown")
	return true
}

//On registers event functions to be called on individual Socket connections
//when the server's socket receives an Emit from the client's socket.
//
//Any event functions registered with On, must be safe for concurrent use by multiple
//go routines
func (serv *SocketServer) On(eventName string, handleFunc func(*Socket, []byte)) {
	serv.events[eventName] = &event{eventName, handleFunc} //you think you can handle the func?
}

//OnConnect registers an event function to be called whenever a new Socket connection
//is created
func (serv *SocketServer) OnConnect(handleFunc func(*Socket)) {
	serv.onConnectFunc = handleFunc
}

//OnDisconnect registers an event function to be called as soon as a Socket connection
//is closed
func (serv *SocketServer) OnDisconnect(handleFunc func(*Socket)) {
	serv.onDisconnectFunc = handleFunc
}

//WebHandler returns a http.Handler to be passed into http.Handle
func (serv *SocketServer) WebHandler() http.Handler {
	return websocket.Server{
		Handshake: func(c *websocket.Config, r *http.Request) error {
			if !protocolSupported(c) {
				return websocket.ErrBadWebSocketProtocol
			}
			return nil
		},
		Handler: serv.loop,
	}
}

//SetMultihomeBackend registers a MultihomeBackend interface and calls it's Init() method
func (serv *SocketServer) SetMultihomeBackend(b MultihomeBackend) {
	serv.hub.setMultihomeBackend(b)
}

//Roomcast dispatches an event to all Sockets in the specified room.
func (serv *SocketServer) Roomcast(roomName, eventName string, data interface{}) {
	serv.hub.roomcast(&RoomMsg{roomName, eventName, data})
}

//Broadcast dispatches an event to all Sockets on the SocketServer.
func (serv *SocketServer) Broadcast(eventName string, data interface{}) {
	serv.hub.broadcast(&BroadcastMsg{eventName, data})
}

//loop handles all the coordination between new sockets
//reading frames and dispatching events
func (serv *SocketServer) loop(ws *websocket.Conn) {
	s := newSocket(serv, ws)
	log.Debug.Println(s.ID(), "connected")
	defer s.Close()

	serv.l.RLock()
	e := serv.onConnectFunc
	serv.l.RUnlock()

	if e != nil {
		e(s)
	}

	for {
		var msg []byte

		err := s.receive(&msg)
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Err.Println(err)
			return
		}

		eventName := ""
		contentIdx := 0

		for idx, chr := range msg {
			if chr == startOfDataByte {
				eventName = string(msg[:idx])
				contentIdx = idx + 1
				break
			}
		}
		if eventName == "" {
			continue //no event to dispatch
		}

		serv.l.RLock()
		e, exists := serv.events[eventName]
		serv.l.RUnlock()

		if exists {
			go e.eventHandler(s, msg[contentIdx:])
		}
	}
}

func protocolSupported(conf *websocket.Config) bool {
	for _, p := range conf.Protocol {
		if p == SupportedSubProtocol {
			return true
		}
	}
	return false
}
