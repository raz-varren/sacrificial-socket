package ss

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/raz-varren/log"
	"math/rand"
	"sync"
	"time"
)

var (
	socketRNG = newRNG()
)

//Socket represents a websocket connection
type Socket struct {
	l      *sync.RWMutex
	id     string
	ws     *websocket.Conn
	closed bool
	serv   *SocketServer
	roomsl *sync.RWMutex
	rooms  map[string]bool
}

const (
	idLen int = 24

	typeJSON string = "J"
	typeBin         = "B"
	typeStr         = "S"
)

func newSocket(serv *SocketServer, ws *websocket.Conn) *Socket {
	s := &Socket{
		l:      &sync.RWMutex{},
		id:     newSocketID(),
		ws:     ws,
		closed: false,
		serv:   serv,
		roomsl: &sync.RWMutex{},
		rooms:  make(map[string]bool),
	}
	serv.hub.addSocket(s)
	return s
}

func newSocketID() string {
	idBuf := make([]byte, idLen)
	socketRNG.Read(idBuf)
	return base64.StdEncoding.EncodeToString(idBuf)
}

func (s *Socket) receive() ([]byte, error) {
	_, data, err := s.ws.ReadMessage()
	return data, err
}

func (s *Socket) send(msgType int, data []byte) error {
	s.l.Lock()
	defer s.l.Unlock()
	return s.ws.WriteMessage(msgType, data)
}

//InRoom returns true if s is currently a member of roomName
func (s *Socket) InRoom(roomName string) bool {
	s.roomsl.RLock()
	defer s.roomsl.RUnlock()
	inRoom := s.rooms[roomName]
	return inRoom
}

//GetRooms returns a list of rooms that s is a member of
func (s *Socket) GetRooms() []string {
	s.roomsl.RLock()
	defer s.roomsl.RUnlock()

	var roomList []string
	for room := range s.rooms {
		roomList = append(roomList, room)
	}
	return roomList
}

//Join adds s to the specified room. If the room does
//not exist, it will be created
func (s *Socket) Join(roomName string) {
	s.roomsl.Lock()
	defer s.roomsl.Unlock()
	s.serv.hub.joinRoom(&joinRequest{roomName, s})
	s.rooms[roomName] = true
}

//Leave removes s from the specified room. If s
//is not a member of the room, nothing will happen. If the room is
//empty upon removal of s, the room will be closed
func (s *Socket) Leave(roomName string) {
	s.roomsl.Lock()
	defer s.roomsl.Unlock()
	s.serv.hub.leaveRoom(&leaveRequest{roomName, s})
	delete(s.rooms, roomName)
}

//Roomcast dispatches an event to all Sockets in the specified room.
func (s *Socket) Roomcast(roomName, eventName string, data interface{}) {
	s.serv.hub.roomcast(&RoomMsg{roomName, eventName, data})
}

//Broadcast dispatches an event to all Sockets on the SocketServer.
func (s *Socket) Broadcast(eventName string, data interface{}) {
	s.serv.hub.broadcast(&BroadcastMsg{eventName, data})
}

//Emit dispatches an event to s.
func (s *Socket) Emit(eventName string, data interface{}) error {
	d, msgType := emitData(eventName, data)
	return s.send(msgType, d)
}

//ID returns the unique ID of s
func (s *Socket) ID() string {
	return s.id
}

//emitData combines the eventName and data into a payload that is understood
//by the sac-sock protocol.
func emitData(eventName string, data interface{}) ([]byte, int) {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(eventName)
	buf.WriteByte(startOfHeaderByte)

	switch d := data.(type) {
	case string:
		buf.WriteString(typeStr)
		buf.WriteByte(startOfDataByte)
		buf.WriteString(d)
		return buf.Bytes(), websocket.TextMessage

	case []byte:
		buf.WriteString(typeBin)
		buf.WriteByte(startOfDataByte)
		buf.Write(d)
		return buf.Bytes(), websocket.BinaryMessage

	default:
		buf.WriteString(typeJSON)
		buf.WriteByte(startOfDataByte)
		jsonData, err := json.Marshal(d)
		if err != nil {
			log.Err.Println(err)
		} else {
			buf.Write(jsonData)
		}
		return buf.Bytes(), websocket.TextMessage
	}
}

//Close closes the Socket connection and removes the Socket
//from any rooms that it was a member of
func (s *Socket) Close() {
	s.l.Lock()
	isAlreadyClosed := s.closed
	s.closed = true
	s.l.Unlock()

	if isAlreadyClosed { //can't reclose the socket
		return
	}

	defer log.Debug.Println(s.ID(), "disconnected")

	s.ws.Close()

	rooms := s.GetRooms()

	for _, room := range rooms {
		s.Leave(room)
	}

	s.serv.l.RLock()
	event := s.serv.onDisconnectFunc
	s.serv.l.RUnlock()

	if event != nil {
		event(s)
	}

	s.serv.hub.removeSocket(s)
}

type rng struct {
	r  *rand.Rand
	mu *sync.Mutex
}

func (r *rng) Read(b []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.r.Read(b)
}

func newRNG() *rng {
	return &rng{
		r:  rand.New(rand.NewSource(time.Now().UnixNano())),
		mu: &sync.Mutex{},
	}
}
