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

func main() {
	conns = make(map[string]net.Conn)
	// tcp 监听并接受端口
	l, err := net.Listen("tcp", "0.0.0.0:10000")
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
			// fmt.Printf("%q", arr)
			if len(arr) > 0 {
				var send SendJson
				send.SendType = "111"
				send.CommandName = arr[0]
				send.Params = arr[1:]
				Broadcasting(send)
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

func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

//用于向所有客户端发送指令
func Broadcasting(send SendJson) {
	var i int
	for _, conn := range conns {
		if conn != nil {
			bt, _ := json.Marshal(send)
			conn.Write(append(bt, '\n'))
			i++
		}
	}
	fmt.Println("send num:", i)
}
