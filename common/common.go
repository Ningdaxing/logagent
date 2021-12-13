package common

// CollectEntry 收集日志的配置的结构体
type CollectEntry struct {
	Path string `json:"path"`
	Topic string `json:"topic"`
}
