package wsock

import (
	"bytes"
	"code.google.com/p/go.net/websocket"
	"io"
	"net/http"
	"strconv"
)

const HeaderLength = 8

/*
Protocol:

The first eight bytes of any message are the length of the message body (not including
the eight bytes themselves), hex-encoded. The message body can be anything at all.
Examples:
  A message consisting of the word "hello":
	00000005hello

  An empty message
    00000000

  A JSON-encoded struct
    0000001f{"Head":"Hello","Body":"World"}

A general SocketHandler is determined by two functions, one of which handles
the connection itself, the other handling individual messages.

The connection handler can be used for setup and teardown. It should call SocketInstance
loop. If there is no particular setup/teardown to be done, the default can be used.

The message handler gets the length of the message and a reader (as well as the socket instance).
It does not have to read the entire message if it doesn't want to; the system will take care
of flushing.
*/

//
// ConnectionHandler is the type of functions that handle connections.
//
type ConnectionHandler interface {
	// Handle should run the main loop and return whatever error terminated it.
	// Returning either nil or io.EOF means that the connection was closed.
	Handle(inst *SocketInstance) error
}

//
// ConnectionHandlerFunc is an instantiation for simple functions.
//
type ConnectionHandlerFunc func(inst *SocketInstance) error

func (f ConnectionHandlerFunc) Handle(inst *SocketInstance) error {
	return f(inst)
}

//
// MessageHandler is the type of functions that handle individual messages.
//
type MessageHandler interface {
	Handle(inst *SocketInstance, msgLength int64, reader io.Reader) error
}

//
// MessageHandlerFunc is an instantiation for simple functions.
//
type MessageHandlerFunc func(inst *SocketInstance, msgLength int64, reader io.Reader) error

func (f MessageHandlerFunc) Handle(inst *SocketInstance, msgLength int64, reader io.Reader) error {
	return f(inst, msgLength, reader)
}

//
// SocketHandler stores message handling information, and creates
// SocketInstances to manage connections. There is one per route.
//
type SocketHandler struct {
	ConnectionHandler ConnectionHandler
	MessageHandler    MessageHandler
}

func simpleConnectionHandlerFunc(inst *SocketInstance) error {
	return inst.ReadLoop()
}

func SocketHandlerMessageFunc(mhf MessageHandlerFunc) *SocketHandler {
	return &SocketHandler{
		ConnectionHandler : ConnectionHandlerFunc(simpleConnectionHandlerFunc),
		MessageHandler : (mhf),
	}
}

func SocketHandlerFuncs(cf ConnectionHandlerFunc, mhf MessageHandlerFunc) *SocketHandler {
	return &SocketHandler{
		ConnectionHandler : cf,
		MessageHandler : mhf,
	}
}

//
// SocketInstance is a wrapper around websocket connections. It
// does the actual message handling
//
type SocketInstance struct {
	Parent *SocketHandler
	Conn   *websocket.Conn
}

func (inst *SocketInstance) Write(data []byte) (int, error) {
	return inst.Conn.Write(data)
}

func (inst *SocketInstance) ReadLoop() error {
	header := make([]byte, HeaderLength)
	for {
		// Read the header
		_, err := io.ReadFull(inst.Conn, header)
		if err == io.EOF {
			// We're done!
			return err
		}
		if err != nil {
			// Handle error
		}
		length, err := strconv.ParseInt(string(header), 16, 32)
		if err != nil {
			// Handle error
		}
		body := make([]byte, length)
		_, err = io.ReadFull(inst.Conn, body)
		if err != nil {
		}
		err = inst.Parent.MessageHandler.Handle(inst, length, bytes.NewReader(body))
	}
	return nil
}

func handlerFunc(h *SocketHandler, ws *websocket.Conn) {
	inst := &SocketInstance{Parent: h, Conn: ws}
	// TODO: write loop?
	h.ConnectionHandler.Handle(inst)
}

func (h *SocketHandler) Handler() websocket.Handler {
	return func(ws *websocket.Conn) {
		handlerFunc(h, ws)
	}
}

func (h *SocketHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.Handler().ServeHTTP(w, req)
}
