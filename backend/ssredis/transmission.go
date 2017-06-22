package ssredis

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/raz-varren/sacrificial-socket/log"
)

const (
	ttErr int = iota
	ttStr
	ttBin
	ttJSON
)

var (
	ErrBadDataType = errors.New("bad data type")
	ErrNoEventName = errors.New("no event name")
)

type transmission struct {
	DataType   int         `json:"d"`
	EventName  string      `json:"e"`
	RoomName   string      `json:"r,omitempty"`
	Payload    string      `json:"p"`
	ServerName string      `json:"s"`
	Data       interface{} `json:"-"`
}

func (t *transmission) toJSON() ([]byte, error) {
	data, dType := getDataType(t.Data)
	t.Payload = base64.StdEncoding.EncodeToString(data)
	t.DataType = dType

	return json.Marshal(t)
}

func (t *transmission) fromJSON(data []byte) error {
	err := json.Unmarshal(data, t)
	if err != nil {
		return err
	}

	if t.DataType == ttErr {
		return ErrBadDataType
	}

	if t.EventName == "" {
		return ErrNoEventName
	}

	d, err := base64.StdEncoding.DecodeString(t.Payload)
	if err != nil {
		return err
	}

	switch t.DataType {
	case ttStr:
		t.Data = string(d)
	case ttBin:
		t.Data = d
	case ttJSON:
		err = json.Unmarshal(d, &t.Data)
		if err != nil {
			return err
		}
	}

	return nil
}

func getDataType(in interface{}) ([]byte, int) {
	switch i := in.(type) {
	case string:
		return []byte(i), ttStr
	case []byte:
		return i, ttBin
	default:
		j, err := json.Marshal(i)
		if err != nil {
			log.Err.Println(err)
			return []byte{}, ttStr
		}
		return j, ttJSON
	}
}
