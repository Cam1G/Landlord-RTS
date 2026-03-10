package main

import (
	"bytes"
	"database/sql"
	"flag"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Cam1G/Landlord-RTS/internal/common"
	"github.com/Cam1G/Landlord-RTS/internal/netcode"
	"github.com/Cam1G/Landlord-RTS/internal/protocol"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/argon2"
)

func hash(data string) []byte {
	return argon2.Key([]byte(data), []byte("0,;39H%>E5b6DOiZVW?_||dRZL?h8,a"), 3, 32*1024, 4, 32)
}

func main() {
	sysConfigDir, err := os.UserConfigDir()
	defaultDbPath := filepath.Join(sysConfigDir, "landlord-rts", "server.db")
	port := flag.Int("port", 6767, "port to run server on")
	dbPath := flag.String("db", defaultDbPath, "path to database file")
	flag.Parse()

	var fp string
	if *dbPath == defaultDbPath {
		_, err := common.CreateConfigDir()
		if err != nil {
			log.Fatalln(err)
		}
		fp = defaultDbPath
	} else {
		fp = *dbPath
	}
	log.Printf("Opening database %s", fp)
	db, err := sql.Open("sqlite3", fp)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		username TEXT NOT NULL PRIMARY KEY,
		password TEXT NOT NULL
	);
	`)
	if err != nil {
		log.Fatalln(err)
	}

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
			p, msg, err := netcode.RecvMessage(conn)
			if err != nil {
				log.Println(err)
				return
			}

			switch p {
			case protocol.AuthCreateUser:
				userdata := strings.Split(msg, " ")
				if len(userdata) != 2 {
					log.Println("AuthCreateUser: Expected 2 (username, password) fields, found %i fields", len(userdata))
					netcode.SendMessage(conn, p, "f")
					break
				}
				var exists bool
				err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE username=?);", userdata[0]).Scan(&exists)
				if err != nil {
					log.Println("AuthCreateUser: " + err.Error())
					netcode.SendMessage(conn, p, "f")
					break
				}
				if exists {
					netcode.SendMessage(conn, p, "x")
					break
				}
				_, err = db.Exec("INSERT INTO users VALUES(?,?);", userdata[0], hash(userdata[1]))
				if err != nil {
					log.Println("AuthCreateUser: " + err.Error())
					netcode.SendMessage(conn, p, "f")
					break
				}
				netcode.SendMessage(conn, p, "s")
			case protocol.Auth:
				userdata := strings.Split(msg, " ")
				if len(userdata) != 2 {
					log.Println("Auth: Expected 2 (username, password) fields, found %i fields", len(userdata))
					netcode.SendMessage(conn, p, "f")
					break
				}
				var exists bool
				err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE username=?);", userdata[0]).Scan(&exists)
				if err != nil {
					log.Println("AuthCreateUser: " + err.Error())
					netcode.SendMessage(conn, p, "f")
					break
				}
				if !exists {
					netcode.SendMessage(conn, p, "d")
					break
				}
				var passwdHash string
				err = db.QueryRow("SELECT password FROM users WHERE username=?;", userdata[0]).Scan(&passwdHash)
				if err != nil {
					log.Println("Auth: " + err.Error())
					netcode.SendMessage(conn, p, "f")
					break
				}
				if !bytes.Equal(hash(userdata[1]), []byte(passwdHash)) {
					log.Println("Auth: Wrong password for user " + userdata[0])
					netcode.SendMessage(conn, p, "p")
					break
				}
				netcode.SendMessage(conn, p, "s")
			default:
				log.Printf("Unknown message %d\n", p)
			}

			conn.Close()
		}(conn)
	}
}
