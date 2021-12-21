package model


type Config struct {
	KafkaConf `ini:"kafka"`
	EsConf `ini:"es"`
}


type KafkaConf struct {
	Address []string `ini:"address"`
	Topic string `ini:"topic"`
}


type EsConf struct {
	Address string `ini:"address"`
	Index string `ini:"index"`
	MaxSize int `ini:"max_chan_size"`
	GoroutineNum int `ini:"goroutine_num"`
}

type LogMsg struct {

}