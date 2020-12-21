package tritonhttp

import (
	"net"
	"log"
	"fmt"
	"io"
	"strings"
	"time"
)

/*
For a connection, keep handling requests until
	1. a timeout occurs or
	2. client closes connection or
	3. client sends a bad request
*/
func (hs *HttpServer) handleConnection(conn net.Conn) {
	remaining := ""
	var message string
	defer func() {
        fmt.Println("Closing connection for: ", conn.RemoteAddr().String())
        conn.Close()
    }()

	timeoutDuration := 5 * time.Second
	buf := make([]byte, 512)
	req := new(HttpRequestHeader)
	req.Valid = true

	for{
		conn.SetReadDeadline(time.Now().Add(timeoutDuration))
		size, err := conn.Read(buf)
		//fmt.Println(err)
		//fmt.Println(timeErr)
		if err!=nil{
			if err==io.EOF{
				//fmt.Println("yes")
				Deadline:=time.Now().Add(timeoutDuration)
					for{
						if !time.Now().Before(Deadline){
							if remaining != "" || message != ""{
								log.Println("timeout! request is incomplete!")
								hs.handleBadRequest(conn)
								return
							} else {
								log.Println("timeout!no requests in 5 seconds!")
								return
							}
						}
						size, err= conn.Read(buf)
						if size != 0 {
							break
						}
					}
			} else if err, ok := err.(net.Error); ok && err.Timeout(){
				if remaining != "" || message != ""{
					hs.handleBadRequest(conn)
				}
				return
			} else {
				log.Println("read error:", err)
				return
			}
		}

		data := buf[:size]
		remaining += string(data)
		for strings.Contains(remaining, "\r\n") {
			idx := strings.Index(remaining, "\r\n")
			message = remaining[:idx]
			if message == "" {
				re := hs.handleResponse(req, conn)
				if re == "close" || req.Connection == "close"{
					return
				}
				req := new(HttpRequestHeader)
				req.Valid = true

			} else if message[:3] == "GET" {
				initial := strings.Fields(message)
				if len(initial) == 3 && initial[0] == "GET" && initial[2] == "HTTP/1.1" {
					req.URL = initial[1]
				} else {
					req.Valid = false
				}
			} else {
				colon := strings.Index(message, ":")
				if colon == -1 {
					req.Valid = false
				} else {
					key := message[:colon]
					value := strings.Fields(message[colon+1:])
					if key == "Host" {
						if len(value) != 0 {
							req.Host = value[0]
						} else {
							req.Valid = false
						}
					}
					if key == "Connection" && len(value) != 0 {
						req.Connection = value[0]
					}
				}
			}
			fmt.Println("Got a message:", message)
			remaining = remaining[idx+2:]
		}
		//log.Println("Received: " + string(data))

	}
	// Start a loop for reading requests continuously
	// Set a timeout for read operation
	// Read from the connection socket into a buffer
	// Validate the request lines that were read
	// Handle any complete requests
	// Update any ongoing requests
	// If reusing read buffer, truncate it before next read
}
