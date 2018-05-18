//Package ssdummy is a mockup multihome backend. It only exists to help illustrate the mechanics of the ss.MultihomeBackend interface. ssdumy should not be used in production as it provides no actual multihome functionality.
package ssdummy

import (
	"github.com/raz-varren/log"
	ss "github.com/raz-varren/sacrificial-socket"
	"time"
)

//DummyMHB is a mockup multihome backend that satisfies the ss.MultihomeBackend interface.
type DummyMHB struct {
	fromRoomName           string
	fromBroadcastEventName string
	fromRoomcastEventName  string
	fromData               interface{}
	fromBackendFrequency   time.Duration
}

//NewBackend returns a DummyMHB which satisfies the ss.MultihomeBackend interface.
//
//fromRoomName, fromBroacastEventName, fromRoomcastEventName, and fromData are the RoomName, EventName, and Data
//that will be used to send broadcasts and roomcasts to local websockets from the dummy backend.
//
//fromBackendFrequency is how often broadcasts and roomcasts will be sent to local websockets from the dummy backend.
func NewBackend(fromRoomName, fromBroadcastEventName, fromRoomcastEventName string, fromData interface{}, fromBackendFrequency time.Duration) *DummyMHB {
	return &DummyMHB{
		fromRoomName:           fromRoomName,
		fromBroadcastEventName: fromBroadcastEventName,
		fromRoomcastEventName:  fromRoomcastEventName,
		fromData:               fromData,
		fromBackendFrequency:   fromBackendFrequency,
	}
}

//Init prints a log message when the Init method is called
func (d *DummyMHB) Init() {
	log.Info.Println("dummy multihome backend initialized")
}

//Shutdown prints a log message when the Shutdown method is called
func (d *DummyMHB) Shutdown() {
	log.Info.Println("dummy multihome backend shutdown")
}

//BroadcastToBackend prints out a broadcast message whenever a local websocket sends a broadcast to the backend
func (d *DummyMHB) BroadcastToBackend(b *ss.BroadcastMsg) {
	log.Info.Println("broadcast message to backend, EventName:", b.EventName, "Data:", b.Data)
}

//RoomcastToBackend prints out a roomcast message whenever a local websocket sends a roomcast to the backend
func (d *DummyMHB) RoomcastToBackend(r *ss.RoomMsg) {
	log.Info.Println("roomcast message to backend, RoomName:", r.RoomName, "EventName:", r.EventName, "Data:", r.Data)
}

//BroadcastFromBackend prints a log message when the method is called and when it inserts ss.BroadcastMsg messages
//into the bCast channel according to the arguments used in NewBackend
func (d *DummyMHB) BroadcastFromBackend(bCast chan<- *ss.BroadcastMsg) {
	log.Info.Println("BroadcastFromBackend method called")
	for {
		time.Sleep(d.fromBackendFrequency)
		bCast <- &ss.BroadcastMsg{EventName: d.fromBroadcastEventName, Data: d.fromData}
		log.Info.Println("broadcast message from backend, EventName:", d.fromBroadcastEventName, "Data:", d.fromData)
	}
}

//RoomcastFromBackend prints a log message when the method is called and when it inserts ss.RoomMsg messages
//into the rCast channel according to the arguments used in NewBackend
func (d *DummyMHB) RoomcastFromBackend(rCast chan<- *ss.RoomMsg) {
	log.Info.Println("RoomcastFromBackend method called")
	for {
		time.Sleep(d.fromBackendFrequency)
		rCast <- &ss.RoomMsg{RoomName: d.fromRoomName, EventName: d.fromRoomcastEventName, Data: d.fromData}
		log.Info.Println("roomcast message from backend, RoomName:", d.fromRoomName, "EventName:", d.fromRoomcastEventName, "Data:", d.fromData)
	}
}
