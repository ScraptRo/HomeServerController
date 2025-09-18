package main

import (
	"ServerController/src/API_Handler"
	"ServerController/src/HTML_Handler"
	"ServerController/src/User_Handler"
	"context"
)

var serverRunning = make(chan bool)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	User_Handler.Load_users()
	User_Handler.Load_requests()
	go HTML_Handler.StartWebHoster(serverRunning)
	go API_Handler.StartAPIHoster(ctx, serverRunning)
	for isRunning := range serverRunning {
		if !isRunning {
			break
		}
	}

	println("Server has been stopped.")
	cancel()
	HTML_Handler.StopWebHoster()
	API_Handler.StopAPIHoster()
	HTML_Handler.WaitForWebHoster()
	API_Handler.WaitForAPIHoster()
}
