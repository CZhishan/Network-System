package tritonhttp

import(
	"net"
	"log"
	//"fmt"
)
/** 
	Initialize the tritonhttp server by populating HttpServer structure
**/
func NewHttpdServer(port, docRoot, mimePath string) (*HttpServer, error) {
	var server HttpServer
	// Initialize mimeMap for server to refer
	mime, err := ParseMIME(mimePath)
	if err!=nil{
		panic(err)
	}
	server.MIMEMap = mime
	server.ServerPort = port
	server.DocRoot = docRoot
	server.MIMEPath = mimePath
	server_pointer := &server
	// Return pointer to HttpServer
	return server_pointer, err	
}

/** 
	Start the tritonhttp server
**/
func (hs *HttpServer) Start() (err error) {
	// Start listening to the server port
	c, err := net.Listen("tcp", hs.ServerPort)
	if err != nil {
		log.Println("Failed to listen for tcp connections. Error: ", err)
		return err
	}
	
	// Accept connection from client
	for {
		conn, err := c.Accept()
		if err != nil {
			log.Println("Failed to accept connection ", conn, " due to error ", err)
			continue
		}
		log.Println("Client ", conn.RemoteAddr(), " connected")
		// Spawn a go routine to handle request
		go hs.handleConnection(conn)
	}
	return nil
}

