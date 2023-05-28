package kafka

import (
	"testing"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
)

func TestSendMessage(t *testing.T) {
	// Given
	mockProducer := mocks.NewSyncProducer(t, nil)
	mockProducer.ExpectSendMessageWithCheckerFunctionAndSucceed(func(val []byte) error {
		if val == nil {
			return sarama.ErrInvalidMessage
		}
		return nil
	})

	producer := &Producer{
		producer: mockProducer,
	}

	// When
	topic := "test-topic"
	key := "test-key"
	value := "test-value"

	_, _, err := producer.SendMessage(topic, key, value)

	// Then
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}
