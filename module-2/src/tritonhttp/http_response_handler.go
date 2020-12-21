package tritonhttp

import (
	"net"
	"fmt"
	"os"
	"path"
	"log"
	"strconv"
	"time"
)

func (hs *HttpServer) handleBadRequest(conn net.Conn) {
	initial := "HTTP/1.1 400 Bad Request\r\n"
	server := "Server: Go-Triton-Server-1.0\r\n"
	connection := "Connection: close\r\n"
	fmt.Printf(initial + server + connection + "\r\n")
	_,err := conn.Write([]byte(initial + server + connection + "\r\n"))
	if err != nil {
		fmt.Println("Error in handleBadRequest!")
	}
	conn.Close()
}

func (hs *HttpServer) handleFileNotFoundRequest(req *HttpRequestHeader, conn net.Conn) {
	initial := "HTTP/1.1 404 Not Found\r\n"
	server := "Server: Go-Triton-Server-1.0\r\n"
	message := initial + server
	if req.Connection == "close" {
		message += "Connection: close\r\n"
	}
	fmt.Printf(message + "\r\n")
	_, err := conn.Write([]byte(message + "\r\n"))
	if err != nil{
		fmt.Println("Error in handleFileNotFoundRequest!")
	}
}

func (hs *HttpServer) handleResponse(requestHeader *HttpRequestHeader, conn net.Conn) (result string) {
	if requestHeader.Valid == false || requestHeader.Host == "" {
		hs.handleBadRequest(conn)
		return "close"
	}
	if requestHeader.URL == "" || requestHeader.URL[:1] != "/" {
		hs.handleBadRequest(conn)
		return "close"
	} else {
		cleanPath := path.Clean(hs.DocRoot + requestHeader.URL)
		if cleanPath[:2] == ".." {
			hs.handleFileNotFoundRequest(requestHeader, conn)
			return
		}
		fileInfo, err := os.Stat(cleanPath)
		if err != nil {
			if os.IsNotExist(err) {
				hs.handleFileNotFoundRequest(requestHeader, conn)
				return
			}
		} else {
			var res HttpResponseHeader
			res.Path = cleanPath
			res.Server = "Go-Triton-Server-1.0"
			if fileInfo.IsDir() {
				fileInfo, _ = os.Stat(cleanPath + "/index.html")
				res.Path = cleanPath + "/index.html"
			}
			t := fileInfo.ModTime()
			res.Last_Modified = t.Format(time.RFC850)
			res.Content_Length = fileInfo.Size()
			filename := fileInfo.Name()
			fileSuffix := path.Ext(filename)
			if hs.MIMEMap[fileSuffix] == "" {
				res.Content_Type = "application/octet-stream"
			} else {
				res.Content_Type = hs.MIMEMap[fileSuffix]
			}
			if requestHeader.Connection == "close" {
				res.Connection = "close"
			}
			//fmt.Printf("%+v\n", res)
			hs.sendResponse(res,conn)
		}
	}
	return
}

func (hs *HttpServer) sendResponse(responseHeader HttpResponseHeader, conn net.Conn) {
	initial := "HTTP/1.1 200 OK\r\n"
	server := "Server: Go-Triton-Server-1.0\r\n"
	last_modified := "Last-Modified: " + responseHeader.Last_Modified +"\r\n"
	Content_Length := "Content-Length: " + strconv.FormatInt(responseHeader.Content_Length, 10) +"\r\n"
	Content_Type := "Content-Type: " + responseHeader.Content_Type +"\r\n"
	message := initial + server + last_modified + Content_Length + Content_Type
	if responseHeader.Connection != "" {
		message += "Connection: close\r\n"
	}
	fmt.Printf(message + "\r\n")
	_,err := conn.Write([]byte(message + "\r\n"))
	if err!=nil{
		log.Println("error with writing to the client!")
	}
	hs.readFile(responseHeader,conn)
	/*if responseHeader.Connection == "close" {
		conn.Close()
	}*/
}


func (hs*HttpServer) readFile(res HttpResponseHeader, conn net.Conn){
	raw_file, _ := os.Open(res.Path)

	defer func() {
		if err := raw_file.Close(); err != nil {
			log.Panicln(err)
		}
	}()

	buf := make([]byte, res.Content_Length)

	n, err := raw_file.Read(buf)
	if err != nil {
		fmt.Println("file.Read err:", err)
		return
	}
	_, err = conn.Write(buf[:n])
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return
	}


}

	// Send headers

	// Send file if required

	// Hint - Use the bufio package to write response
