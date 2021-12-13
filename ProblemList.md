v1.0版本问题：
- 当前只能读取一个日志文件，不支持多个日志文件采集
- 无法管理日志的topic

思路：
- 用etcd存储要收集的日志项，使用json格式
```
[
    {
        "path": "path1/kafka1.log",
        "topic": "kafka1",
    },
    {
        "path": "path1/web.log",
        "topic": "web_log"
    },
]
```
