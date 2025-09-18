package API_Handler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var listener net.Listener
var numberOfConnections int
var connectionsMutex sync.Mutex // Mutex to protect the counter
var wg sync.WaitGroup
var StartTime time.Time
var serverPort int

type user_info struct {
	username           string
	current_connection net.Conn
	is_admin           bool
	close_connection   bool
}

type request_format struct {
	Status  string   `json:"status"`
	Command string   `json:"cmd"`
	Args    []string `json:"args"`
}

var commandsMap = map[string]func(*request_format, *user_info) []byte{
	"console_cmd":            runCommandInConsole,
	"login_attempt":          login_attempt,
	"request_account":        request_account,
	"list_account_requests":  list_account_requests,
	"accept_account_request": accept_account_request,
	"list_user_folder":       list_user_folder,
	"create_user_folder":     create_user_folder,
	"upload_user_file":       upload_user_file,
	"upload_script":          upload_script,
	"list_scripts":           list_scripts,
	"run_script":             run_script,
	"exit":                   close_user_connection,
}

func formatPort() {
	address := listener.Addr().String()
	resultPort, err := strconv.Atoi(address[strings.LastIndex(address, ":")+1:])
	if err != nil {
		println("Unable to format port")
		serverPort = 5050 // set to default port
		return
	}
	serverPort = resultPort
}

func StartAPIHoster(ctx context.Context, stopChannel chan bool) {
	wg.Add(1)
	StartTime = time.Now()
	go func() {
		defer wg.Done()

		// cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// config := &tls.Config{
		// 	Certificates: []tls.Certificate{cert},
		// 	MinVersion:   tls.VersionTLS12,
		// }
		var err error
		listener, err = net.Listen("tcp", ":0")
		if err != nil {
			fmt.Printf("Error starting TCP server: %s\n", err)
			stopChannel <- false
			return
		}
		defer listener.Close()
		formatPort()
		fmt.Println("TCP server is running on port ", serverPort)

		for {
			select {
			case <-ctx.Done():
				fmt.Println("Shutting down TCP server...")
				return
			default:
				conn, err := listener.Accept()
				if err != nil {
					fmt.Printf("Error accepting connection: %s\n", err)
					continue
				}
				go handleConnection(conn)
			}
		}
	}()
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	incrementConnections()
	defer decrementConnections()

	fmt.Printf("New connection established: %s\n", conn.RemoteAddr())
	scanner := bufio.NewScanner(conn)

	// Setting up the user info
	var session_info user_info
	session_info.is_admin = false
	session_info.username = ""
	session_info.close_connection = false
	session_info.current_connection = conn
	// Start handling commands
	for scanner.Scan() {
		text := scanner.Text()
		var m request_format
		err := json.Unmarshal([]byte(text), &m)
		if err != nil {
			println("Unable to parse API command: %s", err)
			continue
		}

		conn.Write(commandsMap[m.Command](&m, &session_info))

		if session_info.close_connection {
			fmt.Printf("Closing connection: %s\n", conn.RemoteAddr())
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading from connection: %s\n", err)
	}
}

func incrementConnections() {
	connectionsMutex.Lock()
	numberOfConnections++
	connectionsMutex.Unlock()
}

func decrementConnections() {
	connectionsMutex.Lock()
	numberOfConnections--
	connectionsMutex.Unlock()
}

func GetNumberOfConnections() int {
	connectionsMutex.Lock()
	defer connectionsMutex.Unlock()
	return numberOfConnections
}

func StopAPIHoster() {
	if listener != nil {
		listener.Close()
		fmt.Println("TCP server stopped.")
	}
}

func WaitForAPIHoster() {
	wg.Wait()
}

func GetServerPort() int {
	return serverPort
}
