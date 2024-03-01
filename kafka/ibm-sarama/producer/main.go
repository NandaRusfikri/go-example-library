package main

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/IBM/sarama"
	log "github.com/sirupsen/logrus"

	"time"
)

func main() {

	topic := "coba"
	CreateTopic(topic)
	//CreateTopic("ayam")

	kafkaProducer := NewKafkaProducer()

	defer func() {
		if err := kafkaProducer.Producer.Close(); err != nil {
			log.Errorf("Unable to stop kafka producer: %v", err)
			return
		}
	}()

	productID := rand.Int()

	ayam := kafkaProducer.SendMessage(topic, map[string]interface{}{
		"product_id": productID,
	}, 1)
	if ayam != nil {
		fmt.Println(ayam.Error())
	}
	fmt.Println("success ", productID)

}

func NewKafkaProducer() *Producer {

	address, config := getKafkaConfig()

	producers, err := sarama.NewSyncProducer(address, config)

	if err != nil {
		log.Errorf("Unable to create kafka producer got error %v", err)
		//return
	}

	kafka := &Producer{
		Producer: producers,
	}

	return kafka
}

func CreateTopic(topic string) error {
	address, config := getKafkaConfig()
	admin, err := sarama.NewClusterAdmin(address, config)
	if err != nil {
		log.Fatal("Error while creating cluster admin: ", err.Error())
		return err
	}
	defer func() { _ = admin.Close() }()
	//err = admin.CreateTopic(topic, &sarama.TopicDetail{
	//	NumPartitions:     1,
	//	ReplicationFactor: 1,
	//}, false)
	//if err != nil {
	//	log.Errorf("Error while creating topic: %s", err.Error())
	//	return err
	//}

	err = admin.CreatePartitions(topic, 3, [][]int32{}, false)
	if err != nil {
		log.Errorf("Error while creating partitions for topic: %s", err.Error())
		return err
	}
	return err
}

var (
	//brokers  = ""
	version  = ""
	group    = "ServiceProduct"
	assignor = ""
	oldest   = true
)

func getKafkaConfig() ([]string, *sarama.Config) {

	//version, err := sarama.ParseKafkaVersion(version)
	//if err != nil {
	//	log.Panicf("Error parsing Kafka version: %v", err)
	//}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Net.WriteTimeout = 5 * time.Second
	config.Producer.Retry.Max = 0
	//config.Version = version

	if oldest {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}

	return []string{"localhost:9092"}, config
}

type Producer struct {
	Producer sarama.SyncProducer
}

func (p *Producer) SendMessage(topic string, msg map[string]interface{}, partition int32) error {

	jsonByte, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	kafkaMsg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(jsonByte),
		//Partition: partition,
	}

	partition, offset, err := p.Producer.SendMessage(kafkaMsg)
	if err != nil {
		log.Errorf("Send message error: %v", err)
		return err
	}

	log.Infof("Send message success, Topic %v, Partition %v, Offset %d", topic, partition, offset)
	return nil
}
