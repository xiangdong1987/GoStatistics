package server

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type RpcHeader struct {
	length     int32 //包体长度
	encodeType int32 //编码形式
	uid        int32 //uid
	serid      int32 //请求序列
}

const buffer_maxlen = 10240  //最大待处理区排队长度, 超过后将丢弃最早入队数据
const buffer_clear_num = 128 //超过最大长度后，清理100个数据

const ERR_HEADER = 9001      //错误的包头
const ERR_TOOBIG = 9002      //请求包体长度超过允许的范围
const ERR_SERVER_BUSY = 9003 //服务器繁忙，超过处理能力
const ERR_UNPACK = 9204      //解包失败
const ERR_PARAMS = 9205      //参数错误
const ERR_NOFUNC = 9206      //函数不存在
const ERR_CALL = 9207        //执行错误
const ERR_ACCESS_DENY = 9208 //访问被拒绝，客户端主机未被授权
const ERR_USER = 9209        //用户名密码错误

const HEADER_SIZE = 16
const HEADER_STRUCT = "Nlength/Ntype/Nuid/Nserid"
const HEADER_PACK = "NNNN"

const DECODE_PHP = 1     //使用PHP的serialize打包
const DECODE_JSON = 2    //使用json_encode打包
const DECODE_MSGPACK = 3 //使用msgpack打包
const DECODE_SWOOLE = 4  //使用swoole_serialize打包
const DECODE_GZIP = 128  //启用GZIP压缩

const ALLOW_IP = 1
const ALLOW_USER = 2

type RpcServer struct {
	headerLength int
	decodeType   int
}

func InitRpcServer() (*RpcServer, error) {
	s := &RpcServer{}
	s.StartServer()
	return s, nil
}

//开启server
func (s *RpcServer) StartServer() {
	// listen to incoming udp packets
	listenSock, err := net.Listen("tcp", "localhost:9904")
	defer listenSock.Close()
	if err != nil {
		log.Fatal(err)
	}
	for {
		newConn, err := listenSock.Accept()
		if err != nil {
			continue
		}
		go s.rpcUnpack(newConn)
	}
}

//解包
func (s *RpcServer) rpcUnpack(conn net.Conn) {
	buf := make([]byte, buffer_maxlen)
	for {
		n, err := conn.Read(buf)

		if err != nil {
			fmt.Println("conn closed")
			return
		}
		//获取包头
		header := &RpcHeader{}
		s.ParsingHeader(buf[0:HEADER_SIZE], header)
		if n != int(header.length+HEADER_SIZE) {
			s.sendError(ERR_UNPACK, conn, header)
			continue
		}
		fmt.Println(header)
		//获取包体
		bodyStr := buf[HEADER_SIZE:n]
		buffer := new(bytes.Buffer)
		buffer.Write(bodyStr)
		fmt.Println(string(bodyStr))
		var request map[string]string
		err1 := json.NewDecoder(buffer).Decode(&request)
		if err1 != nil {
			log.Println(err1)
		}
		fmt.Println(request)
		data := []int{1, 2}
		s.sendSuccess(conn, header, data)
		return
		//fmt.Println("recv msg:", buf[0:n])
		fmt.Println("recv msg:", string(buf[0:n]))
	}
}

//发送错误
func (s *RpcServer) sendError(errorNo int, conn net.Conn, header *RpcHeader) {
	m1 := map[string]interface{}{"errno": errorNo}
	body, err := json.Marshal(m1)
	if err != nil {
		fmt.Println("json.Marshal failed:", err)
		return
	}
	bodyLength := len(body)

	rs := s.HeaderToByte(header, bodyLength, body)
	fmt.Println(string(rs))
	conn.Write(rs)
}

//发送成功
func (s *RpcServer) sendSuccess(conn net.Conn, header *RpcHeader, data interface{}) {
	result := make(map[string]interface{})
	result["errno"] = 0
	result["data"] = data
	body, err := json.Marshal(result)
	if err != nil {
		fmt.Println("json.Marshal failed:", err)
		return
	}
	bodyLength := len(body)
	rs := s.HeaderToByte(header, bodyLength, body)
	//fmt.Println(string(rs))
	conn.Write(rs)
}

//解析字段
func (s *RpcServer) ParsingHeader(one []byte, content *RpcHeader) *RpcHeader {
	ByteToHeader(one[0:4], content, 1)
	ByteToHeader(one[4:8], content, 2)
	ByteToHeader(one[8:12], content, 3)
	ByteToHeader(one[12:16], content, 4)
	return content
}

//流转结构
func ByteToHeader(bt []byte, m *RpcHeader, t int) {
	data := bytes.NewReader(bt)
	var err error
	switch t {
	case 1:
		err = binary.Read(data, binary.BigEndian, &m.length)
	case 2:
		err = binary.Read(data, binary.BigEndian, &m.encodeType)
	case 3:
		err = binary.Read(data, binary.BigEndian, &m.uid)
	case 4:
		err = binary.Read(data, binary.BigEndian, &m.serid)
	}
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
}

func (s *RpcServer) HeaderToByte(header *RpcHeader, length int, body []byte) []byte {
	buf := new(bytes.Buffer)
	byteOrder := binary.BigEndian
	binary.Write(buf, byteOrder, int32(length))
	binary.Write(buf, byteOrder, int32(header.encodeType))
	binary.Write(buf, byteOrder, int32(header.uid))
	binary.Write(buf, byteOrder, int32(header.serid))
	binary.Write(buf, byteOrder, body)
	return buf.Bytes()
}
