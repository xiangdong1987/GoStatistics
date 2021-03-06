# GoStatistics统计系统
## 实现方式
利用udp服务器实现高并发数据统计（针对数据要求并不严格的应用，udp 不保证数据包不丢失，可以扩展为Tcp服务保证数据不丢失）
## 数据结构
统计相关的结构
```
type stats struct {
	Key                 string           //接口名
	TotalCount          int32            //总共次数
	TotalTime           int32            //总时间
	MaxTime             int32            //最大时间
	MinTime             int32            //最小时间
	FailCount           int32            //失败次数
	TotalFailTime       int32            //失败总时间
	IpServerList        map[string]int32 //访问ip列表
	IpClientList        map[string]int32 //客户端访问列表
	IpFailClientList    map[string]int32 //失败客户端ip列表
	IpFailServerList    map[string]int32 //失败服务端ip列表
	FailRetCodeList     map[int32]int32  //失败返回code
	IpSuccessClientList map[string]int32 //成功客户端列表
	IpSuccessServerList map[string]int32 //成功服务端列表
	SuccessRetCodeList  map[int32]int32
}
```
## 落地方式
* 提供落地方法，自实现落地方式
    1. 日志模式
    2. 数据库模式
## 代码测试
* 基准测试：测试代码性能测试代码在test目录下
    * 测试方法使用go test -bench=. -run=none (不带内存操作消耗)
    * go test -bench=. -benchmem -run=none (带内存操作消耗)
* 测试结果：
    * 虚拟机1核 2G内存
    * 每秒1000000 次
    * 每次操作1720ns
    * 每次内存空间占用432b
    * 每次分配内存次数 20
* swoole upd 压测结果
    * php run.php -c 100 -n 10000 -s udp://127.0.0.1:9903 -f udp
![image](https://github.com/xiangdong1987/GoStatistics/blob/master/images/swoole_benchmark.png)
