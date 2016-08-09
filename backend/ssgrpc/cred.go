package ssgrpc

import (
	"golang.org/x/net/context"
	"github.com/raz-varren/sacrificial-socket/backend/ssgrpc/token"
	"github.com/raz-varren/sacrificial-socket/log"
	"sync"
	"time"
)

type perRPCCreds struct {
	tokenStr    string
	tokenExpire int64
	sharedKey   []byte
	l           *sync.RWMutex
}

func (c *perRPCCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	var tok string
	var exp int64
	var sharedKey []byte

	c.l.RLock()
	exp = c.tokenExpire
	tok = c.tokenStr
	sharedKey = c.sharedKey
	c.l.RUnlock()

	meta := make(map[string]string)

	if exp-300 < time.Now().Unix() {
		u, t, err := token.GenUserToken("ssgrpcClient", time.Hour, sharedKey)
		if err != nil {
			log.Err.Println("gen token error:", err)
			return meta, err
		}

		exp = u.EXP
		tok = t

		c.l.Lock()
		c.tokenExpire = exp
		c.tokenStr = tok
		c.l.Unlock()

		log.Info.Println("token refreshed:", tok)
	}

	meta["authorization"] = "Bearer " + tok

	return meta, nil
}

func (c *perRPCCreds) RequireTransportSecurity() bool {
	return true
}
