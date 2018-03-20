/*
Package ssredis provides a ss.MultihomeBackend interface that uses Redis for synchronizing broadcasts and roomcasts between multiple Sacrificial Socket instances.
*/
package ssredis

import (
	"github.com/go-redis/redis"
	ss "github.com/raz-varren/sacrificial-socket"
	"github.com/raz-varren/log"
	"github.com/raz-varren/sacrificial-socket/tools"
)

const (
	//the default redis.PubSub channel that will be subscribed to
	DefServerGroup = "ss-rmhb-group-default"
)

//RMHB implements the ss.MultihomeBackend interface and uses
//Redis to syncronize between multiple machines running ss.SocketServer
type RMHB struct {
	r           *redis.Client
	o           *Options
	rps         *redis.PubSub
	bps         *redis.PubSub
	roomPSName  string
	bcastPSName string
}

type Options struct {

	//ServerName is a unique name for the ss.MultihomeBackend instance.
	//This name must be unique per backend instance or the backend
	//will not broadcast and roomcast properly.
	//
	//Leave this name blank to auto generate a unique name.
	ServerName string

	//ServerGroup is the server pool name that this instance's broadcasts
	//and roomcasts will be published to. This can be used to break up
	//ss.MultihomeBackend instances into separate domains.
	//
	//Leave this empty to use the default group "ss-rmhb-group-default"
	ServerGroup string
}

//NewBackend creates a new *RMHB specified by redis Options and ssredis Options
func NewBackend(rOpts *redis.Options, ssrOpts *Options) (*RMHB, error) {
	rClient := redis.NewClient(rOpts)
	_, err := rClient.Ping().Result()
	if err != nil {
		return nil, err
	}

	if ssrOpts == nil {
		ssrOpts = &Options{}
	}

	if ssrOpts.ServerGroup == "" {
		ssrOpts.ServerGroup = DefServerGroup
	}

	if ssrOpts.ServerName == "" {
		ssrOpts.ServerName = tools.UID()
	}

	roomPSName := ssrOpts.ServerGroup + ":_ss_roomcasts"
	bcastPSName := ssrOpts.ServerGroup + ":_ss_broadcasts"

	rmhb := &RMHB{
		r:           rClient,
		rps:         rClient.Subscribe(roomPSName),
		bps:         rClient.Subscribe(bcastPSName),
		roomPSName:  roomPSName,
		bcastPSName: bcastPSName,
		o:           ssrOpts,
	}

	return rmhb, nil
}

//Init is just here to satisfy the ss.MultihomeBackend interface.
func (r *RMHB) Init() {

}

//Shutdown closes the subscribed redis channel, then the redis connection.
func (r *RMHB) Shutdown() {
	r.rps.Close()
	r.bps.Close()
	r.r.Close()
}

//BroadcastToBackend will publish a broadcast message to the redis backend
func (r *RMHB) BroadcastToBackend(b *ss.BroadcastMsg) {
	t := &transmission{
		ServerName: r.o.ServerName,
		EventName:  b.EventName,
		Data:       b.Data,
	}

	data, err := t.toJSON()
	if err != nil {
		log.Err.Println(err)
		return
	}

	err = r.r.Publish(r.bcastPSName, string(data)).Err()
	if err != nil {
		log.Err.Println(err)
	}
}

//RoomcastToBackend will publish a roomcast message to the redis backend
func (r *RMHB) RoomcastToBackend(rm *ss.RoomMsg) {
	t := &transmission{
		ServerName: r.o.ServerName,
		EventName:  rm.EventName,
		RoomName:   rm.RoomName,
		Data:       rm.Data,
	}

	data, err := t.toJSON()
	if err != nil {
		log.Err.Println(err)
		return
	}

	err = r.r.Publish(r.roomPSName, string(data)).Err()
	if err != nil {
		log.Err.Println(err)
	}
}

//BroadcastFromBackend will receive broadcast messages from redis and propogate them to the neccessary sockets
func (r *RMHB) BroadcastFromBackend(bc chan<- *ss.BroadcastMsg) {
	bChan := r.bps.Channel()

	for d := range bChan {
		var t transmission

		err := t.fromJSON([]byte(d.Payload))
		if err != nil {
			log.Err.Println(err)
			continue
		}

		if t.ServerName == r.o.ServerName {
			continue
		}

		bc <- &ss.BroadcastMsg{
			EventName: t.EventName,
			Data:      t.Data,
		}
	}
}

//RoomcastFromBackend will receive roomcast messages from redis and propogate them to the neccessary sockets
func (r *RMHB) RoomcastFromBackend(rc chan<- *ss.RoomMsg) {
	rChan := r.rps.Channel()

	for d := range rChan {
		var t transmission

		err := t.fromJSON([]byte(d.Payload))
		if err != nil {
			log.Err.Println(err)
			continue
		}

		if t.ServerName == r.o.ServerName {
			continue
		}

		rc <- &ss.RoomMsg{
			EventName: t.EventName,
			RoomName:  t.RoomName,
			Data:      t.Data,
		}
	}
}
