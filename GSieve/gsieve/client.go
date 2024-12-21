package gsieve

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func Connect(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()
	fmt.Println("Successfully connected")

	// 创建输入读取器
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("gsieve " + address + ">")
		command, _ := reader.ReadString('\n')
		command = strings.TrimSpace(command)

		if len(command) == 0 {
			continue
		}

		if command == "exit" {
			fmt.Println("Exiting client...")
			return
		}


		_, err := conn.Write([]byte(command + "\n"))
		if err != nil {
			fmt.Println("Error sending command to server:", err)
			// 连接断开，尝试重新连接
			conn.Close()
			conn, err = net.Dial("tcp", address)
			if err != nil {
				fmt.Println("Error reconnecting to server:", err)
				return
			}
			fmt.Println("Reconnected to server.")
			continue
		}
		// 设置一个超时，用于判断连接是否断开
		conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // 设置超时为60秒

		response, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from server:", err)
			conn.Close()
			// 尝试重新连接
			conn, err = net.Dial("tcp", address)
			if err != nil {
				fmt.Println("Error reconnecting to server:", err)
				return
			}
			fmt.Println("Reconnected to server.")
			continue
		}

		if command == "for" {
			response = response[:max(0, len(response)-2)]
			fmt.Println("Response:", strings.Split(response, "\r"))
			continue
		}

		fmt.Println("Response:", response)
	}
}

func ListenHTTP(address string, serverAddr string) {
	handleHTTPRequest := func(w http.ResponseWriter, r *http.Request) {
		command := r.URL.Query().Get("command")
		if command == "" {
			http.Error(w, "Missing 'command' parameter", http.StatusBadRequest)
			return
		}

		conn, err := net.Dial("tcp", serverAddr)
		if err != nil {
			http.Error(w, "Error connecting to TCP server: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		_, err = conn.Write([]byte(command + "\n"))
		if err != nil {
			http.Error(w, "Error sending command to TCP server: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 设置超时并读取 TCP 服务器的响应
		conn.SetReadDeadline(time.Now().Add(60 * time.Second)) 
		response, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			http.Error(w, "Error reading from TCP server: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(response))
	}

	http.HandleFunc("/", handleHTTPRequest)
	log.Printf("Starting HTTP server on http://" + address)
	log.Fatal(http.ListenAndServe(address, nil))
}
