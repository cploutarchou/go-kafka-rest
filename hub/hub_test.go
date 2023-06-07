package hub

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

type MockConn struct {
	*bytes.Buffer
}

func (MockConn) WriteMessage(messageType int, data []byte) error {
	return nil
}

func (MockConn) Close() error {
	return nil
}

func (MockConn) ReadMessage() (int, []byte, error) {
	return 0, []byte("mocked message"), nil
}

func TestHub(t *testing.T) {
	h := NewHub(nil, nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h.IsTest(true)
	client := &Client{
		Conn: &MockConn{bytes.NewBuffer(nil)},
		Send: make(chan Message),
	}
	h.RegisterClient(client)

	go h.Run(ctx, log.New(os.Stdout, "HUB: ", log.Ldate|log.Ltime))

	message := Message{
		Data:  "Hello, world!",
		Topic: "test",
		Key:   "test",
	}
	h.BroadcastMessage(message)

	go func() {

		receivedMessage := <-client.Send

		assert.Equal(t, message, receivedMessage, "received message does not match the broadcasted message")
	}()

	h.UnregisterClient(client)
}

func TestClient(t *testing.T) {
	hub_ := NewHub(nil, nil, nil)
	hub_.IsTest(true)
	mockConn := &MockConn{bytes.NewBuffer(nil)}

	if mockConn == nil {
		t.Error("mockConn is nil")
	}

	client := &Client{
		Conn: mockConn,
		Send: make(chan Message),
	}

	go client.ReadPump(hub_, log.New(os.Stdout, "READPUMP: ", log.Ldate|log.Ltime))

	message := Message{
		Data:  "Hello, world!",
		Topic: "test",
		Key:   "test",
	}

	go func() {
		client.Send <- message
	}()

	receivedMessage := <-client.Send
	assert.Equal(t, message.Data, receivedMessage.Data, "client should correctly read the message from the WebSocket connection")

	go client.WritePump(log.New(os.Stdout, "WRITEPUMP: ", log.Ldate|log.Ltime))
	client.Send <- Message{
		Data:  "Hello, world!",
		Topic: "test",
		Key:   "test",
	}
	err := client.CloseSend()
	assert.NoError(t, err, "closing client's send channel should not return error")
}
