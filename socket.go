package ss

import (
	"bytes"
	"encoding/json"
	"github.com/raz-varren/sacrificial-socket/log"
	"github.com/raz-varren/sacrificial-socket/tools"
	"golang.org/x/net/websocket"
	"sync"
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
	idLen int = 32

	typeJSON string = "J"
	typeBin         = "B"
	typeStr         = "S"
)

var (
	idChars = []string{
		"0", "1", "2", "3", "4",
		"5", "6", "7", "8", "9",
		"A", "B", "C", "D", "E",
		"F", "G", "H", "I", "J",
		"K", "L", "M", "N", "O",
		"P", "Q", "R", "S", "T",
		"U", "V", "W", "X", "Y",
		"Z", "a", "b", "c", "d",
		"e", "f", "g", "h", "i",
		"j", "k", "l", "m", "n",
		"o", "p", "q", "r", "s",
		"t", "u", "v", "w", "x",
		"y", "z", "=", "_", "-",
		"#", ".",
	}

	idCharLen int = len(idChars) - 1
)

func newSocket(serv *SocketServer, ws *websocket.Conn) *Socket {
	buf := bytes.NewBuffer(nil)
	for i := 0; i < idLen; i++ {
		buf.WriteString(idChars[tools.RandomInt(0, idCharLen)])
	}
	s := &Socket{
		l:      &sync.RWMutex{},
		id:     buf.String(),
		ws:     ws,
		closed: false,
		serv:   serv,
		roomsl: &sync.RWMutex{},
		rooms:  make(map[string]bool),
	}
	serv.hub.addSocket(s)
	return s
}

func (s *Socket) receive(v interface{}) error {
	return websocket.Message.Receive(s.ws, v)
}

func (s *Socket) send(data interface{}) error {
	return websocket.Message.Send(s.ws, data)
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
	for room, _ := range s.rooms {
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
	return s.send(emitData(eventName, data))
}

//ID returns the unique ID of s
func (s *Socket) ID() string {
	s.l.RLock()
	defer s.l.RUnlock()
	id := s.id
	return id
}

//emitData combines the eventName and data into a payload that is understood
//by the sac-sock protocol. It will return either a string or a []byte
func emitData(eventName string, data interface{}) interface{} {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(eventName)
	buf.WriteByte(startOfHeaderByte)

	switch d := data.(type) {
	case string:
		buf.WriteString(typeStr)
		buf.WriteByte(startOfDataByte)
		buf.WriteString(d)
		return buf.String()

	case []byte:
		buf.WriteString(typeBin)
		buf.WriteByte(startOfDataByte)
		buf.Write(d)
		return buf.Bytes()

	default:
		buf.WriteString(typeJSON)
		buf.WriteByte(startOfDataByte)
		jsonData, err := json.Marshal(d)
		if err != nil {
			log.Err.Println(err)
		} else {
			buf.Write(jsonData)
		}
		return buf.String()
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

	err := s.ws.Close()
	if err != nil {
		log.Err.Println(err)
	}

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
