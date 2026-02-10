package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	var default_server = "landlord-rts.tmkcell.xyz:6767"
	fmt.Print("Select server to connect to (default: " + default_server + "): ")
	server, _ := reader.ReadString('\n')
	if server == "\n" {
		server = default_server
	}

	var selected uint = 2
	for selected > 1 {
		fmt.Print("Sign in [0] or create account [1]: ")
		str, _ := reader.ReadString('\n')
		str = strings.TrimSuffix(str, "\n")
		selected_int, err := strconv.Atoi(str)
		if err != nil {
			fmt.Println("\nSorry, try again")
			continue
		}
		selected = uint(selected_int)
	}
}
