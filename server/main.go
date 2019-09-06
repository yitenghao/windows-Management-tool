package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/axgle/mahonia"
)

type SendJson struct {
	SendType    string
	CommandName string
	Params      []string
}

var conns map[string]net.Conn

const version = "1.0"

func main() {
	defer func() {
		recover()
	}()
	conns = make(map[string]net.Conn)
	// tcp 监听并接受端口
	l, err := net.Listen("tcp", "0.0.0.0:6611")
	if err != nil {
		fmt.Println(err)
		return
	}
	//最后关闭
	defer l.Close()
	fmt.Println("tcp服务端开始监听10000端口...")
	// 使用循环一直接受连接
	go func() {
		//从命令行接收指令
		input := bufio.NewScanner(os.Stdin)
		for input.Scan() {
			line := input.Text()
			// 输入exit时 结束
			if line == "exit" {
				break
			}
			arr := strings.Split(line, " ")
			for index, item := range arr {
				if item == "" {
					arr = append(arr[:index], arr[index+1:]...)
				}
			}
			fmt.Println(arr)
			if len(arr) > 0 {
				switch arr[0] {
				case "-help":
					fmt.Println("-a                     获取所有连接")
					fmt.Println("-t [ip:port] [Command] 向指定连接发送信息")
					fmt.Println("-v                     获取服务端版本信息")
					fmt.Println("[Command]              向全部连接发送信息")
					fmt.Println("if CommandName==-v     返回客户端版本信息")
				case "-a":
					GetAllConn(conns)
				case "-t":
					var send SendJson
					send.SendType = "111"
					send.CommandName = arr[2]
					send.Params = arr[2:]
					SendTo(arr[1], send)
				case "-v":
					fmt.Println("server ", version)
				default:
					var send SendJson
					send.SendType = "111"
					send.CommandName = arr[0]
					send.Params = arr[1:]
					Broadcasting(send)
				}
			}
		}
		panic("exit")
	}()
	for {
		//Listener.Accept() 接受连接
		c, err := l.Accept()
		if err != nil {
			return
		}
		//处理tcp请求
		go handleConnection(c)
	}
}

//用于读取客户端传输的数据
func handleConnection(c net.Conn) {
	// fmt.Println("tcp服务端开始处理请求...")
	fmt.Println(c.RemoteAddr().String() + " 连接...")
	conns[c.RemoteAddr().String()] = c
	defer func(conn net.Conn) {
		delete(conns, c.RemoteAddr().String())
		conn.Close()
	}(c)
	r := bufio.NewReader(c)
	// var i int
	for {
		//读取响应
		data, err := r.ReadBytes('\n')
		if err != nil || io.EOF == err {
			fmt.Println(err)
			break
		}
		fmt.Print(ConvertToString(string(data), "GBK", "UTF-8"))
	}
	fmt.Println(c.RemoteAddr().String() + " 断开连接...")
}

//ConvertToString 编码转换 src是待转换的字符串 srccode是src的编码 tagcode是要转成的编码
func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

//GetAllConn 获取所有连接
func GetAllConn(conns map[string]net.Conn) {
	var i int
	for key := range conns {
		i++
		fmt.Println(i, ". ", key)
	}
}

//SendTo 向指定连接发送指令
func SendTo(who string, send SendJson) {
	conn, ok := conns[who]
	if ok {
		bt, _ := json.Marshal(send)
		_, err := conn.Write(append(bt, '\n'))
		if err != nil {
			fmt.Println(who, " err:", err.Error())
		} else {
			fmt.Println("send to ", who, " success")
		}
	} else {
		fmt.Println("The " + who + " doesn't exist")
	}
}

//Broadcasting 用于向所有客户端发送指令
func Broadcasting(send SendJson) {
	var i int
	for key, conn := range conns {
		if conn != nil {
			bt, _ := json.Marshal(send)
			_, err := conn.Write(append(bt, '\n'))
			if err != nil {
				fmt.Println(key, " err:", err.Error())
			} else {
				i++
			}

		}
	}
	fmt.Println("send num:", i)
}
