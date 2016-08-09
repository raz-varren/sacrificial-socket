package main

import (
	"encoding/json"
	"log"
	"net/http"
	ss "github.com/raz-varren/sacrificial-socket"
)

func main() {
	s := ss.NewServer()

	s.On("join", join)
	s.On("message", message)

	http.Handle("/socket", s.WebHandler())
	http.Handle("/", http.FileServer(http.Dir("webroot")))

	log.Fatalln(http.ListenAndServe(":80", nil))
}

func join(s *ss.Socket, data []byte) {
	//just one room at a time for the simple example
	currentRooms := s.GetRooms()
	for _, room := range currentRooms {
		s.Leave(room)
	}
	s.Join(string(data))
	s.Emit("joinedRoom", string(data))
}

type msg struct {
	Room    string
	Message string
}

func message(s *ss.Socket, data []byte) {
	var m msg
	json.Unmarshal(data, &m)
	s.Roomcast(m.Room, "message", m.Message)
}
