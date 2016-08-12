package ss_test

import (
	"encoding/json"
	"github.com/raz-varren/sacrificial-socket"
	"log"
	"net/http"
	"os"
)

func ExampleNewServer() {
	serv := ss.NewServer()
	serv.On("echo", Echo)
	serv.On("join", Join)
	serv.On("leave", Leave)
	serv.On("roomcast", Roomcast)
	serv.On("broadcast", Broadcast)

	done := make(chan bool)

	go func() {
		serv.EnableSignalShutdown(done)
		<-done
		os.Exit(0)
	}()

	http.Handle("/socket", s.WebHandler())
	log.Fatalln(http.ListenAndServe(":8080", nil))
}

type JoinJSON struct {
	Room string
}

type LeaveJSON struct {
	Room string
}

type RoomcastJSON struct {
	Room, Event, Data string
}

type BroadcastJSON struct {
	Event, Data string
}

func Echo(s *ss.Socket, data []byte) {
	s.Emit("echo", string(data))
}

func Join(s *ss.Socket, data []byte) {
	var j JoinJSON
	err := json.Unmarshal(data, &j)
	check(err)

	s.Join(j.Room)
	s.Emit("echo", "joined: "+j.Room)
}

func Leave(s *ss.Socket, data []byte) {
	var l LeaveJSON
	err := json.Unmarshal(data, &l)
	check(err)

	s.Leave(l.Room)
	s.Emit("echo", "left: "+l.Room)
}

func Broadcast(s *ss.Socket, data []byte) {
	var b BroadcastJSON
	err := json.Unmarshal(data, &b)
	check(err)

	s.Broadcast(b.Event, b.Data)
}

func Roomcast(s *ss.Socket, data []byte) {
	var r RoomcastJSON
	err := json.Unmarshal(data, &r)
	check(err)

	s.Roomcast(r.Room, r.Event, r.Data)
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
