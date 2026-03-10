package main

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/Cam1G/Landlord-RTS/internal/common"
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

func createUser(config *Config, server string, reader *bufio.Reader, conn net.Conn) userConfig {
	var username string
	for {
		fmt.Print("Please enter a username: ")
		username, _ = reader.ReadString('\n')
		username = strings.TrimSuffix(username, "\n")
		// auto generate password because we do not enforce encryption and don't want to send a potentially important password over the internet
		fmt.Println("Auto generating password (this is auto generated for your safety)")
		passwd := rand.Text()
		config.Servers[server] = userConfig{username, passwd}
		fmt.Println("Creating account on the server")
		err := netcode.SendMessage(conn, protocol.AuthCreateUser, username+" "+passwd)
		if err != nil {
			log.Fatalln(err)
		}
		_, err = netcode.RecvMessageHandleErr(conn, protocol.AuthCreateUser)
		if err != nil {
			log.Println(err.Error() + ", try again")
		} else {
			break
		}
	}
	fmt.Println("Done, saving details")
	saveConfig(config)
	return config.Servers[server]
}

func saveConfig(config *Config) {
	raw, err := json.Marshal(*config)
	if err != nil {
		log.Println("Warning: unable to generate config: " + err.Error())
	}
	err = os.WriteFile(confPath, raw, 0o644)
	if err != nil {
		log.Printf("Warning: failed to save config file (%s). Outputting it instead.\n\n%s\n\n", err.Error(), raw)
	}
}

var confPath string

func main() {
	// load/create config file
	var config Config
	configDir, err := common.CreateConfigDir()
	if err != nil {
		log.Fatalln("Error creating config dir: " + err.Error())
	}
	confPath = filepath.Join(configDir, "config.json")
	file, err := os.OpenFile(confPath, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		log.Fatalln("Error opening config.json: " + err.Error())
	}
	err = json.NewDecoder(file).Decode(&config)
	if err != nil && err != io.EOF {
		log.Fatalln("Error decoding the config.json: " + err.Error())
	}
	// we need it later on
	file.Close()

	// make map if there's nothing there
	if config.Servers == nil {
		config.Servers = make(map[string]userConfig)
		saveConfig(&config)
	}

	// ask for server
	reader := bufio.NewReader(os.Stdin)
	var default_server = "localhost"
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
	conn, err := net.Dial("tcp", server+":"+port)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Success!")
	defer conn.Close()

	user, ok := config.Servers[server]
	if !ok {
		fmt.Println("Account not found, creating user")
		user = createUser(&config, server, reader, conn)
	}
	for {
		fmt.Printf("Would you like to sign in with %s? (Y/n): ", user.Username)
		str, _ := reader.ReadString('\n')
		if str == "y\n" || str == "\n" {
			break
		} else if str == "n\n" {
			fmt.Println("Creating new user **(this will delete your other user for this server!)**")
			user = createUser(&config, server, reader, conn)
			break
		}
		fmt.Println("\nSorry, response not understood.")
	}

	// auth with server
	err = netcode.SendMessage(conn, protocol.Auth, user.Username+" "+user.Password)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = netcode.RecvMessageHandleErr(conn, protocol.Auth)
	if err != nil {
		log.Fatalln(err)
	}
}
