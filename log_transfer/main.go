package main

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
	"log_transfer/es"
	"log_transfer/kafka"
	"log_transfer/model"
)

// 从kafka消费日志读取，写入es

func main()  {
	//1、loadConfig
	var cfg = new(model.Config)
	err := ini.MapTo(cfg, "./config/logtransfer.ini")
	if err != nil {
		logrus.Errorf("get local ip faild, err:%v", err)
		return
	}
	logrus.Info("load config success!!!")

	// 2、连接kafka init
	err = kafka.Init(cfg.KafkaConf.Address, cfg.KafkaConf.Topic)
	if err != nil {
		logrus.Error("kafka start faild, err:", err)
		return
	}
	logrus.Info("kafka init success")

	// 3、连接es   init
	err = es.Init(cfg.EsConf.Address, cfg.EsConf.Index, cfg.EsConf.GoroutineNum, cfg.EsConf.MaxSize)
	if err != nil {
		logrus.Error("es start faild, err:", err)
		return
	}
	logrus.Info("es init success")


	// select 不会占cpu，写for会占cpu
	select {}
}
