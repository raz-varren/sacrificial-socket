package ssmongo

import (
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
)

type backendServer struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	ServerName  string        `bson:"ServerName"`
	ServerGroup string        `bson:"ServerGroup"`
	Expire      time.Time     `bson:"Expire"`
	l           *sync.RWMutex
}

type broadcast struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	ServerName  string        `bson:"ServerName"`
	ServerGroup string        `bson:"ServerGroup"`
	Expire      time.Time     `bson:"Expire"`
	EventName   string        `bson:"EventName"`
	Data        interface{}   `bson:"Data"`
	JSON        bool          `bson:"JSON"`
	Read        bool          `bson:"Read"`
}

type roomcast struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	ServerName  string        `bson:"ServerName"`
	ServerGroup string        `bson:"ServerGroup"`
	Expire      time.Time     `bson:"Expire"`
	RoomName    string        `bson:"RoomName"`
	EventName   string        `bson:"EventName"`
	Data        interface{}   `bson:"Data"`
	JSON        bool          `bson:"JSON"`
	Read        bool          `bson:"Read"`
}

func (s *backendServer) setNextExpire() {
	s.l.Lock()
	defer s.l.Unlock()
	s.Expire = time.Now().Add(time.Minute * 5)
}

func (b *broadcast) setNextExpire() {
	b.Expire = time.Now().Add(time.Minute * 5)
}

func (b *broadcast) expireNow() {
	b.Expire = time.Now().Add(time.Minute * 5 * -1)
}

func (r *roomcast) setNextExpire() {
	r.Expire = time.Now().Add(time.Minute * 5)
}

func (r *roomcast) expireNow() {
	r.Expire = time.Now().Add(time.Minute * 5 * -1)
}
