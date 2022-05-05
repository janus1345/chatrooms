package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", ":8000")
	if err != nil {
		fmt.Println("connect fail:", err)
		return
	}
	defer conn.Close()
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err2 := os.Stdin.Read(buf)
			if err2 != nil {
				fmt.Println("os.Stdin.Read err:", err)
				return
			}
			_, err1 := conn.Write(buf[:n])
			if err1 != nil {
				fmt.Println("conn.Write err:", err)
				return
			}
		}
	}()

	buf := make([]byte, 1024)
	for {
		n, err1 := conn.Read(buf)
		if err != nil {
			fmt.Println("conn.Read err:", err1)
			return
		}
		if n == 0 {
			fmt.Println("logout")
			return
		}
		fmt.Println("服务器发送：", string(buf[:n]))
	}

}
