package types

// MessagePayload holds data related to a message payload
type MessagePayload struct {
	Topic string `json:"topic" validate:"required"`
	Data  string `json:"data" validate:"required"`
	Key   string `json:"key" validate:"required"`
}
