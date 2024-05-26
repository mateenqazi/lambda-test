package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-stomp/stomp/v3"
)

func main() {
	brokerEndpointIP := os.Getenv("MQ_ENDPOINT_IP")
	brokerUsername := os.Getenv("BROKER_USERNAME")
	brokerPassword := os.Getenv("BROKER_PASSWORD")

	log.Println("Broker Endpoint MATEEN:", brokerEndpointIP)
	log.Println("Broker Username MATEEN:", brokerUsername)
	log.Println("Broker Password MATEEN:", brokerPassword)

	// Ensure the broker endpoint is correctly formatted
	if strings.HasPrefix(brokerEndpointIP, "ssl://") {
		brokerEndpointIP = strings.TrimPrefix(brokerEndpointIP, "ssl://")
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // for testing purposes only
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
	}

	netConn, err := tls.Dial("tcp", brokerEndpointIP, tlsConfig)
	if err != nil {
		log.Fatalln("Error connecting to broker MATEEN:", err)
	}
	defer netConn.Close()

	conn, err := stomp.Connect(netConn,
		stomp.ConnOpt.Login(brokerUsername, brokerPassword))
	if err != nil {
		log.Fatalln("Failed to connect to the broker MATEEN:", err)
	}
	defer conn.Disconnect()

	fmt.Println("Connection established")
}
