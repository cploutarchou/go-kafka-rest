package kafka

import (
	"reflect"
	"sync"
	"testing"
	"time"

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

	factory := func(brokers []string, config *sarama.Config) (sarama.SyncProducer, sarama.AsyncProducer, error) {
		return mocks.NewSyncProducer(t, nil), mocks.NewAsyncProducer(t, nil), nil
	}

	// When
	producer1, _ := NewProducer(nil, config, factory)
	producer2, _ := NewProducer(nil, config, factory)

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

	_, _, err := producer.SendMessageSync(topic, key, value)

	// Then
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestSendMessageAsync(t *testing.T) {
	// Given
	mockAsyncProducer := mocks.NewAsyncProducer(t, nil)
	mockAsyncProducer.ExpectInputAndSucceed()

	producer := &Producer{
		asyncProducer: mockAsyncProducer,
		mutex:         &sync.Mutex{},
	}

	// When
	topic := "test-topic"
	key := "test-key"
	value := "test-value"

	producer.SendMessageAsync(topic, key, value)

	// Then
	select {
	case err := <-mockAsyncProducer.Errors():
		t.Errorf("did not expect producer error: %s", err)
	case <-time.After(time.Second):
		// no error occurred, test passed
	}
}
