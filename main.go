package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
	"logagent/common"
	"logagent/etcd"
	"logagent/kafka"
	"logagent/tailfile"
)

// 日志收集的客户端
// filebeat

//目标：收集目录下的日志文件，发送到kafka中

// kafka 写入和消费
// tailf 读日志文件

//ini 用法

//cfg, err := ini.Load("./conf/config.ini")
//if err != nil {
//	fmt.Printf("Fail to read file: %v", err)
//	os.Exit(1)
//}
//
//kafkaAddress := cfg.Section("kafka").Key("address").String()
//logrus.Info("kafka:", kafkaAddress)

// Config 整个logagent的配置

type Config struct {
	KafkaConfig   `ini:"kafka"`
	CollectConfig `ini:"collect"`
	EtcdConfig    `ini:"etcd"`
}

type KafkaConfig struct {
	Address  []string `ini:"address"`
	Topic    string   `ini:"topic"`
	ChanSize int64    `ini:"chan_size"`
}

type CollectConfig struct {
	LogFilePath string `ini:"logfile_path"`
}

type EtcdConfig struct {
	Address    []string `ini:"address"`
	CollectKey string   `ini:"collect_key"`
}

//// 真正的业务逻辑
//func run() (err error) {
//	// TailObj --> log --> Client --> kafka
//	for {
//		// 循环读数据
//		line, ok := <- tailfile.Tails.Lines
//		for !ok {
//			logrus.Warnf("tail file close reopen, filename:%s\n", tailfile.Tails.Filename)
//			time.Sleep(time.Second) // 读取出错
//			continue
//		}
//		// 如果是空行的话，直接跳过
//		if len(strings.Trim(line.Text, "\r")) == 0 {
//			continue
//		}
//		// 将消息发到 chan里面，新起一个gorontine 去获取   异步操作
//		logrus.Info(line.Text)
//		msg := &sarama.ProducerMessage{}
//		msg.Topic = "kafkademo"
//		msg.Value = sarama.StringEncoder(line.Text)
//		kafka.ToMsgChan(msg)  // 不直接暴露变量，暴露一个函数
//	}
//}

func main() {
	// 0、初始化(做好准备工作) `go-ini`
	// 1、读配置文件  读取etcd配置
	// 2、跟进文件中的配置传入tailf去收集日志 ==> fix: 从etcd中拉取日志配置项
	// 3、将日志写入kafka (通过sarama)
	////////////////////////

	// -1 获取本机ip，插入ip去获取不同机器的监控项
	ip, err := common.GetOutBoundIp()
	if err != nil {
		logrus.Errorf("get local ip faild, err:%v", err)
		return
	}
	var configObj = new(Config)
	err = ini.MapTo(configObj, "./conf/config.ini")

	if err != nil {
		logrus.Error("load config faild, err:%v", err)
		return
	}
	logrus.Info(configObj)
	err = kafka.Init(configObj.KafkaConfig.Address, configObj.KafkaConfig.ChanSize)
	if err != nil {
		logrus.Error("kafka start faild, err:", err)
		return
	}
	logrus.Info("kafka init success")

	// 初始化ectd，从ectd中拉取要收集的配置项
	err = etcd.Init(configObj.EtcdConfig.Address)
	if err != nil {
		logrus.Error("init etcd faild, err: %v\n", err)
		return
	}
	collectKey := fmt.Sprintf(configObj.EtcdConfig.CollectKey, ip) // 组合etcd key用主机ip去填补
	allConf, err := etcd.GetConf(collectKey)
	if err != nil {
		logrus.Error("get ectd conf faild, err: %v\n", err)
		return
	}
	logrus.Info("etcd init success!!!")
	logrus.Info(allConf)
	// 这里要监控etcd中 configObj.EtcdConfig.CollectKey 这个key对应值的变化
	go etcd.WatchConf(configObj.EtcdConfig.CollectKey)
	err = tailfile.Init(allConf)
	if err != nil {
		logrus.Error("init tail faild: err:", err)
	}

	logrus.Info("tail init success")
}
