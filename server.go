package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type Client struct {
	name string
	addr string
	c    chan string
}

var onlineMap map[string]Client

var message = make(chan string)

func WriteMsgToClient(clnt Client, conn net.Conn) {
	for msg := range clnt.c {
		conn.Write([]byte(msg + "\n"))
	}
}

func MakeMsg(clnt Client, msg string) (buf string) {
	buf = "[" + clnt.addr + "]" + clnt.name + ":" + msg
	return buf
}

func main() {
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	defer listener.Close()

	go Manager()
	for {
		conn, err1 := listener.Accept()
		if err1 != nil {
			fmt.Println("err1:", err1)
			return
		}
		defer conn.Close()
		go HandleConnect(conn)
	}
}

func HandleConnect(conn net.Conn) {
	defer conn.Close()
	addr := conn.RemoteAddr().String()
	clnt := Client{addr, addr, make(chan string)}
	onlineMap[addr] = clnt
	go WriteMsgToClient(clnt, conn)
	message <- MakeMsg(clnt, "login")
	isQuit := make(chan bool)
	hasData := make(chan bool)

	go receiveMsg(conn, clnt, isQuit, hasData)

	for {
		select {
		case <-isQuit:
			delete(onlineMap, addr)
			message <- MakeMsg(clnt, "logout")
			return
		case <-hasData:

		case <-time.After(time.Second * 60):
			delete(onlineMap, addr)
			message <- MakeMsg(clnt, "time out leave")
			return

		}
	}
}

func receiveMsg(conn net.Conn, clnt Client, isQuit chan<- bool, hasData chan<- bool) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("read err:", err)
			return
		}
		if n == 0 {
			isQuit <- true
			fmt.Println("用户%s退出登录", clnt.name)
			return
		}
		msg := string(buf[:n-1])
		if msg == "who" && len(msg) == 3 {
			conn.Write([]byte("user list:\n"))
			for _, user := range onlineMap {
				userInfo := user.addr + ":" + user.name + "\n"
				conn.Write([]byte(userInfo))
			}
		} else if len(msg) > 8 && msg[:6] == "rename" {
			newName := strings.Split(msg, "|")[1]
			clnt.name = newName
			onlineMap[conn.RemoteAddr().String()] = clnt
			conn.Write([]byte("rename successful\n"))
		} else {
			message <- MakeMsg(clnt, msg)
		}
		hasData <- true

	}
}

func Manager() {
	onlineMap = make(map[string]Client)
	for mes := range message {
		for _, client := range onlineMap {
			client.c <- mes
		}
	}

}
