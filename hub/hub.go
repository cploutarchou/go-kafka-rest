package hub

import (
	"context"
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/cploutarchou/go-kafka-rest/kafka"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

var brokers []string         // Kafka brokers
var producer *kafka.Producer // Kafka producer

type Message struct {
	Topic string `json:"topic"`
	Key   string `json:"key"`
	Data  string `json:"data"`
}

type ConnectionInterface interface {
	Close() error
	WriteMessage(int, []byte) error
	ReadMessage() (int, []byte, error)
}

type ClientInterface interface {
	ReadPump(hub TheHub, logger Logger)
	WritePump(logger Logger)
	GetConnection() ConnectionInterface
	SendMessage(msg Message) error
	CloseSend() error
}

type Client struct {
	Conn     ConnectionInterface
	Send     chan Message
	Producer kafka.Producer
}

type WebSocketConnection struct {
	Conn *websocket.Conn
}

func (wsc *WebSocketConnection) Close() error {
	return wsc.Conn.Close()
}

func (wsc *WebSocketConnection) WriteMessage(messageType int, data []byte) error {
	return wsc.Conn.WriteMessage(messageType, data)
}

func (wsc *WebSocketConnection) ReadMessage() (int, []byte, error) {
	return wsc.Conn.ReadMessage()
}

func (c *Client) GetConnection() ConnectionInterface {
	return c.Conn
}

func (c *Client) SendMessage(msg Message) error {
	c.Send <- msg
	return nil
}

func (c *Client) ReadMessage() (int, []byte, error) {
	return c.Conn.ReadMessage()
}

func (c *Client) CloseSend() error {
	close(c.Send)
	return nil
}

type TheHub interface {
	Run(ctx context.Context, logger Logger)
	RegisterClient(client ClientInterface)
	UnregisterClient(client ClientInterface)
	BroadcastMessage(message Message)
	HandleWebSocketMessage(message Message)
	UpgradeWebSocket(c *websocket.Conn, logger Logger)
}

type Logger interface {
	Printf(format string, v ...interface{})
}

type Hub struct {
	Clients      map[ClientInterface]bool
	Broadcast    chan Message
	Register     chan ClientInterface
	Unregister   chan ClientInterface
	mutex        sync.RWMutex
	Logger       Logger
	producer     kafka.Producer
	SaramaConfig *sarama.Config
}

func NewHub(brokers_ []string, logger *log.Logger, config *sarama.Config) TheHub {
	var err error

	h := &Hub{
		Clients:    make(map[ClientInterface]bool),
		Broadcast:  make(chan Message, 100),
		Register:   make(chan ClientInterface, 100),
		Unregister: make(chan ClientInterface, 100),
	}

	if config == nil {
		h.SaramaConfig = sarama.NewConfig()
		producer, err = kafka.NewProducer(brokers_, h.SaramaConfig, kafka.TheProducerFactory)
		if err != nil {
			panic(err)
		}
	} else {
		producer, err = kafka.NewProducer(brokers, h.SaramaConfig, kafka.TheProducerFactory)
		if err != nil {
			panic(err)
		}
	}
	if logger == nil {
		h.Logger = log.New(log.Writer(), "hub: ", log.LstdFlags)
	} else {
		h.Logger = logger
	}
	h.producer = *producer

	log.Print("🚀 Hub initialized")
	return h
}

func (h *Hub) Run(ctx context.Context, logger Logger) {
	for {
		select {
		case <-ctx.Done():
			logger.Printf("Hub shutting down...")
			return
		case client := <-h.Register:
			h.mutex.Lock()
			h.Clients[client] = true
			h.mutex.Unlock()
			logger.Printf("WebSocket connected")
		case client := <-h.Unregister:
			h.mutex.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				err := client.CloseSend()
				if err != nil {
					logger.Printf("Error closing send channel: %v", err)
					return
				}
				logger.Printf("WebSocket disconnected")
			}
			h.mutex.Unlock()
		case message := <-h.Broadcast:
			h.mutex.RLock()
			for client := range h.Clients {
				err := client.SendMessage(message)
				if err != nil {
					logger.Printf("Failed to send message to client")
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func (h *Hub) RegisterClient(client ClientInterface) {
	h.Register <- client
}

func (h *Hub) UnregisterClient(client ClientInterface) {
	h.Unregister <- client
}

func (h *Hub) BroadcastMessage(message Message) {
	h.Broadcast <- message
}

func (h *Hub) HandleWebSocketMessage(message Message) {
	// Send the message to Kafka
	go func() {
		value := message.Data
		_, _, err := h.producer.SendMessageSync(message.Topic, message.Key, value)
		if err != nil {
			h.Logger.Printf("Failed to send message to Kafka: %v", err)
		}
	}()

	// Broadcast the message to other WebSocket clients
	h.BroadcastMessage(message)
}

func (h *Hub) UpgradeWebSocket(c *websocket.Conn, logger Logger) {
	client := &Client{
		Conn: &WebSocketConnection{Conn: c},
		Send: make(chan Message),
	}
	h.RegisterClient(client)

	go func() {
		defer h.UnregisterClient(client)
		client.WritePump(logger)
	}()

	client.ReadPump(h, logger)
}

func (c *Client) WritePump(logger Logger) {
	defer func() {
		err := c.GetConnection().Close()
		if err != nil {
			logger.Printf("Error closing connection: %v", err)
		}
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				err := c.GetConnection().Close()
				if err != nil {
					logger.Printf("Error closing connection: %v", err)
				}
				return
			}

			err := c.GetConnection().WriteMessage(websocket.TextMessage, []byte(message.Data))
			if err != nil {
				logger.Printf("Error writing message: %v", err)
				return
			}
		}
	}
}

func (c *Client) ReadPump(hub TheHub, logger Logger) {
	defer func() {
		hub.UnregisterClient(c)
		err := c.GetConnection().Close()
		if err != nil {
			logger.Printf("Error closing connection: %v", err)
		}
	}()

	for {
		_, data, err := c.GetConnection().ReadMessage()
		if err != nil {
			logger.Printf("Error reading message: %v", err)
			return
		}

		type ReceivedMessage struct {
			Message Message `json:"message"`
		}
		var message Message
		err = json.Unmarshal(data, &message)
		if err != nil {
			logger.Printf("Error unmarshalling message: %v", err)
			return
		}

		hub.HandleWebSocketMessage(message) // Call the HandleWebSocketMessage method of the hub
	}
}
