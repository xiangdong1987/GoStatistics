package tool

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

func Long2ip(ip int32) string {
	return fmt.Sprintf("%d.%d.%d.%d", ip>>24, ip<<8>>24, ip<<16>>24, ip<<24>>24)
}

func pack(format string, params ...interface{}) (rs []byte, err error) {
	if len(format) != len(params) {
		err = errors.New("Format is not correct ")
	}
	i := 0
	buf := new(bytes.Buffer)
	byteOrder := binary.BigEndian
	for _, value := range params {
		if string(format[i]) == "N" {
			binary.Write(buf, byteOrder, value)
		}
		i++
	}
	return buf.Bytes(), err
}

func unpack(format string, data []byte, params ...interface{}) error {
	if len(format) != len(params) {
		return errors.New("Format is not correct ")
	}
	buffer := bytes.NewReader(data)
	var err error
	i := 0
	for _, value := range params {
		if string(format[i]) == "N" {
			err = binary.Read(buffer, binary.BigEndian, &value)
		}
		i++
	}
	return err
}
