package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"reflect"
	"sync"
	"testing"
)

func TestSingletonProducer(t *testing.T) {
	// Given
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true
	config.Producer.Compression = sarama.CompressionSnappy

	factory := func(brokers []string, config *sarama.Config) (sarama.SyncProducer, sarama.AsyncProducer, error) {
		return mocks.NewSyncProducer(t, nil), mocks.NewAsyncProducer(t, nil), nil
	}

	// When
	producer1, err := NewProducer(nil, config, factory)
	if err != nil {
		t.Fatalf("Failed to create producer: %v", err)
	}
	producer2, err := NewProducer(nil, config, factory)
	if err != nil {
		t.Fatalf("Failed to create producer: %v", err)
	}

	// Then
	if !reflect.DeepEqual(producer1, producer2) {
		t.Errorf("NewProducer does not return the same instance")
	}
}

func TestSendMessageSync(t *testing.T) {
	// Given
	mockSyncProducer := mocks.NewSyncProducer(t, nil)
	mockSyncProducer.ExpectSendMessageAndSucceed()

	producer := &Producer{
		syncProducer: mockSyncProducer,
		mutex:        &sync.Mutex{},
	}

	// When
	topic := "test-topic"
	key := "test-key"
	value := "test-value"

	partition, offset, err := producer.SendMessageSync(topic, key, value)

	// Then
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if partition != 0 {
		t.Errorf("Expected partition 0, got %d", partition)
	}
	if offset != 1 { // Update the expected offset value to 1
		t.Errorf("Expected offset 1, got %d", offset)
	}
}
