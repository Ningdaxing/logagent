package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

var (
	Client sarama.SyncProducer
	msgChan chan *sarama.ProducerMessage  // 直接放一个内存地址肯定比字符串省空间
)

func Init(address []string, chanSize int64) (err error){
	// 生成一个生产者

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll  // ack
	config.Producer.Partitioner = sarama.NewRandomPartitioner  // 分区
	config.Producer.Return.Successes = true  // 确认

	Client, err = sarama.NewSyncProducer(address, config)
	if err != nil {
		//logrus.Error("producer closed, err:", err)
		return err
	}
	// 初始化chan
	msgChan = make(chan *sarama.ProducerMessage, chanSize)
	go sendMsg2Kafka()
	return
}

func sendMsg2Kafka() {
	for  {
		select {
		case msg := <- msgChan:
			pid, offset, err := Client.SendMessage(msg)
			if err != nil{
				logrus.Warn("send msg faild, err:", err)
				return
			}
			logrus.Infof("send msg to kafka success, pid:%v, offset:%v\n", pid, offset)
		}
	}

}

func ToMsgChan(msg *sarama.ProducerMessage)  {
	msgChan <- msg
}