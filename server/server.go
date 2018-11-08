package server

import (
	"../tool"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type ModuleStats struct {
	interfaceId int32  //接口ID
	moduleId    int32  //模块ID
	success     int8   //成功或失败
	retCode     int32  //返回码
	serverIp    int32  //服务器端IP
	clientIp    string //客户端IP
	millisecond int32  //调用耗时单位毫秒
	time        int32  //时间单位秒
}

type moduleCounts struct {
	mutex       sync.Mutex //互斥锁
	TotalStatus *Stats
	ServerCount map[string]*Stats //服务端统计
	ClientCount map[string]*Stats //客户端统计
}

type Stats struct {
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

type StatsServer struct {
	pLog            *log.Logger
	TimeInterval    int
	TimeKeyInterval int
	allCount        map[string]*moduleCounts
}

func New() (*StatsServer, error) {
	s := &StatsServer{}
	s.TimeKeyInterval = 5
	s.TimeInterval = 5
	fileName := "xdd.log"
	logFile, err := os.Create(fileName)
	defer logFile.Close()
	if err != nil {
		log.Fatalln("open file error")
	}
	s.pLog = tool.New(logFile, "[Info]", 1)
	s.StartServer()
	return s, nil
}

//开启server
func (s *StatsServer) StartServer() {
	// listen to incoming udp packets
	pc, err := net.ListenPacket("udp", ":9903")
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()
	//添加定时
	interVal := time.Second * time.Duration(s.TimeInterval)
	ticker := time.NewTicker(interVal)
	go func() {
		for {
			select {
			case <-ticker.C:
				//fmt.Printf("ticked at %v\n", time.Now())
				for key, value := range s.allCount {
					go s.dataLand(key, value)
				}
			}
		}
	}()
	for {
		buf := make([]byte, 1024)
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			continue
		}
		go s.unpack(pc, addr, buf[:n])
	}
}

//数据落地
func (s *StatsServer) dataLand(key string, content *moduleCounts) {
	if content.TotalStatus.IpServerList != nil {
		content.mutex.Lock()
		b, err := json.Marshal(*content)
		if err != nil {
			fmt.Println("Umarshal failed:", err)
			return
		}
		s.pLog.Println(string(b))
		//清空数据
		content.TotalStatus = &Stats{}
		content.ServerCount = make(map[string]*Stats)
		content.ClientCount = make(map[string]*Stats)
		content.mutex.Unlock()
	}
}

//计算
func (s *StatsServer) CalculateModule(content *ModuleStats) {
	m := getM()
	key := fmt.Sprintf("%d_%d_%d_%d", content.moduleId, content.interfaceId, int32(math.Floor(float64(m/s.TimeInterval))), int32(math.Floor(float64(m/s.TimeKeyInterval))))
	clientIp := content.clientIp
	serverIp := tool.Long2ip(content.serverIp)
	//all被调
	if _, ok := s.allCount[key]; ok {
		s.allCount[key].TotalStatus = s.calculateItem(key, s.allCount[key].TotalStatus, serverIp, clientIp, content)
	} else {
		s.allCount = make(map[string]*moduleCounts)
		s.allCount[key] = new(moduleCounts)
		s.allCount[key].TotalStatus = s.calculateItem(key, s.allCount[key].TotalStatus, serverIp, clientIp, content)
	}
	//server被调
	if _, ok := s.allCount[key].ServerCount[serverIp]; ok {
		s.allCount[key].ServerCount[serverIp] = s.calculateItem(key, s.allCount[key].ServerCount[serverIp], serverIp, clientIp, content)
	} else {
		s.allCount[key].ServerCount = map[string]*Stats{}
		s.allCount[key].ServerCount[serverIp] = s.calculateItem(key, s.allCount[key].ServerCount[serverIp], serverIp, clientIp, content)
	}
	//client被调
	if _, ok := s.allCount[key].ClientCount[clientIp]; ok {
		s.allCount[key].ClientCount[clientIp] = s.calculateItem(key, s.allCount[key].ClientCount[clientIp], serverIp, clientIp, content)
	} else {
		s.allCount[key].ClientCount = map[string]*Stats{}
		s.allCount[key].ClientCount[clientIp] = s.calculateItem(key, s.allCount[key].ClientCount[clientIp], serverIp, clientIp, content)
	}
	//fmt.Println(fmt.Println(s.allCount[Key]))
}

//计算单个统计
func (s *StatsServer) calculateItem(key string, item *Stats, serverIp string, clientIp string, content *ModuleStats) *Stats {
	s.allCount[key].mutex.Lock()
	if item != nil {
		item.Key = key
		//存在
		item.TotalCount += 1
		item.TotalTime += content.millisecond
		if content.success == 0 {
			item.FailCount += 1
			item.TotalFailTime += content.millisecond
		}
		if content.millisecond > item.MaxTime {
			item.MaxTime = content.millisecond
		}
		if content.millisecond < item.MinTime {
			item.MinTime = content.millisecond
		}
	} else {
		item = &Stats{}
		item.TotalCount = 1
		item.TotalTime = 1
		item.MaxTime = content.millisecond
		item.MinTime = content.millisecond
		if content.success == 0 {
			item.FailCount = 1
			item.TotalFailTime = content.millisecond
		} else {
			item.FailCount = 0
			item.TotalFailTime = 0
		}
		item.Key = key
	}
	if _, ok := item.IpServerList[serverIp]; ok {
		item.IpServerList[serverIp] += 1
	} else {
		item.IpServerList = make(map[string]int32)
		item.IpServerList[serverIp] = 1
	}
	if _, ok := item.IpClientList[serverIp]; ok {
		item.IpClientList[serverIp] += 1
	} else {
		item.IpClientList = make(map[string]int32)
		item.IpClientList[serverIp] = 1
	}
	if content.success == 0 {
		if _, ok := item.IpFailServerList[serverIp]; ok {
			item.IpFailServerList[serverIp] += 1
		} else {
			item.IpFailServerList = make(map[string]int32)
			item.IpFailServerList[serverIp] = 1
		}
		if _, ok := item.IpFailClientList[clientIp]; ok {
			item.IpFailClientList[clientIp] += 1
		} else {
			item.IpFailClientList = make(map[string]int32)
			item.IpFailClientList[clientIp] = 1
		}
		if _, ok := item.FailRetCodeList[content.retCode]; ok {
			item.FailRetCodeList[content.retCode] += 1
		} else {
			item.FailRetCodeList = make(map[int32]int32)
			item.FailRetCodeList[content.retCode] = 1
		}
	} else {
		if _, ok := item.IpSuccessServerList[serverIp]; ok {
			item.IpSuccessServerList[serverIp] += 1
		} else {
			item.IpSuccessServerList = make(map[string]int32)
			item.IpSuccessServerList[serverIp] = 1
		}
		if _, ok := item.IpSuccessClientList[clientIp]; ok {
			item.IpSuccessClientList[clientIp] += 1
		} else {
			item.IpSuccessClientList = make(map[string]int32)
			item.IpSuccessClientList[clientIp] = 1
		}
		if _, ok := item.SuccessRetCodeList[content.retCode]; ok {
			item.SuccessRetCodeList[content.retCode] += 1
		} else {
			item.SuccessRetCodeList = make(map[int32]int32)
			item.SuccessRetCodeList[content.retCode] = 1
		}
	}
	s.allCount[key].mutex.Unlock()
	//fmt.Println(item)
	return item
}

//获取分钟
func getM() int {
	h := time.Now().Hour()
	i := time.Now().Minute()
	return h*60 + i
}

//解包
func (s *StatsServer) unpack(pc net.PacketConn, addr net.Addr, buf []byte) {
	length := len(buf)
	if math.Mod(float64(length), 25) == 0 {
		i := 0
		n := length / 25
		var content *ModuleStats
		for i < n {
			content = new(ModuleStats)
			s.ParsingStats(buf[i*25:(i+1)*25], content, addr)
			//计算
			s.CalculateModule(content)
			i++
		}
	} else {
		panic("pack length is invalid")
	}
	//压测试使用
	//buffer := &bytes.Buffer{}
	//backToStringSlice := []string{"ok"}
	//gob.NewEncoder(buffer).Encode(backToStringSlice)
	//byteSlice := buffer.Bytes()
	//pc.WriteTo(byteSlice, addr)
}

//解析字段
func (s *StatsServer) ParsingStats(one []byte, content *ModuleStats, addr net.Addr) *ModuleStats {
	ByteToModule(one[0:4], content, 1)
	ByteToModule(one[4:8], content, 2)
	ByteToModule(one[8:9], content, 3)
	ByteToModule(one[9:13], content, 4)
	ByteToModule(one[13:17], content, 5)
	ByteToModule(one[17:21], content, 6)
	ByteToModule(one[21:25], content, 7)
	content.clientIp = getIp(addr.String())
	return content
}

//获取ip
func getIp(str string) string {
	a := strings.Split(str, ":")
	return a[0]
}

//流转结构
func ByteToModule(bt []byte, m *ModuleStats, t int) {
	data := bytes.NewReader(bt)
	var err error
	switch t {
	case 1:
		err = binary.Read(data, binary.BigEndian, &m.interfaceId)
	case 2:
		err = binary.Read(data, binary.BigEndian, &m.moduleId)
	case 3:
		err = binary.Read(data, binary.BigEndian, &m.success)
	case 4:
		err = binary.Read(data, binary.BigEndian, &m.retCode)
	case 5:
		err = binary.Read(data, binary.BigEndian, &m.serverIp)
	case 6:
		err = binary.Read(data, binary.BigEndian, &m.millisecond)
	case 7:
		err = binary.Read(data, binary.BigEndian, &m.time)
	}
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
}
