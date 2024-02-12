package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/dantecatalfamo/remotex/pkg/server"
	"golang.org/x/term"
)

const listenAddress = "localhost:3344"
const defaultConfigPath = "/etc/remotex/remotex.json"

func main() {
	configPath := flag.String("config", defaultConfigPath, "Configutation file")
	flag.Parse()

	cmd := flag.Args()
	if len(cmd) == 0 {
		usage()
		return
	}

	if cmd[0] == "newconfig" {
		if len(cmd) != 2 {
			fmt.Println("usage: remotex-server newconfig <path>")
			os.Exit(1)
		}
		if err := server.WriteNewConfig(cmd[1]); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	if *configPath == "" {
		fmt.Println("Specify config path")
		os.Exit(1)
	}

	config, err := server.ReadAndInitializeConfig(*configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	log.Printf("Server config: %+v", config)

	switch cmd[0] {
	case "server":
		log.Printf("Listening on http://%s", config.ListenAddress)
		if err := server.RunServer(config); err != nil {
			log.Fatal(err)
		}
	case "useradd":
		if len(cmd) < 2 {
			fmt.Println("usage: remotex-server useradd <username>")
			return
		}
		user := cmd[1]
		if err := server.CreateUser(config, user); err != nil {
			log.Fatal(err)
		}
		log.Printf("Added user %s", user)
	case "userdel":
		if len(cmd) < 2 {
			fmt.Println("usage: remotex-server userdel <username>")
			return
		}
		user := cmd[1]
		if err := server.DeleteUser(config, user); err != nil {
			log.Fatal(err)
		}
		log.Printf("Deleted user %s", user)
	case "passwd":
		if len(cmd) < 2 {
			fmt.Println("usage: remotex-server passwd <username> [password]")
			return
		}
		user := cmd[1]
		var password string
		if len(cmd) == 3 {
			password = cmd[2]
		} else {
			fmt.Print("Enter new password: ")
			passwd, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println()
			fmt.Print("Enter password again: ")
			passwd2, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println()
			if !slices.Equal(passwd, passwd2) {
				log.Fatal("Passwords are not the same")
			}
			password = string(passwd)
		}
		if err := server.SetUserPassword(config, user, password); err != nil {
			log.Fatal(err)
		}
		log.Printf("Password set for %s\n", user)
	case "tokenadd":
		if len(cmd) < 3 {
			fmt.Println("usage: remotex-server tokenadd <username> <description>")
			os.Exit(1)
		}
		user := cmd[1]
		desc := cmd[2]
		token, err := server.CreateUserToken(config, user, desc)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Token: %s\n", token)
	case "tokendel":
		if len(cmd) < 2 {
			fmt.Println("usage: remotex-server tokendel <token>")
			return
		}
		token := cmd[1]
		if err := server.DeleteUserToken(config, token); err != nil {
			log.Fatal(err)
		}
		log.Println("Token deleted")
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage: remotex-server [options] <command> [args]")
	flag.PrintDefaults()
	fmt.Printf(`
  commands:
    newconfig <file>
    server
    useradd   <username>
    userdel   <username>
    passwd    <username> [password]
    tokenadd  <username>
    tokendel  <token>
`)
}
