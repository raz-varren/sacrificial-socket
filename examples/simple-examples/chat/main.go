/*
A simple web chat app example that does not implement any multihome backends.
*/
package main

import (
	"encoding/json"
	"github.com/raz-varren/log"
	ss "github.com/raz-varren/sacrificial-socket"
	"net/http"
)

func main() {
	s := ss.NewServer()

	s.On("join", join)
	s.On("message", message)

	http.Handle("/socket", s)
	http.Handle("/", http.FileServer(http.Dir("webroot")))

	log.Err.Fatalln(http.ListenAndServe(":80", nil))
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
