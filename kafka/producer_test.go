package kafka

import (
	"reflect"
	"sync"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
)

func TestSingletonProducer(t *testing.T) {
	// Given
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true
	config.Producer.Compression = sarama.CompressionSnappy

	factory := func(brokers []string, config *sarama.Config) (sarama.SyncProducer, error) {
		return mocks.NewSyncProducer(t, nil), nil
	}

	// When
	producer1, _ := NewProducer(nil, config, factory)
	producer2, _ := NewProducer(nil, config, factory)

	// Then
	if !reflect.DeepEqual(producer1, producer2) {
		t.Errorf("NewProducer does not return the same instance")
	}
}

func TestSendMessage(t *testing.T) {
	// Given
	mockProducer := mocks.NewSyncProducer(t, nil)
	mockProducer.ExpectSendMessageAndSucceed()

	producer := &Producer{
		producer: mockProducer,
		mutex:    &sync.Mutex{},
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
