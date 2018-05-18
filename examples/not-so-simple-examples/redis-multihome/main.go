/*
A complex web app example that implements ssredis for synchronizing multiple Sacrificial Socket instances
*/
package main

import (
	"encoding/json"
	"flag"
	"github.com/go-redis/redis"
	"github.com/raz-varren/log"
	"github.com/raz-varren/sacrificial-socket"
	"github.com/raz-varren/sacrificial-socket/backend/ssredis"
	"net/http"
	"os"
)

type roomcast struct {
	Room string `json:"room"`
	Data string `json:"data"`
}

type message struct {
	Message string `json:"message"`
}

var (
	webPort   = flag.String("webport", ":8081", "host:port number used for webpage and socket connections")
	redisPort = flag.String("redisport", ":6379", "host:port number used to connect to the redis server")
	key       = flag.String("key", "./keys/snakeoil.key", "tls key used for https")
	cert      = flag.String("cert", "./keys/snakeoil.crt", "tls cert used for https")
	pass      = flag.String("p", "", "redis password, if there is one")
	db        = flag.Int("db", 0, "redis db (default 0)")
)

func main() {
	flag.Parse()

	log.SetDefaultLogger(log.NewLogger(os.Stdout, log.LogLevelDbg))

	s := ss.NewServer()

	s.On("echo", Echo)
	s.On("echobin", EchoBin)
	s.On("echojson", EchoJSON)
	s.On("join", Join)
	s.On("leave", Leave)
	s.On("roomcast", Roomcast)
	s.On("roomcastbin", RoomcastBin)
	s.On("roomcastjson", RoomcastJSON)
	s.On("broadcast", Broadcast)
	s.On("broadcastbin", BroadcastBin)
	s.On("broadcastjson", BroadcastJSON)

	b, err := ssredis.NewBackend(&redis.Options{
		Addr:     *redisPort,
		Password: *pass,
		DB:       *db,
	}, nil)

	if err != nil {
		log.Err.Fatalln(err)
	}

	s.SetMultihomeBackend(b)

	c := make(chan bool)
	s.EnableSignalShutdown(c)

	go func() {
		<-c
		os.Exit(0)
	}()

	http.Handle("/socket", s)
	http.Handle("/", http.FileServer(http.Dir("webroot")))

	if *cert == "" || *key == "" {
		err = http.ListenAndServe(*webPort, nil)
	} else {
		err = http.ListenAndServeTLS(*webPort, *cert, *key, nil)
	}

	if err != nil {
		log.Err.Fatalln(err)
	}
}

func Echo(s *ss.Socket, data []byte) {
	s.Emit("echo", string(data))
}

func EchoBin(s *ss.Socket, data []byte) {
	s.Emit("echobin", data)
}

func EchoJSON(s *ss.Socket, data []byte) {
	var m message
	err := json.Unmarshal(data, &m)
	check(err)

	s.Emit("echojson", m)
}

func Join(s *ss.Socket, data []byte) {
	d := string(data)
	s.Join(d)
	s.Emit("echo", "joined room:"+d)
}

func Leave(s *ss.Socket, data []byte) {
	d := string(data)
	s.Leave(d)
	s.Emit("echo", "left room:"+d)
}

func Roomcast(s *ss.Socket, data []byte) {
	var r roomcast
	err := json.Unmarshal(data, &r)
	check(err)

	s.Roomcast(r.Room, "roomcast", r.Data)
}

func RoomcastBin(s *ss.Socket, data []byte) {
	var r roomcast
	err := json.Unmarshal(data, &r)
	check(err)

	s.Roomcast(r.Room, "roomcastbin", []byte(r.Data))
}

func RoomcastJSON(s *ss.Socket, data []byte) {
	var r roomcast
	err := json.Unmarshal(data, &r)
	check(err)

	s.Roomcast(r.Room, "roomcastjson", r)
}

func Broadcast(s *ss.Socket, data []byte) {
	s.Broadcast("broadcast", string(data))
}

func BroadcastBin(s *ss.Socket, data []byte) {
	s.Broadcast("broadcastbin", data)
}

func BroadcastJSON(s *ss.Socket, data []byte) {
	var m message
	err := json.Unmarshal(data, &m)
	check(err)

	s.Broadcast("broadcastjson", m)
}

func check(err error) {
	if err != nil {
		log.Err.Println(err)
	}
}
