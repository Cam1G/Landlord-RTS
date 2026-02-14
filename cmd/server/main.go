package main

import (
	"flag"
	"log"
	"net"
	"strconv"

	"github.com/Cam1G/Landlord-RTS/internal/protocol"
)

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

		go func(conn net.Conn) {
			buf := make([]byte, 256)
			_, err := conn.Read(buf)
			if err != nil {
				log.Println(err)
				return
			}

			switch buf[0] {
			case protocol.Auth:
			default:
				log.Printf("Unknown message %d\n", buf[0])
			}

			defer conn.Close()
		}(conn)
	}
}
