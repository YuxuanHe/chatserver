package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
)

var idAssignmentChan = make(chan string)
var conns = make(map[string]net.Conn)

func ConnectionHandler(conn net.Conn, messages chan string) {
	b := bufio.NewReader(conn)
	client_id := <-idAssignmentChan
	conns[client_id] = conn

	for {
		line, err := b.ReadBytes('\n')
		if err != nil {
			conn.Close()
			break
		}
		reg := regexp.MustCompile(`^whoami:`)
		receiveStr := ""
		if len(reg.FindAllString(string(line), -1)) > 0 {
			receiveStr = "chitter: " + client_id + "\n"
		} else {
			receiveStr = client_id + ": " + string(line)
		}
		messages <- receiveStr
	}
}

func BroadcastHandler(conns *map[string]net.Conn, messages chan string) {
	for {
		msg := <-messages
		//fmt.Println(msg)

		reg1 := regexp.MustCompile(`^\d:`)
		reg2 := regexp.MustCompile(`^chitter:`)
		reg3 := regexp.MustCompile(`^all:`)
		reg4 := regexp.MustCompile(`^all :`)
		if len(reg1.FindAllString(string(msg[3:]), -1)) > 0 {
			if int(msg[3])-47 > len((*conns)) {
				continue
				//os.Exit(1);
			}
			_, err := (*conns)[string(msg[3])].Write([]byte(msg[6:]))
			if err != nil {
				fmt.Println("This connection is not valid")
				delete(*conns, string(msg[3]))
			}
		} else if len(reg2.FindAllString(string(msg), -1)) > 0 {
			_, err := (*conns)[string(msg[9])].Write([]byte(msg))
			if err != nil {
				fmt.Println("This connection is not valid")
				delete(*conns, string(msg[9]))
			}

		} else if len(reg3.FindAllString(string(msg[3:]), -1)) > 0 {
			msg = msg[0:3] + msg[8:]
			for key, value := range *conns {
				_, err := value.Write([]byte(msg))
				if err != nil {
					fmt.Println("This connection is not valid")
					delete(*conns, key)
				}

			}

		} else if len(reg4.FindAllString(string(msg[3:]), -1)) > 0 {
			msg = msg[0:3] + msg[9:]
			for key, value := range *conns {
				_, err := value.Write([]byte(msg))
				if err != nil {
					fmt.Println("This connection is not valid")
					delete(*conns, key)
				}

			}

		} else {
			for key, value := range *conns {
				_, err := value.Write([]byte(msg))
				if err != nil {
					fmt.Println("This connection is not valid")
					delete(*conns, key)
				}

			}
		}
	}
}
func IdManager() {
	var i uint64
	for i = 0; ; i++ {
		idAssignmentChan <- strconv.FormatUint(i, 10)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usege: chitter <port-number>\n")
		os.Exit(1)
		return
	}
	port := os.Args[1]
	fmt.Println("I'll be using port:", port)

	server, err := net.Listen("tcp", ":"+port)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Can't connect to port")
		os.Exit(1)
	}

	fmt.Println("So far so good")
	go IdManager()
	messages := make(chan string, 10)
	go BroadcastHandler(&conns, messages)
	for {
		conn, err := server.Accept()
		if err != nil {
			fmt.Println("Can not accept")
			conn.Close()
			os.Exit(1)
		}
		go ConnectionHandler(conn, messages)

	}

}
