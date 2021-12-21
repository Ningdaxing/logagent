package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
	"log_transfer/es"
)



// Init 初始化
// 从kafka中取出数据
func Init(addr []string, topic string) (err error) {

	consumer, err := sarama.NewConsumer(addr, nil)
	if err != nil {
		fmt.Printf("fail to start consumer, err:%v\n", err)
		return
	}

	partitionsList, err := consumer.Partitions(topic)
	if err != nil {
		fmt.Printf("fail to get list of partition:err%v\n", err)
		return
	}

	for partition := range partitionsList {
		pc, err := consumer.ConsumePartition(topic, int32(partition), sarama.OffsetNewest)  // 从最新处消费
		if err != nil{
			logrus.Errorf("consumer faild:err%v\n", err)
			return err
		}
		//defer pc.AsyncClose()

		go func(sarama.PartitionConsumer) {
			for msg := range pc.Messages(){
				//logDataChan <- msg  //流程异步化，取的日志数据先放到golang的chan里面
				fmt.Println(msg.Topic, string(msg.Value))
				var m1 map[string]interface{}
				err = json.Unmarshal(msg.Value, &m1)
				if err != nil{
					logrus.Errorf("json change faild,%s",err)
					continue
				}
				es.PutLogData(m1)
			}
		}(pc)

	}

	return 
}


