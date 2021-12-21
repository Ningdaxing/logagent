package es

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

import (
	"context"
	"github.com/olivere/elastic/v7"
)
// Elasticsearch demo

type Person struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Married bool   `json:"married"`
}

type EsClient struct {
	client *elastic.Client
	index string
	logDataChan chan interface{}
}


var (
	//esClient *EsClient  // 直接是个空指针，给空指针赋值会直接报错
	esClient = &EsClient{}  // 初始化一个指针
)

func Init(addr,index string, goRoutineNum int, maxSize int) (err error)  {
	client, err := elastic.NewClient(
		elastic.SetURL(fmt.Sprintf("http://%s", addr)),
		elastic.SetSniff(false),  //  连接docker必须这个配置，不会自动转换地址
	)
	if err != nil {
		// Handle error
		panic(err)
	}

	esClient.client = client
	esClient.index = index
	esClient.logDataChan = make(chan interface{}, maxSize)

	fmt.Println("connect to es success")
	for i:=0; i < goRoutineNum; i++ {
		go sendToEs()
	}
	return
}

func sendToEs() {
	for msg := range esClient.logDataChan{
		//b, err := json.Marshal(msg)
		//if err != nil{
		//	logrus.Errorf("marshal m1 faild, err:%v\n", err)
		//	continue
		//}
		put1, err := esClient.client.Index().
			Index(esClient.index).
			BodyJson(msg).
			Do(context.Background())
		if err != nil {
			// Handle error
			panic(err)
		}
		logrus.Infof("Indexed user %s to index %s, type %s\n", put1.Id, put1.Index, put1.Type)
	}

}

// PutLogData 通过一个首字母大写的函数暴露出去，接收数据
func PutLogData(msg interface{}) {
	esClient.logDataChan <- msg
}





