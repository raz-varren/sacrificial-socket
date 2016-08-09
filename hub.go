package ss

type socketHub struct {
	sockets map[string]*Socket
	rooms   map[string]*room

	shutdownCh       chan bool
	socketList       chan []*Socket
	addCh            chan *Socket
	delCh            chan *Socket
	joinRoomCh       chan *joinRequest
	leaveRoomCh      chan *leaveRequest
	roomMsgCh        chan *RoomMsg
	broomcastCh      chan *RoomMsg //for passing data from the backend
	broadcastCh      chan *BroadcastMsg
	bbroadcastCh     chan *BroadcastMsg
	multihomeEnabled bool
	multihomeBackend MultihomeBackend
}

type room struct {
	name    string
	sockets map[string]*Socket
}

type joinRequest struct {
	roomName string
	socket   *Socket
}

type leaveRequest struct {
	roomName string
	socket   *Socket
}

//RoomMsg represents an event to be dispatched to a room of sockets
type RoomMsg struct {
	RoomName  string
	EventName string
	Data      interface{}
}

//BroadcastMsg represents an event to be dispatched to all Sockets on the SocketServer
type BroadcastMsg struct {
	EventName string
	Data      interface{}
}

func (h *socketHub) addSocket(s *Socket) {
	h.addCh <- s
}

func (h *socketHub) removeSocket(s *Socket) {
	h.delCh <- s
}

func (h *socketHub) joinRoom(j *joinRequest) {
	h.joinRoomCh <- j
}

func (h *socketHub) leaveRoom(l *leaveRequest) {
	h.leaveRoomCh <- l
}

func (h *socketHub) roomcast(msg *RoomMsg) {
	h.roomMsgCh <- msg
}

func (h *socketHub) broadcast(b *BroadcastMsg) {
	h.broadcastCh <- b
}

func (h *socketHub) setMultihomeBackend(b MultihomeBackend) {
	if h.multihomeEnabled {
		return //can't have two backends... yet
	}

	h.multihomeBackend = b
	h.multihomeEnabled = true

	h.multihomeBackend.Init()

	go h.multihomeBackend.BroadcastFromBackend(h.bbroadcastCh)
	go h.multihomeBackend.RoomcastFromBackend(h.broomcastCh)
}

func (h *socketHub) listen() {
	for {
		select {
		case c := <-h.addCh:
			h.sockets[c.ID()] = c
		case c := <-h.delCh:
			delete(h.sockets, c.ID())
		case c := <-h.joinRoomCh:
			if _, exists := h.rooms[c.roomName]; !exists { //make the room if it doesn't exist
				h.rooms[c.roomName] = &room{c.roomName, make(map[string]*Socket)}
			}
			h.rooms[c.roomName].sockets[c.socket.ID()] = c.socket
		case c := <-h.leaveRoomCh:
			if room, exists := h.rooms[c.roomName]; exists {
				delete(room.sockets, c.socket.ID())
				if len(room.sockets) == 0 { //room is empty, delete it
					delete(h.rooms, c.roomName)
				}
			}
		case c := <-h.roomMsgCh:
			if room, exists := h.rooms[c.RoomName]; exists {
				for _, s := range room.sockets {
					s.Emit(c.EventName, c.Data)
				}
			}
			if h.multihomeEnabled { //the room may exist on the other end
				go h.multihomeBackend.RoomcastToBackend(c)
			}
		case c := <-h.broomcastCh:
			if room, exists := h.rooms[c.RoomName]; exists {
				for _, s := range room.sockets {
					s.Emit(c.EventName, c.Data)
				}
			}
		case c := <-h.broadcastCh:
			for _, s := range h.sockets {
				s.Emit(c.EventName, c.Data)
			}
			if h.multihomeEnabled {
				go h.multihomeBackend.BroadcastToBackend(c)
			}
		case c := <-h.bbroadcastCh:
			for _, s := range h.sockets {
				s.Emit(c.EventName, c.Data)
			}
		case _ = <-h.shutdownCh:
			var socketList []*Socket
			for _, s := range h.sockets {
				socketList = append(socketList, s)
			}
			h.socketList <- socketList
		}
	}
}

func newHub() *socketHub {
	h := &socketHub{
		shutdownCh:       make(chan bool),
		socketList:       make(chan []*Socket),
		sockets:          make(map[string]*Socket),
		rooms:            make(map[string]*room),
		addCh:            make(chan *Socket),
		delCh:            make(chan *Socket),
		joinRoomCh:       make(chan *joinRequest),
		leaveRoomCh:      make(chan *leaveRequest),
		roomMsgCh:        make(chan *RoomMsg),
		broomcastCh:      make(chan *RoomMsg),
		broadcastCh:      make(chan *BroadcastMsg),
		bbroadcastCh:     make(chan *BroadcastMsg),
		multihomeEnabled: false,
	}

	go h.listen()

	return h
}

//MultihomeBackend is an interface for implementing a mechanism
//to syncronize Broadcasts and Roomcasts to multiple SocketServers
//running separate machines.
//
//Sacrificial-Socket provides a MultihomeBackend for use with MongoDB
//in sacrificial-socket/backend
type MultihomeBackend interface {
	//Init is called as soon as the MultihomeBackend is
	//registered using SocketServer.SetMultihomeBackend
	Init()

	//Shutdown is called immediately after all sockets have
	//been closed
	Shutdown()

	//BroadcastToBackend is called everytime a BroadcastMsg is
	//sent by a Socket
	//
	//BroadcastToBackend must be safe for concurrent use by multiple
	//go routines
	BroadcastToBackend(*BroadcastMsg)

	//RoomcastToBackend is called everytime a RoomMsg is sent
	//by a socket, even if none of this server's sockets are
	//members of that room
	//
	//RoomcastToBackend must be safe for concurrent use by multiple
	//go routines
	RoomcastToBackend(*RoomMsg)

	//BroadcastFromBackend is called once and only once as a go routine as
	//soon as the MultihomeBackend is registered using
	//SocketServer.SetMultihomeBackend
	//
	//b consumes a BroadcastMsg and dispatches
	//it to all sockets on this server
	BroadcastFromBackend(b chan<- *BroadcastMsg)

	//RoomcastFromBackend is called once and only once as a go routine as
	//soon as the MultihomeBackend is registered using
	//SocketServer.SetMultihomeBackend
	//
	//r consumes a RoomMsg and dispatches it to all sockets
	//that are members the specified room
	RoomcastFromBackend(r chan<- *RoomMsg)
}
