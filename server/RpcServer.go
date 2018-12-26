package server

import (
	"GoStatistics/myTool"
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
const HEADER_PACK = "NNNN"

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
		go s.handleConn(newConn)
	}
}

//解包
func (s *RpcServer) handleConn(conn net.Conn) {
	buf := make([]byte, buffer_maxlen)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("conn closed")
			return
		}
		//获取包头
		header := &RpcHeader{}
		myTool.Unpack(HEADER_PACK, buf[0:HEADER_SIZE], &header.length, &header.encodeType, &header.uid, &header.serid)
		if n != int(header.length+HEADER_SIZE) {
			s.sendError(ERR_UNPACK, conn, header)
			continue
		}
		//获取包体
		bodyStr := buf[HEADER_SIZE:n]
		buffer := new(bytes.Buffer)
		buffer.Write(bodyStr)
		//fmt.Println(string(bodyStr))
		var request map[string]string
		err1 := json.NewDecoder(buffer).Decode(&request)
		if err1 != nil {
			log.Println(err1)
		}
		data := []int{1, 2}
		s.sendSuccess(conn, header, data)
		return
		//fmt.Println("recv msg:", string(buf[0:n]))
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
	rs := s.packData(header, bodyLength, body)
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
	rs := s.packData(header, bodyLength, body)
	//fmt.Println(string(rs))
	conn.Write(rs)
}

func (s *RpcServer) packData(header *RpcHeader, length int, body []byte) []byte {
	buf := new(bytes.Buffer)
	byteOrder := binary.BigEndian
	head, err := myTool.Pack(HEADER_PACK, int32(length), int32(header.encodeType), int32(header.uid), int32(header.serid))
	//fmt.Println(string(head))
	if err != nil {
		fmt.Println("json.Marshal failed:", err)
	}
	binary.Write(buf, byteOrder, head)
	binary.Write(buf, byteOrder, body)
	return buf.Bytes()
}
