package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"logagent/common"
	"logagent/tailfile"
	"time"
)

var (
	client *clientv3.Client
)

func Init(address []string) (err error) {
	client, err = clientv3.New(clientv3.Config{
		Endpoints:   address,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logrus.Error("connect etcd faid, err:", err)
		return
	}
	//defer client.Close()   共享client变量的时候，不能关闭
	return nil
}

// GetConf 拉取日志收集配置项
func GetConf(key string) (collectEntryList []common.CollectEntry, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.Get(ctx, key)
	if err != nil {
		logrus.Errorf("faild to get key:%s, err:%v", key, err)
		return
	}
	if len(resp.Kvs) == 0 {
		logrus.Warnf("get len:0,faild to get key:%s, err:%v\n", key, err)
		return
	}
	ret := resp.Kvs[0]
	err = json.Unmarshal(ret.Value, &collectEntryList)
	if err != nil {
		logrus.Errorf("json unmarshal faild, err:%v\n", err)
		return
	}
	return

}

// WatchConf 监控etcd中日志收集配置变化的函数
func WatchConf(key string) {
	watchCh := client.Watch(context.Background(), key)
	for resp := range watchCh {
		for _, evt := range resp.Events {
			fmt.Printf("type: %s, key: %s, values: %s\n", evt.Type, evt.Kv.Key, evt.Kv.Value)
			var newConf []common.CollectEntry
			if evt.Type == clientv3.EventTypeDelete {
				logrus.Warning("warning: etcd will delete this key!!!")
				tailfile.SendNewConf(newConf)
				continue
			}
			err := json.Unmarshal(evt.Kv.Value, &newConf)
			if err != nil {
				logrus.Errorf("json unmarshl new conf faild, err:%v", err)
				continue
			}
			// 告诉tailfile模块应该启动新的配置了
			tailfile.SendNewConf(newConf) // 没有任何接收就是阻塞的
		}
	}
}
