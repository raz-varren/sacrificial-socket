/*
Package ssmongo provides a ss.MultihomeBackend interface that uses MongoDB for synchronizing broadcasts and roomcasts between multiple Sacrificial Socket instances.
*/
package ssmongo

import (
	"encoding/json"
	ss "github.com/raz-varren/sacrificial-socket"
	"github.com/raz-varren/sacrificial-socket/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
	"sync"
	"time"
)

//MMHB implements ss.MultihomeBackend and uses MongoDB to syncronize between
//multiple machines running ss.SocketServer
type MMHB struct {
	session       *mgo.Session
	serverC       *mgo.Collection
	roomcastC     *mgo.Collection
	broadcastC    *mgo.Collection
	server        backendServer
	pollFrequency time.Duration
	l             *sync.RWMutex
}

//NewMongoDBBackend returns a new instance of MMHB which satisfies the ss.MultihomeBackend interface.
//A new database "SSMultihome" will be created at the specified mongoURL, and under it 3 collections "activeServers",
//"ss.roomcasts", and "ss.broadcasts" will be created if they don't already exist.
//
//serverName must be unique per running ss.SocketServer instance, otherwise broadcasts, and roomcasts
//will not propogate correctly to the other running instances
//
//serverGroup is used to break up broadcast and roomcast domains between multiple ss.SocketServer instances.
//Most of the time you will want this to be the same for all of your running ss.SocketServer instances
//
//pollFrequency is used to determine how frequently MongoDB is queried for broadcasts or roomcasts
func NewBackend(mongoURL, serverName, serverGroup string, pollFrequency time.Duration) *MMHB {
	m, err := mgo.Dial(mongoURL)
	log.CheckFatal(err)
	db := m.DB("SSMultihome")

	s := backendServer{
		ServerName:  serverName,
		ServerGroup: serverGroup,
		l:           &sync.RWMutex{},
	}

	return &MMHB{
		session:       m,
		serverC:       db.C("ss.activeServers"),
		roomcastC:     db.C("ss.roomcasts"),
		broadcastC:    db.C("ss.broadcasts"),
		server:        s,
		pollFrequency: pollFrequency,
		l:             &sync.RWMutex{},
	}
}

func (mmhb *MMHB) getActiveServers() []backendServer {
	server := mmhb.getServer()

	var servers []backendServer

	err := mmhb.serverC.Find(bson.M{
		"ServerGroup": server.ServerGroup,
		"ServerName":  bson.M{"$ne": server.ServerName},
	}).All(&servers)
	if err != nil {
		log.Err.Println(err)
	}
	return servers
}

//Init will create the "SSMultihome" database along with the "ss.activeServers", "ss.broadcasts", and "ss.roomcasts"
//collections, as well as any neccessary indexes
func (mmhb *MMHB) Init() {
	cols := []*mgo.Collection{mmhb.serverC, mmhb.broadcastC, mmhb.roomcastC}
	indexes := []mgo.Index{
		mgo.Index{
			Key:         []string{"Expire"},
			ExpireAfter: time.Second * 1,
		},
		mgo.Index{
			Key: []string{"ServerGroup"},
		},
		mgo.Index{
			Key: []string{"ServerName"},
		},
	}
	for _, col := range cols {
		for _, i := range indexes {
			err := col.EnsureIndex(i)
			if err != nil {
				log.Err.Println(err)
			}
		}
	}

	mmhb.beat()
	go mmhb.heartbeat()
}

//Shutdown will remove this server from the activeServers collection
func (mmhb *MMHB) Shutdown() {
	defer mmhb.session.Close()
	server := mmhb.getServer()
	err := mmhb.serverC.Remove(bson.M{"ServerGroup": server.ServerGroup, "ServerName": server.ServerName})
	if err != nil {
		log.Err.Println(err)
	}
}

//BroadcastToBackend will insert one broadcast document into the ss.broadcasts collection for each
//server in the activeServers collection excluding itself, each time BroadcastToBackend is called.
//
//See documentation on the ss.MultihomeBackend interface for more information
func (mmhb *MMHB) BroadcastToBackend(b *ss.BroadcastMsg) {
	servers := mmhb.getActiveServers()

	if len(servers) == 0 {
		return
	}
	bulk := mmhb.broadcastC.Bulk()
	d, isJ := isJSON(b.Data)
	for _, s := range servers {
		bcast := broadcast{
			ServerName:  s.ServerName,
			ServerGroup: s.ServerGroup,
			EventName:   b.EventName,
			Data:        d,
			JSON:        isJ,
			Read:        false,
		}
		bcast.setNextExpire()
		bulk.Insert(bcast)
	}
	_, err := bulk.Run()
	if err != nil {
		log.Err.Println(err)
	}
}

//RoomcastToBackend will insert one roomcast document into the roomcasts collection for each
//server in the activeServers collection excluding itself, each time RoomcastToBackend is called.
//
//See documentation on the ss.MultihomeBackend interface for more information
func (mmhb *MMHB) RoomcastToBackend(r *ss.RoomMsg) {
	servers := mmhb.getActiveServers()

	if len(servers) == 0 {
		return
	}
	bulk := mmhb.roomcastC.Bulk()
	d, isJ := isJSON(r.Data)
	for _, s := range servers {
		rcast := roomcast{
			ServerName:  s.ServerName,
			ServerGroup: s.ServerGroup,
			RoomName:    r.RoomName,
			EventName:   r.EventName,
			Data:        d,
			JSON:        isJ,
			Read:        false,
		}
		rcast.setNextExpire()
		bulk.Insert(rcast)
	}
	_, err := bulk.Run()
	if err != nil {
		log.Err.Println(err)
	}
}

//BroadcastFromBackend polls the ss.broadcasts collection, based on the pollFrequency provided to NewBackend, for new messages designated
//to this serverName and inserts a ss.BroadcastMsg into b to be dispatched by the server
//
//See documentation on the ss.MultihomeBackend interface for more information
func (mmhb *MMHB) BroadcastFromBackend(b chan<- *ss.BroadcastMsg) {
	server := mmhb.getServer()
	for {
		time.Sleep(mmhb.pollFrequency)

		q := mmhb.broadcastC.Find(bson.M{
			"ServerName":  server.ServerName,
			"ServerGroup": server.ServerGroup,
			"Read":        false,
		}).Sort("Expire")

		count, err := q.Count()
		if err == io.EOF {
			panic(err)
		}
		if err != nil {
			log.Err.Println(err)
			continue
		}
		if count == 0 {
			continue
		}

		bulk := mmhb.broadcastC.Bulk()
		iter := q.Iter()
		var bcast broadcast
		i := 0
		for iter.Next(&bcast) {
			var d interface{}
			d = bcast.Data
			if bcast.JSON {
				d = make(map[string]interface{})
				err = json.Unmarshal(bcast.Data.([]byte), &d)
				if err != nil {
					log.Err.Println(err)
					d = ""
				}
			}
			b <- &ss.BroadcastMsg{bcast.EventName, d}
			bcast.expireNow()
			bcast.Read = true
			bulk.Update(bson.M{"_id": bcast.ID}, bson.M{"$set": bcast})
			i++
			if i >= 900 {
				_, err = bulk.Run()
				if err != nil {
					log.Err.Println(err)
				}
				bulk = mmhb.broadcastC.Bulk()
				i = 0
			}
		}
		_, err = bulk.Run()
		if err != nil {
			log.Err.Println(err)
		}
	}
}

//RoomcastFromBackend polls the roomcasts collection, based on the pollFrequency provided to NewBackend, for new messages designated
//to this serverName and inserts a ss.RoomMsg into r to be dispatched by the ss.SocketServer
//
//See documentation on the ss.MultihomeBackend interface for more information
func (mmhb *MMHB) RoomcastFromBackend(r chan<- *ss.RoomMsg) {
	server := mmhb.getServer()
	for {
		time.Sleep(mmhb.pollFrequency)
		q := mmhb.roomcastC.Find(bson.M{
			"ServerName":  server.ServerName,
			"ServerGroup": server.ServerGroup,
			"Read":        false,
		}).Sort("Expire")

		count, err := q.Count()

		if err == io.EOF {
			panic(err)
		}
		if err != nil {
			log.Err.Println(err)
			continue
		}
		if count == 0 {
			continue
		}

		bulk := mmhb.roomcastC.Bulk()
		iter := q.Iter()
		var rcast roomcast
		i := 0
		for iter.Next(&rcast) {
			var d interface{}
			d = rcast.Data
			if rcast.JSON {
				d = make(map[string]interface{})
				err = json.Unmarshal(rcast.Data.([]byte), &d)
				if err != nil {
					log.Err.Println(err)
					d = ""
				}
			}
			r <- &ss.RoomMsg{rcast.RoomName, rcast.EventName, d}
			rcast.expireNow()
			rcast.Read = true
			bulk.Update(bson.M{"_id": rcast.ID}, bson.M{"$set": rcast})
			i++
			if i >= 900 {
				_, err = bulk.Run()
				if err != nil {
					log.Err.Println(err)
				}
				bulk = mmhb.roomcastC.Bulk()
				i = 0
			}
		}
		_, err = bulk.Run()
		if err != nil {
			log.Err.Println(err)
		}
	}
}

//beat updates the Expire key for this server in the activeServers collection
func (mmhb *MMHB) beat() {
	server := mmhb.getServer()
	server.setNextExpire()
	_, err := mmhb.serverC.Upsert(bson.M{
		"ServerName":  server.ServerName,
		"ServerGroup": server.ServerGroup,
	}, bson.M{
		"$set": server,
	})

	if err != nil {
		log.Err.Println(err)
	}
}

//heartbeat calls beat every minute
func (mmhb *MMHB) heartbeat() {
	for {
		time.Sleep(time.Minute * 1)
		mmhb.beat()
	}
}

func (mmhb *MMHB) getServer() backendServer {
	mmhb.l.RLock()
	s := mmhb.server
	mmhb.l.RUnlock()
	return s
}

func isJSON(in interface{}) (interface{}, bool) {
	switch i := in.(type) {
	case string, []byte:
		return i, false
	default:
		j, err := json.Marshal(i)
		if err != nil {
			log.Err.Println(err)
			return "", false
		}
		return j, true
	}
}
