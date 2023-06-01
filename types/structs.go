package types

type MessagePayload struct {
	Topic string `json:"topic"`
	Data  string `json:"data"`
	Key   string `json:"key"`
}
