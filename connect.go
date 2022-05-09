package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

func readHeader(reader *bufio.Reader) (ver, cmd, rsv, atyp byte, err error) {
	buf := make([]byte, 4)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return
	}
	ver, cmd, rsv, atyp = buf[0], buf[1], buf[2], buf[3]
	return
}

func readDstAddr(atyp byte, reader *bufio.Reader) (addr string, err error) {
	switch atyp {
	case atypIPV4:
		buf := make([]byte, 4)
		_, err = io.ReadFull(reader, buf)
		if err != nil {
			err = fmt.Errorf("read atyp failed: %w", err)
			return
		}
		addr = fmt.Sprintf("%d.%d.%d.%d", buf[0], buf[1], buf[2], buf[3])
	case atypHOST:
		hostSize, err := reader.ReadByte()
		if err != nil {
			err = fmt.Errorf("read hostSize failed:%w", err)
			return "", err
		}
		host := make([]byte, hostSize)
		_, err = io.ReadFull(reader, host)
		if err != nil {
			err = fmt.Errorf("read host failed:%w", err)
			return "", err
		}
		addr = string(host)
	case atypIPV6:
		// TODO: IPv6暂不支持
		err = errors.New("IPv6: no supported yet")
	default:
		// 未知的atyp
		err = errors.New("invalid atyp")
	}
	return
}
func readDstPort(reader *bufio.Reader) (port uint16, err error) {
	buf := make([]byte, 2)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		err = fmt.Errorf("read port failed:%w", err)
		return
	}
	port = binary.BigEndian.Uint16(buf)
	return
}
func connect(reader *bufio.Reader, conn net.Conn) error {
	ver, cmd, _, atyp, err := readHeader(reader)
	if err != nil {
		return fmt.Errorf("read header failed:%w", err)
	}
	if ver != socks5Version {
		return fmt.Errorf("not supported ver:%v", ver)
	}
	// TODO: 目前仅支持CONNECT连接
	if cmd != cmdBind {
		return fmt.Errorf("not supported cmd:%v", cmd)
	}

	addr, err := readDstAddr(atyp, reader)
	if err != nil {
		return err
	}
	port, err := readDstPort(reader)
	if err != nil {
		return err
	}

	// 与目标服务器建立连接
	dest, err := net.Dial("tcp", fmt.Sprintf("%v:%v", addr, port))
	if err != nil {
		return fmt.Errorf("dial dst failed:%w", err)
	}
	defer dest.Close()
	log.Println("dial", addr, port)

	// TODO: 这里暂时直接简单地响应
	_, err = conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
	if err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	// 建立双向转发
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_, _ = io.Copy(dest, reader)
		cancel()
	}()

	go func() {
		_, _ = io.Copy(conn, dest)
		cancel()
	}()

	<-ctx.Done()
	return nil
}
