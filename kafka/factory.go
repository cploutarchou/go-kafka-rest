package kafka

import "github.com/Shopify/sarama"

func TheProducerFactory(brokers []string, config *sarama.Config) (sarama.SyncProducer, sarama.AsyncProducer, error) {

	syncProducer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, nil, err
	}

	asyncProducer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		err := syncProducer.Close()
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, err
	}

	return syncProducer, asyncProducer, nil
}
