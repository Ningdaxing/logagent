package tailfile

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/hpcloud/tail"
	"github.com/sirupsen/logrus"
	"logagent/kafka"
	"strings"
	"time"
)

type tailTask struct {
	path   string
	topic  string
	tObj   *tail.Tail
	ctx    context.Context
	cancel context.CancelFunc
}

func newTailTask(path, topic string) *tailTask {
	ctx, cancel := context.WithCancel(context.Background())
	return &tailTask{
		path:   path,
		topic:  topic,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (t *tailTask) Init() (err error) {
	cfg := tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
		MustExist: false,
		Poll:      true,
	}
	t.tObj, err = tail.TailFile(t.path, cfg)
	return
}

func (t *tailTask) run() {
	// 读取日志，发向kafka
	// TailObj --> log --> Client --> kafka
	for {
		select {
		case <-t.ctx.Done(): // 触发主动cancel的话，在这里可以接收到。
			logrus.Infof("the task collect path:%s is stopping....", t.path)
			return
		case line, ok := <-t.tObj.Lines:
			// 循环读数据
			for !ok {
				logrus.Warnf("tail file close reopen, filename:%s\n", t.path)
				time.Sleep(time.Second) // 读取出错
				continue
			}
			// 如果是空行的话，直接跳过
			if len(strings.Trim(line.Text, "\r")) == 0 {
				continue
			}
			// 将消息发到 chan里面，新起一个gorontine 去获取   异步操作
			logrus.Info(line.Text)
			msg := &sarama.ProducerMessage{}
			msg.Topic = t.topic
			msg.Value = sarama.StringEncoder(line.Text)
			kafka.ToMsgChan(msg) // 不直接暴露变量，暴露一个函数
		}

	}
}
