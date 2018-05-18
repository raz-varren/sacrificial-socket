package ssgrpc

import (
	"encoding/json"
	"errors"
	"github.com/raz-varren/log"
	ss "github.com/raz-varren/sacrificial-socket"
	"github.com/raz-varren/sacrificial-socket/backend/ssgrpc/token"
	"github.com/raz-varren/sacrificial-socket/backend/ssgrpc/transport"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"strings"
	"sync"
)

var (
	ErrBadRPCCredentials = errors.New("the client provided invalid RPC credentials")

	ErrNilBroadcastChannel = errors.New("broadcast channel is not open yet")
	ErrNilRoomcastChannel  = errors.New("roomcast channel is not open yet")

	ErrBadDataType = errors.New("bad data type used")
	ErrBadContext  = errors.New("bad context used in transport")
)

type propagateServer struct {
	sharedKey []byte
	bChan     chan<- *ss.BroadcastMsg
	rChan     chan<- *ss.RoomMsg
	insecure  bool
	l         *sync.RWMutex
}

func (p *propagateServer) checkCreds(ctx context.Context) error {
	if p.insecure {
		return nil
	}

	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ErrBadContext
	}

	if auth, exists := meta["authorization"]; exists && len(auth) == 1 {
		t := strings.Split(auth[0], " ")
		if len(t) != 2 || t[0] != "Bearer" {
			return token.ErrBadBearerValue
		}

		_, err := token.ValidateUserToken(t[1], p.sharedKey)
		if err != nil {
			return err
		}
	} else {
		return ErrBadRPCCredentials
	}

	return nil
}

func (p *propagateServer) DoBroadcast(ctx context.Context, b *transport.Broadcast) (*transport.Result, error) {
	tr := &transport.Result{Timestamp: b.Timestamp, Success: false}

	err := p.checkCreds(ctx)
	if err != nil {
		log.Err.Println(err)
		return tr, err
	}

	p.l.RLock()
	bChan := p.bChan
	p.l.RUnlock()

	//channel is not open yet
	if bChan == nil {
		return tr, ErrNilBroadcastChannel
	}

	bCast := &ss.BroadcastMsg{EventName: b.Event, Data: b.Data}

	switch b.DataType {
	case transport.DataType_JSON:
		var d interface{}
		err := json.Unmarshal(b.Data, &d)
		if err != nil {
			return tr, err
		}
		bCast.Data = d
		bChan <- bCast

	case transport.DataType_STR:
		bCast.Data = string(b.Data)
		bChan <- bCast

	case transport.DataType_BIN:
		bChan <- bCast

	default:
		return tr, ErrBadDataType
	}

	tr.Success = true
	return tr, nil
}

func (p *propagateServer) DoRoomcast(ctx context.Context, r *transport.Roomcast) (*transport.Result, error) {
	tr := &transport.Result{Timestamp: r.Timestamp, Success: false}

	err := p.checkCreds(ctx)
	if err != nil {
		log.Err.Println(err)
		return tr, err
	}

	p.l.RLock()
	rChan := p.rChan
	p.l.RUnlock()
	if rChan == nil {
		return tr, ErrNilRoomcastChannel
	}

	rCast := &ss.RoomMsg{RoomName: r.Room, EventName: r.Event, Data: r.Data}

	switch r.DataType {
	case transport.DataType_JSON:
		var d interface{}
		err := json.Unmarshal(r.Data, &d)
		if err != nil {
			return tr, err
		}
		rCast.Data = d
		rChan <- rCast

	case transport.DataType_STR:
		rCast.Data = string(r.Data)
		rChan <- rCast

	case transport.DataType_BIN:
		rChan <- rCast

	default:
		return tr, ErrBadDataType
	}

	tr.Success = true
	return tr, nil
}
