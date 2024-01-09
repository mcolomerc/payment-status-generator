// Example function-based Apache Kafka producer
package producer

import (
	"context"
	"encoding/json"
	"fmt"

	model "mcolomerc/synth-payment-producer/pkg/avro"
	"mcolomerc/synth-payment-producer/pkg/config"
	"os"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avro"
	"github.com/m-mizutani/zlog"
	"github.com/m-mizutani/zlog/filter"
)

var logger *zlog.Logger

type Producer struct {
	kafka          *kafka.Producer
	schemaRegistry *schemaregistry.Client
	ser            *avro.SpecificSerializer
	config         config.Config
}

func NewProducer(config config.Config) Producer {
	logger = zlog.New(zlog.WithFilters(filter.Tag()))

	kConfig := &kafka.ConfigMap{
		"bootstrap.servers": config.Kafka.BootstrapServers,
		"client.id":         config.Kafka.ClientId,
		"sasl.mechanisms":   config.Kafka.SaslMechanisms,
		"security.protocol": config.Kafka.SecurityProtocol,
		"sasl.username":     config.Kafka.SaslUsername,
		"sasl.password":     config.Kafka.SaslPassword,
	}
	for k, v := range config.Kafka.ConfigMap {
		kConfig.SetKey(k, v)
	}
	vnum, vstr := kafka.LibraryVersion()
	logger.Info("Library Version: %s (0x%x)", vstr, vnum)
	logger.Info("Link Info:       %s", kafka.LibrdkafkaLinkInfo)

	producer, err := kafka.NewProducer(kConfig)
	if err != nil {
		logger.Info("Failed to create producer: %s", err)
		os.Exit(1)
	}

	client, err := schemaregistry.NewClient(schemaregistry.NewConfigWithAuthentication(
		config.SchemaRegistry.Endpoint,
		config.SchemaRegistry.ApiKey,
		config.SchemaRegistry.ApiSecret))
	if err != nil {
		logger.Info("Failed to create schema registry client: %s\n", err)
		os.Exit(1)
	}
	ser, err := avro.NewSpecificSerializer(client, serde.ValueSerde, avro.NewSerializerConfig())
	if err != nil {
		logger.Info("Failed to create serializer: %s\n", err)
		os.Exit(1)
	}
	// Listen to all the events on the default events channel
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				m := ev
				if m.TopicPartition.Error != nil {
					logger.Info("Delivery failed: %v", m.TopicPartition.Error)
				} else {
					logger.Info("Kafka Producer: Delivered message to topic %s [%d] at offset %v",
						*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
				}
			case kafka.Error:
				fmt.Printf("Error: %v\n", ev)
			case *kafka.Stats:
				// https://github.com/confluentinc/librdkafka/blob/master/STATISTICS.md
				var stats map[string]interface{}
				json.Unmarshal([]byte(e.String()), &stats)
				logger.Info("------librdkafka---------")
				logger.Info("Stats: %v messages (%v bytes) produced",
					stats["txmsgs"], stats["txmsg_bytes"])
				logger.Info(" %v messages ", stats["msg_cnt"])
				logger.Info(" %v number of bytes received from Kafka brokers", stats["rx_bytes"])
				mb := stats["txmsg_bytes"]
				mbb := mb.(float64) / 1024 / 1024
				logger.Info("%v requests sent  (%v bytes / %v Mbytes) bytes transmitted to Kafka brokers\n",
					stats["tx"], stats["tx_bytes"], mbb)

			default:
				logger.Info("Ignored event: %s\n", ev)
			}
		}
	}()
	return Producer{
		kafka:          producer,
		schemaRegistry: &client,
		ser:            ser,
		config:         config,
	}
}

func (p Producer) Produce(payment model.Payment) {
	// Get topic
	topic := fmt.Sprintf("payment-%s", strings.ToLower(payment.Status))
	// Serialize Payment
	payload, err := p.ser.Serialize(topic, &payment)
	if err != nil {
		logger.Info("Failed to serialize payload: %s\n", err)
		os.Exit(1)
	}
	// Produce Payment status update
	err = p.kafka.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(payment.Id),
		Value:          payload,
		Headers:        []kafka.Header{{Key: payment.Id, Value: []byte(payment.Status)}},
	}, nil)
	if err != nil {
		logger.Info("Failed to produce message: %s\n", err)
		os.Exit(1)
	}
	// Flush and close the producer and the events channel
	//for p.kafka.Flush(10000) > 0 {
	//	log.Printf(" Still waiting to flush outstanding messages ")
	//}
	// close(done)
	//p.kafka.Close()
}

func (p Producer) Close() {
	p.kafka.Close()
}

func (p Producer) Flush() {
	for p.kafka.Flush(10000) > 0 {
		logger.Info(" Still waiting to flush outstanding messages ")
	}
}

func (p Producer) CreateTopics() {
	// Create topics
	topics := p.config.Kafka.Topics
	// Create topics
	admin, err := kafka.NewAdminClientFromProducer(p.kafka)
	if err != nil {
		logger.Info("Failed to create admin client: %s\n", err)
		os.Exit(1)
	}
	var topicsSpec []kafka.TopicSpecification
	for k, v := range topics {
		topicsSpec = append(topicsSpec, createTopic(k, v, 3))
	}
	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results, err := admin.CreateTopics(ctx, topicsSpec)
	if err != nil {
		logger.Info("Failed to create topics: %s\n", err)
		os.Exit(1)
	}
	for _, result := range results {
		logger.Info("%s", result)
	}
	return
}

func createTopic(topic string, numParts int, replicationFactor int) kafka.TopicSpecification {
	return kafka.TopicSpecification{
		Topic:             topic,
		NumPartitions:     numParts,
		ReplicationFactor: replicationFactor}
}
