package service

import (
	"fmt"
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/spf13/cast"
	"log"
	"strings"
	"sync"
	//"github.com/spf13/cast"
)

var (
	wg sync.WaitGroup

	kafkaConsumer *cluster.Consumer
	kafkaBrokers  = []string{_masterKafkaURL, _slaverKafkaURL}
)

const (
	_masterKafkaURL = "121.5.252.185:9092"
	_slaverKafkaURL = "1.116.69.194:9092"

	_topicName = "tracking"
)

type BrokerConsumer struct {
}

func (BrokerConsumer) InsertDataToCH() {

	consumer, err := sarama.NewConsumer([]string{_slaverKafkaURL}, nil)

	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	fmt.Println(consumer.Topics())
	partitionList, err := consumer.Partitions(_topicName)

	if err != nil {
		log.Fatal(err)
	}

	//log.Println(partitionList)
	for partition := range partitionList {

		pc, err := consumer.ConsumePartition(_topicName, int32(partition), sarama.OffsetNewest)

		if err != nil {
			log.Fatal(err)
		}

		defer pc.AsyncClose()

		wg.Add(1)

		go func(sarama.PartitionConsumer) {
			defer wg.Done()
			for msg := range pc.Messages() {
				//dspo,2860121,286012100,DEVICE_1E3AD084C623343765ACD4DF92A2DD93,113.65.130.153,1618934446

				//fmt.Printf("Partition:%d, Offset:%d, Key:%s, Value:%s\n", msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))

				values := strings.Split(string(msg.Value), ",")

				insertDate, _ := cast.StringToDate(string(msg.Key))
				trackingType := values[0]
				stragetyID := cast.ToInt32(values[1])
				creativeId := cast.ToInt32(values[2])
				deviceId := values[3]
				ipAddr := values[4]
				timestamp := cast.ToInt32(values[5])
				InsertIntoCHExec(insertDate, trackingType, deviceId, ipAddr, stragetyID, creativeId, timestamp)
			}
		}(pc)

		wg.Wait()
	}

}
