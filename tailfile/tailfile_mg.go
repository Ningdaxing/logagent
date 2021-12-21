package tailfile

import (
	"github.com/sirupsen/logrus"
	"logagent/common"
	"sync"
)

// tailTask 的管理者

type tailTaskMg struct {
	tailTaskMap      map[string]*tailTask       // 路径和对应性的map
	collectEntryList []common.CollectEntry      // 所有配置项
	confChan         chan []common.CollectEntry // 等待新配置的通道
}

var (
	ttMg *tailTaskMg
)

func (t *tailTaskMg) watch() {
	for {
		newConf := <-t.confChan
		logrus.Infof("get new conf from etcd, conf:%v", newConf)
		for _, conf := range newConf {
			// 1、原来已经存在的的任务就不用动
			if t.isExist(conf) {
				continue
			}
			// 2、原来没有的需要新创建一个
			tt := newTailTask(conf.Path, conf.Topic)
			err := tt.Init()
			if err != nil {
				logrus.Errorf("create tailObj for path: %s, topic %s is faild", tt.path, tt.topic)
				continue
			}
			logrus.Infof("create a tail task for path:%s", tt.path)
			ttMg.tailTaskMap[tt.path] = tt // 把创建的这个tailTask任务登记在册，方便后续管理
			go tt.run()
		}
		// 3、原来有的任务新传进来的没有了，就得把之前的tailfile停掉
		// 找到tailTaskMap中存在的，但是newconf里面不存在了
		for key, task := range t.tailTaskMap {
			var found bool
			for _, conf := range newConf {
				if key == conf.Path {
					found = true
					break
				}
			}
			if !found {
				// 如果map中的task没有在newconf中，这里处理 停掉的函数
				delete(t.tailTaskMap, key) // 删掉这个key在map中
				task.cancel()
			}
		}
	}
}

func (t *tailTaskMg) isExist(conf common.CollectEntry) bool {
	_, ok := t.tailTaskMap[conf.Path]
	return ok
}

func SendNewConf(newConf []common.CollectEntry) {
	ttMg.confChan <- newConf
}

func Init(allConf []common.CollectEntry) (err error) {
	// allConf 是多个日志的收集项
	var wg sync.WaitGroup

	ttMg = &tailTaskMg{
		tailTaskMap:      make(map[string]*tailTask, 20),
		collectEntryList: allConf,
		confChan:         make(chan []common.CollectEntry),
	}

	for _, conf := range allConf {
		tt := newTailTask(conf.Path, conf.Topic)
		err = tt.Init() // 创建一个日志收集任务
		if err != nil {
			logrus.Errorf("create tailObj for path: %s, topic %s is faild", tt.path, tt.topic)
			continue
		}
		logrus.Infof("create a tail task for path:%s", tt.path)
		ttMg.tailTaskMap[tt.path] = tt // 把创建的这个tailTask任务登记在册，方便后续管理
		// 直接去收集日志
		wg.Add(1)
		go tt.run()
	}
	wg.Add(1)
	go ttMg.watch() // 在后台等新的配置来
	wg.Wait()
	return nil
}
