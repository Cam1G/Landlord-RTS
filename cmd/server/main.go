package main

import (
	"flag"
	"log"
	"net"
	"strconv"
)

func handleClient(conn net.Conn) {
	defer conn.Close()
}

func main() {
	port := flag.Int("port", 6767, "port to run server on")
	flag.Parse()

	listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(*port))
	if err != nil {
		log.Fatalln(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleClient(conn)
	}
}
