package main

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/Cam1G/Landlord-RTS/internal/netcode"
	"github.com/Cam1G/Landlord-RTS/internal/protocol"
)

type userConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type Config struct {
	Servers map[string]userConfig
}

func createUser(config *Config, server string, reader *bufio.Reader, enc *json.Encoder, conn net.Conn) userConfig {
	var username string
	for {
		fmt.Print("Please enter a username: ")
		username, _ = reader.ReadString('\n')
		username = strings.TrimSuffix(username, "\n")
		err := netcode.SendMessage(conn, protocol.AuthCheckUser, username)
		if err != nil {
			log.Fatalln(err)
		}
		cmd, resp, err := netcode.RecvMessage(conn)
		if err != nil {
			log.Fatalln(err)
		}
		if cmd != protocol.AuthCheckUser {
			log.Println("Server did not serve correct response, try again")
		}
		if resp == "x" {
			log.Println("User already exists, try again")
		} else if resp == "s" {
			break
		} else {
			log.Println("Server error, try again")
		}
	}
	// auto generate password because we do not enforce encryption and don't want to send a potentially important password over the internet
	fmt.Println("Auto generating password")
	passwd := rand.Text()
	config.Servers[server] = userConfig{username, passwd}
	fmt.Println("Creating account on the server")
	err := netcode.SendMessage(conn, protocol.AuthCheckUser, username+" "+passwd)
	if err != nil {
		log.Fatalln(err)
	}
	cmd, resp, err := netcode.RecvMessage(conn)
	if err != nil {
		log.Fatalln(err)
	}
	if cmd != protocol.AuthCreateUser {
		log.Fatalln("Server did not serve correct response")
	}
	if resp == "d" {
		log.Fatalln("User does not exist")
	} else if resp == "p" {
		log.Fatalln("Wrong password")
	} else if resp == "s" {
	} else {
		log.Fatalln("Server error, try again")
	}
	fmt.Println("Done, saving details")
	enc.Encode(config)
	return config.Servers[server]
}

func main() {
	// load/create config file
	var config Config
	sysConfigDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalln(err)
	}
	configDir := filepath.Join(sysConfigDir, "landlord-rts")
	err = os.MkdirAll(configDir, 0o755)
	if err != nil {
		log.Fatalln(err)
	}
	file, err := os.OpenFile(filepath.Join(configDir, "config.json"), os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		log.Fatalln(err)
	}
	configEncoder := json.NewEncoder(file)
	configEncoder.SetIndent("", "  ")

	// ask for server
	reader := bufio.NewReader(os.Stdin)
	var default_server = "landlord-rts.tmkcell.xyz"
	fmt.Print("Select server to connect to (default: " + default_server + "): ")
	server, _ := reader.ReadString('\n')
	if server == "\n" {
		server = default_server
	} else {
		server = strings.TrimSuffix(server, "\n")
	}

	// ask for port
	var default_port = "6767"
	fmt.Print("Port (default: " + default_port + "): ")
	port, _ := reader.ReadString('\n')
	if port == "\n" {
		port = default_port
	} else {
		port = strings.TrimSuffix(port, "\n")
	}

	// connect
	fmt.Print("Attempting to connect...")
	conn, err := net.Dial("tcp", server+port)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Success!")
	defer conn.Close()

	val, ok := config.Servers[server]
	if !ok {
		fmt.Println("Account not found, creating user")
		val = createUser(&config, server, reader, configEncoder, conn)
	}
	for {
		fmt.Printf("Would you like to sign in with %s? (y/n): ", val.Username)
		str, _ := reader.ReadString('\n')
		if str == "y\n" {
			break
		} else if str == "n\n" {
			fmt.Println("Creating new user **(this will delete your other user for this server!)**")
			val = createUser(&config, server, reader, configEncoder, conn)
			break
		}
		fmt.Println("\nSorry, response not understood.")
	}
}
