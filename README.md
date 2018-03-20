Sacrificial-Socket [![GoDoc](https://godoc.org/github.com/raz-varren/sacrificial-socket?status.svg)](https://godoc.org/github.com/raz-varren/sacrificial-socket)
==================

A Go server library and pure JS client library for managing communication between websockets, that has an API similar to Socket.IO, but feels less... well, *Javascripty*. Socket.IO is great, but nowadays all modern browsers support websockets natively, so in most cases there is no need to have websocket simulation fallbacks like XHR long polling or Flash. Removing these allows Sacrificial-Socket to be lightweight and very performant.

Sacrificial-Socket supports rooms, roomcasts, broadcasts, and event emitting just like Socket.IO, but with one key difference. The data passed into event functions is not an interface{} that is implied to be a string or map[string]interface{}, but is always passed in as a []byte making it easier to unmarshal into your own JSON data structs, convert to a string, or keep as binary data without the need to check the data's type before processing it. It also means there aren't any unnecessary conversions to the data between the client and the server.

Sacrificial-Socket also has a MultihomeBackend interface for syncronizing broadcasts and roomcasts across multiple instances of Sacrificial-Socket running on multiple machines. Out of the box Sacrificial-Socket provides a MultihomeBackend interface for the popular noSQL database MongoDB, one for the moderately popular key/value storage engine Redis, and one for the not so popular GRPC protocol, for syncronizing instances on multiple machines.

In depth examples can be found in the [__examples__ ](https://github.com/raz-varren/sacrificial-socket/tree/master/examples "Examples") directory.

Usage
-----
#### Client Javascript:
```javascript
(function(SS){ 'use strict';
    var ss = new SS('ws://localhost:8080/socket');
    ss.onConnect(function(){
        ss.emit('echo', 'hello echo!');
    });
    
    ss.on('echo', function(data){
        alert('got echo:', data);
        ss.close();
    });
    
    ss.onDisconnect(function(){
        console.log('socket connection closed');
    });
})(window.SS);
```

#### Server Go:
```go
package main

import(
    "net/http"
    ss "github.com/raz-varren/sacrificial-socket"
)

func doEcho(s *ss.Socket, data []byte) {
    s.Emit("echo", string(data))
}

func main() {
    s := ss.NewServer()
    s.On("echo", doEcho)
    
    http.Handle("/socket", s.WebHandler())
    http.ListenAndServe(":8080", nil);
}
```

