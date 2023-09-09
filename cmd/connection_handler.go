package main

import (
	"io"
	"log/slog"
	"net"
	"weightogo/loadbalancer"
)

type ConnectionHandler struct {
	Listener     net.Listener
	LoadBalancer loadbalancer.LoadBalancer
	Logger       *slog.Logger
}

func NewConnectionHandler(listener net.Listener, lb loadbalancer.LoadBalancer, logger *slog.Logger) *ConnectionHandler {
	return &ConnectionHandler{
		Listener:     listener,
		LoadBalancer: lb,
		Logger:       logger,
	}
}

func (ch *ConnectionHandler) HandleConnection() {
	for {
		clientConn, err := ch.Listener.Accept()
		if err != nil {
			ch.Logger.Error("Error accepting connection", "err", err.Error())
			continue
		}

		targetServer := ch.LoadBalancer.PickServer()
		if !targetServer.Alive {
			for !targetServer.Alive {
				targetServer = ch.LoadBalancer.PickServer()
			}
		}

		targetAddr := targetServer.Address
		targetServer.IncreaseConnections()

		go func() {
			targetConn, err := net.Dial("tcp", targetAddr)
			if err != nil {
				ch.Logger.Error("Error connecting to target server", "err", err.Error(), "target", targetAddr)
				return
			}
			defer targetServer.DecreaseConnections()

			go func() {
				defer clientConn.Close()
				defer targetConn.Close()

				io.Copy(targetConn, clientConn)
			}()

			go func() {
				defer clientConn.Close()
				defer targetConn.Close()

				io.Copy(clientConn, targetConn)
			}()
		}()
	}
}
