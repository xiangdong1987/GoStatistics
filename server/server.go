package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net"
)

/**
 *typedef struct
{
int32_t interface_id; //接口ID
int32_t module_id; //模块ID
int8_t success; //成功或失败
int32_t ret_code; //返回码
int32_t server_ip; //服务器端IP
int32_t millisecond; //调用耗时单位毫秒
int32_t time; //时间单位秒
} module_stats;
*/
type module_stats struct {
	interface_id int32
	module_id    int32
	success      int8
	ret_code     int32
	server_ip    int32
	millisecond  int32
	time         int32
}

//开启server
func StartServer() {
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
		go unpack(pc, addr, buf[:n])
	}
}

func unpack(pc net.PacketConn, addr net.Addr, buf []byte) {
	length := len(buf)
	if math.Mod(float64(length), 25) == 0 {
		i := 0
		n := length / 25
		var content *module_stats
		for i < n {
			content = new(module_stats)
			parsingStats(buf[i*25:(i+1)*25], content)
			fmt.Println(content)
			i++
		}
	} else {
		fmt.Println("xdd")
		panic("pack length is invalid")
	}
	//pc.WriteTo(buf, addr)
}

func parsingStats(one []byte, content *module_stats) *module_stats {
	fmt.Println(one)
	betyToInt(one[0:4], content, 1)
	betyToInt(one[4:8], content, 2)
	betyToInt(one[8:9], content, 3)
	betyToInt(one[9:13], content, 4)
	betyToInt(one[13:17], content, 5)
	betyToInt(one[17:21], content, 6)
	betyToInt(one[21:25], content, 7)
	return content
}

func betyToInt(bt []byte, m *module_stats, t int) {
	data := bytes.NewReader(bt)
	var err error
	switch t {
	case 1:
		err = binary.Read(data, binary.BigEndian, &m.interface_id)
	case 2:
		err = binary.Read(data, binary.BigEndian, &m.module_id)
	case 3:
		err = binary.Read(data, binary.BigEndian, &m.success)
	case 4:
		err = binary.Read(data, binary.BigEndian, &m.ret_code)
	case 5:
		err = binary.Read(data, binary.BigEndian, &m.server_ip)
	case 6:
		err = binary.Read(data, binary.BigEndian, &m.millisecond)
	case 7:
		err = binary.Read(data, binary.BigEndian, &m.time)
	}
	fmt.Println(m)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
}
