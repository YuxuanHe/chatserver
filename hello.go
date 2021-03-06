package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var idAssignmentChan = make(chan string)

type Client struct {
	conn      net.Conn
	client_id string
}

func ConnectionHandler(conn net.Conn, messages chan string, client_add chan<- Client, client_rm chan<- string) {
	client_id := <-idAssignmentChan
	client_add <- Client{conn, client_id}

	go func() {
		b := bufio.NewReader(conn)

		for {
			line, err := b.ReadBytes('\n')
			if err != nil {
				break
			}
			reg := regexp.MustCompile(`^whoami:`)
			receiveStr := ""
			if len(reg.FindAllString(string(line), -1)) > 0 {
				receiveStr = "chitter: " + client_id
			} else {
				receiveStr = client_id + ": " + string(line)
			}
			messages <- receiveStr
		}
		client_rm <- client_id
		conn.Close()
	}()


}

func BroadcastHandler( messages chan string, client_add <-chan Client, client_rm <-chan string) {
    var conns = make(map[string]net.Conn)
    for {

			select {
				//Three cases: incoming message, client in, client out
			case msg := <-messages:
			//handle message income
				reg1 := regexp.MustCompile(`^chitter:`)
				reg2 := regexp.MustCompile(`^all:`)
				reg3 := regexp.MustCompile(`^all :`)
				// Using regualr expression, it might be a bit complecated, but it works

				pos := strings.Index(msg[strings.Index(msg, ":")+2:], ":")
				if pos == -1 {
					for key, value := range conns {
						_, err := value.Write([]byte(msg))
						if err != nil {
							fmt.Println("Seems wrong.......")
							delete(conns, key)
						}

					}
				}  else if len(reg1.FindAllString(string(msg), -1)) > 0 {
					_, err := conns[string(msg[9:])].Write([]byte(msg+"\n"))
					if err != nil {
						fmt.Println("Seems wrong.......")
						delete(conns, string(msg[9:]))
					}
				} else if len(reg2.FindAllString(string(msg[strings.Index(string(msg),":")+2:]), -1)) > 0 {
					msg = strings.Split(msg, ":")[0]+":"+strings.Split(msg, ":")[2]
					for key, value := range conns {
						_, err := value.Write([]byte(msg))
						if err != nil {
							fmt.Println("")
							delete(conns, key)
						}
					}
				} else if len(reg3.FindAllString(string(msg[strings.Index(string(msg),":")+2:]), -1)) > 0 {
				msg = strings.Split(msg, ":")[0]+":"+strings.Split(msg, ":")[2]
				for key, value := range conns {
					_, err := value.Write([]byte(msg))
					if err != nil {
						fmt.Println("Seems wrong.......")
						delete(conns, key)
					}
				}
			} 	else  {
	                 _, ok := conns[strings.Split(msg,": ")[1]]
	                if ok == false {
											conns[strings.Split(msg,": ")[0]].Write([]byte("Client not exist, try again please\n"))
											continue
										}
	                _, err := conns[strings.Split(msg,": ")[1]].Write([]byte(strings.Split(msg, ":")[0]+":"+strings.Split(msg, ":")[2]))
									if err != nil {
											fmt.Println("Seems wrong.......")
											delete(conns, strings.Split(msg,": ")[1])
										}
				}




			case client_in := <-client_add:
				//handle client in
	      conns[client_in.client_id] = client_in.conn
	      fmt.Println("Client " + client_in.client_id + " join the chat room")


			case client_out := <-client_rm:
			//handle client out
				fmt.Println("Client " + client_out + " left the chat room")
				delete(conns, client_out)


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
		fmt.Fprintf(os.Stderr, "Use more paraments\n")
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

	fmt.Println("Start my server")
	go IdManager()
	messages := make(chan string, 20)
	client_add := make(chan Client)
	client_rm  := make(chan string)
	go BroadcastHandler(messages, client_add, client_rm)
	for {
		conn, err := server.Accept()
		if err != nil {
			fmt.Println("Can not accept")
			conn.Close()
			os.Exit(1)
		}

		go ConnectionHandler(conn, messages, client_add, client_rm)

	}

}
