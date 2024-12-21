package gsieve

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

const logo = `
 ______     ______     __     ______     __   __   ______        ______     ______     ______     __  __     ______
/\  ___\   /\  ___\   /\ \   /\  ___\   /\ \ / /  /\  ___\      /\  ___\   /\  __ \   /\  ___\   /\ \_\ \   /\  ___\
\ \ \__ \  \ \___  \  \ \ \  \ \  __\   \ \ \'/   \ \  __\      \ \ \____  \ \  __ \  \ \ \____  \ \  __ \  \ \  __\
 \ \_____\  \/\_____\  \ \_\  \ \_____\  \ \__|    \ \_____\     \ \_____\  \ \_\ \_\  \ \_____\  \ \_\ \_\  \ \_____\
  \/_____/   \/_____/   \/_/   \/_____/   \/_/      \/_____/      \/_____/   \/_/\/_/   \/_____/   \/_/\/_/   \/_____/ `

func handleConnection(conn net.Conn, gs *GSieve) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		// 读取客户端发送的数据
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading:", err)
			return
		}

		// 解析命令
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		// 处理不同的命令
		switch parts[0] {
		case "set":
			if len(parts) != 4 {
				conn.Write([]byte("error: use 'set key value seconds' format\n"))
				continue
			}
			if ttl, err := strconv.Atoi(parts[3]); err != nil {
				fmt.Println("ttl格式错误, 使用以s为单位的数字")
			} else {
				gs.Insert(parts[1], []byte(parts[2]), ttl)
				conn.Write([]byte("OK\n"))
			}

		case "get":
			if len(parts) != 2 {
				conn.Write([]byte("error: use 'get key' format\n"))
				continue
			}
			if value, exists := gs.Get(parts[1]); exists {
				value = append(value, '\n')
				conn.Write(value)
			} else {
				conn.Write([]byte("key not exist\n"))
			}

		case "del":
			if len(parts) != 2 {
				conn.Write([]byte("error: use 'del key' format\n"))
				continue
			}
			gs.Delete(parts[1])
			conn.Write([]byte("ok\n"))
		case "for":
			for _, v := range gs.ForEach() {
				conn.Write(v)
				conn.Write([]byte("\r"))
			}
			conn.Write([]byte("\n"))
		default:
			conn.Write([]byte("error: unknown command\n"))
		}
	}
}

func Run(address string, cap, ttlRange, cronJobTime int) {
	gs := NewTTLGSieve(cap, ttlRange, cronJobTime)
	fmt.Println(logo)

	listen, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listen.Close()
	log.Printf("Starting GSieve server on http://" + address)
	for {
		// 接受连接
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
	}
		go handleConnection(conn, gs) // 并发处理每个客户端连接
	}
}

