package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

func readMethods(reader *bufio.Reader) (methods []byte, err error) {
	// 读取版本号
	ver, err := reader.ReadByte()
	if err != nil {
		err = fmt.Errorf("read ver failed:%w", err)
		return
	}
	if ver != socks5Version {
		err = fmt.Errorf("not supported ver:%v", ver)
		return
	}
	// 读取方法数
	methodSize, err := reader.ReadByte()
	if err != nil {
		err = fmt.Errorf("read methodSize failed:%w", err)
		return
	}
	// 读取所有的鉴权方法
	methods = make([]byte, methodSize)
	_, err = io.ReadFull(reader, methods)
	if err != nil {
		err = fmt.Errorf("read method failed:%w", err)
		return
	}
	log.Println("version", ver, "methods", methods)
	return
}

func sendSupportedMethod(method byte, conn net.Conn) (err error) {
	_, err = conn.Write([]byte{socks5Version, method})
	if err != nil {
		return fmt.Errorf("write failed:%w", err)
	}
	return
}
func auth(reader *bufio.Reader, conn net.Conn) error {
	_, err := readMethods(reader)
	if err != nil {
		return err
	}
	// TODO: 暂时只支持不需要验证手段
	err = sendSupportedMethod(0x00, conn)
	if err != nil {
		return err
	}
	return nil
}
