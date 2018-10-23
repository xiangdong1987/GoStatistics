package server

import (
	"../tool"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net"
	"strings"
	"sync"
	"time"
)

type moduleStats struct {
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
	mutex               sync.Mutex       //互斥锁
	key                 string           //接口名
	totalCount          int32            //总共次数
	totalTime           int32            //总时间
	maxTime             int32            //最大时间
	minTime             int32            //最小时间
	failCount           int32            //失败次数
	totalFailTime       int32            //失败总时间
	ipServerList        map[string]int32 //访问ip列表
	ipClientList        map[string]int32 //客户端访问列表
	ipFailClientList    map[string]int32 //失败客户端ip列表
	ipFailServerList    map[string]int32 //失败服务端ip列表
	failRetCodeList     map[int32]int32  //失败返回code
	ipSuccessClientList map[string]int32 //成功客户端列表
	ipSuccessServerList map[string]int32 //成功服务端列表
	SuccessRetCodeList  map[int32]int32  //成功放回code
}

type StatsServer struct {
	timeInterval    int
	timeKeyInterval int
	allCount        map[string]moduleCounts
	serverCount     map[string]map[string]moduleCounts
	clientCount     map[string]map[string]moduleCounts
}

func New() (*StatsServer, error) {
	s := &StatsServer{}
	s.timeKeyInterval = 5
	s.timeInterval = 5
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
	for {
		buf := make([]byte, 1024)
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			continue
		}
		go s.unpack(pc, addr, buf[:n])
	}
}

//计算
func (s *StatsServer) calculateModule(content *moduleStats) {
	m := getM()
	key := fmt.Sprintf("%d_%d_%d_%d", content.moduleId, content.interfaceId, int32(math.Floor(float64(m/s.timeInterval))), int32(math.Floor(float64(m/s.timeKeyInterval))))
	clientIp := content.clientIp
	serverIp := tool.Long2ip(content.serverIp)
	//all被调
	if _, ok := s.allCount[key]; ok {
		s.allCount[key] = s.calculateItem(key, s.allCount, serverIp, clientIp, content)
	} else {
		s.allCount = make(map[string]moduleCounts)
		s.allCount[key] = s.calculateItem(key, s.allCount, serverIp, clientIp, content)
	}
	//server被调
	if _, ok := s.serverCount[key]; ok {
		s.serverCount[key][serverIp] = s.calculateItem(serverIp, s.serverCount[key], serverIp, clientIp, content)
	} else {
		s.serverCount = make(map[string]map[string]moduleCounts)
		if _, ok := s.serverCount[key]; ok {
			s.serverCount[key][serverIp] = s.calculateItem(serverIp, s.serverCount[key], serverIp, clientIp, content)
		} else {
			s.serverCount[key] = make(map[string]moduleCounts)
			s.serverCount[key][serverIp] = s.calculateItem(serverIp, s.serverCount[key], serverIp, clientIp, content)
		}
	}
	//client被调
	if _, ok := s.clientCount[key]; ok {
		s.clientCount[key][clientIp] = s.calculateItem(clientIp, s.clientCount[key], serverIp, clientIp, content)
	} else {
		s.clientCount = make(map[string]map[string]moduleCounts)
		if _, ok := s.clientCount[key]; ok {
			s.clientCount[key][clientIp] = s.calculateItem(clientIp, s.clientCount[key], serverIp, clientIp, content)
		} else {
			s.clientCount[key] = make(map[string]moduleCounts)
			s.clientCount[key][clientIp] = s.calculateItem(clientIp, s.clientCount[key], serverIp, clientIp, content)
		}
	}
	fmt.Println(s)
}

//计算单个统计
func (s *StatsServer) calculateItem(key string, item map[string]moduleCounts, serverIp string, clientIp string, content *moduleStats) moduleCounts {
	var tmp moduleCounts
	if _, ok := item[key]; ok {
		//存在
		tmp = item[key]
		//fmt.Println(tmp)
		tmp.mutex.Lock()
		tmp.totalCount += 1
		tmp.totalTime += content.millisecond
		if content.success == 0 {
			tmp.failCount += 1
			tmp.totalFailTime += content.millisecond
		}
		if content.millisecond > tmp.maxTime {
			tmp.maxTime = content.millisecond
		}
		if content.millisecond < tmp.minTime {
			tmp.minTime = content.millisecond
		}
	} else {
		tmp = moduleCounts{}
		item = make(map[string]moduleCounts)
		item[key] = tmp
		tmp.mutex.Lock()
		tmp.totalCount = 1
		tmp.totalTime = 1
		tmp.maxTime = content.millisecond
		tmp.minTime = content.millisecond
		if content.success == 0 {
			tmp.failCount = 1
			tmp.totalFailTime = content.millisecond
		} else {
			tmp.failCount = 0
			tmp.totalFailTime = 0
		}
		tmp.key = key
	}
	if _, ok := tmp.ipServerList[serverIp]; ok {
		tmp.ipServerList[serverIp] += 1
	} else {
		tmp.ipServerList = make(map[string]int32)
		tmp.ipServerList[serverIp] = 1
	}
	if _, ok := tmp.ipClientList[serverIp]; ok {
		tmp.ipClientList[serverIp] += 1
	} else {
		tmp.ipClientList = make(map[string]int32)
		tmp.ipClientList[serverIp] = 1
	}
	if content.success == 0 {
		if _, ok := tmp.ipFailServerList[serverIp]; ok {
			tmp.ipFailServerList[serverIp] += 1
		} else {
			tmp.ipFailServerList = make(map[string]int32)
			tmp.ipFailServerList[serverIp] = 1
		}
		if _, ok := tmp.ipFailClientList[clientIp]; ok {
			tmp.ipFailClientList[clientIp] += 1
		} else {
			tmp.ipFailClientList = make(map[string]int32)
			tmp.ipFailClientList[clientIp] = 1
		}
		if _, ok := tmp.failRetCodeList[content.retCode]; ok {
			tmp.failRetCodeList[content.retCode] += 1
		} else {
			tmp.failRetCodeList = make(map[int32]int32)
			tmp.failRetCodeList[content.retCode] = 1
		}
	} else {
		if _, ok := tmp.ipSuccessServerList[serverIp]; ok {
			tmp.ipSuccessServerList[serverIp] += 1
		} else {
			tmp.ipSuccessServerList = make(map[string]int32)
			tmp.ipSuccessServerList[serverIp] = 1
		}
		if _, ok := tmp.ipSuccessClientList[clientIp]; ok {
			tmp.ipSuccessClientList[clientIp] += 1
		} else {
			tmp.ipSuccessClientList = make(map[string]int32)
			tmp.ipSuccessClientList[clientIp] = 1
		}
		if _, ok := tmp.SuccessRetCodeList[content.retCode]; ok {
			tmp.SuccessRetCodeList[content.retCode] += 1
		} else {
			tmp.SuccessRetCodeList = make(map[int32]int32)
			tmp.SuccessRetCodeList[content.retCode] = 1
		}
	}
	tmp.mutex.Unlock()
	return tmp
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
		var content *moduleStats
		for i < n {
			content = new(moduleStats)
			parsingStats(buf[i*25:(i+1)*25], content, addr)
			//计算
			s.calculateModule(content)
			i++
		}
	} else {
		panic("pack length is invalid")
	}
	//pc.WriteTo(buf, addr)
}

//解析字段
func parsingStats(one []byte, content *moduleStats, addr net.Addr) *moduleStats {
	byteToModule(one[0:4], content, 1)
	byteToModule(one[4:8], content, 2)
	byteToModule(one[8:9], content, 3)
	byteToModule(one[9:13], content, 4)
	byteToModule(one[13:17], content, 5)
	byteToModule(one[17:21], content, 6)
	byteToModule(one[21:25], content, 7)
	content.clientIp = getIp(addr.String())
	return content
}

//获取ip
func getIp(str string) string {
	a := strings.Split(str, ":")
	return a[0]
}

//流转结构
func byteToModule(bt []byte, m *moduleStats, t int) {
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
