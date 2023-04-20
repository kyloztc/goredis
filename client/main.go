package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Printf("net dial err: %v\n", err)
		return
	}
	defer conn.Close()
	inputReader := bufio.NewReader(os.Stdin)
	for {
		input, _ := inputReader.ReadString('\n') // 读取用户输入
		inputInfo := strings.Trim(input, "\r\n")
		if strings.ToUpper(inputInfo) == "Q" { // 如果输入q就退出
			return
		}
		_, err := conn.Write([]byte(inputInfo)) // 发送数据
		if err != nil {
			return
		}
		//buf := [512]byte{}
		//n, err := conn.Read(buf[:])
		//if err != nil {
		//	fmt.Println("recv failed, err:", err)
		//	return
		//}
		//fmt.Println(string(buf[:n]))
	}
}