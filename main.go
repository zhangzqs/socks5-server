package main

import (
	"bufio"
	"log"
	"net"
)

func main() {
	server, err := net.Listen(
		"tcp",
		"0.0.0.0:1080")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Panicf("Accept failed %v", err)
		}
		go process(conn)
	}
}

func process(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	err := auth(reader, conn)
	if err != nil {
		log.Printf("Client %v auth failed: %v", conn.RemoteAddr(), err)
		return
	}
	log.Println("Client auth success")
	err = connect(reader, conn)
	if err != nil {
		log.Printf("Client %v auth failed: %v", conn.RemoteAddr(), err)
		return
	}
}
