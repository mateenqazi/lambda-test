package main

import (
	"crypto/tls"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-stomp/stomp"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	brokerEndpointIP := os.Getenv("MQ_ENDPOINT_IP")
	brokerUsername := os.Getenv("BROKER_USERNAME")
	brokerPassword := os.Getenv("BROKER_PASSWORD")

	// Create a tls.Config with proper settings
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		},
	}

	// Dial the broker using TLS
	netConn, err := tls.Dial("tcp", brokerEndpointIP, tlsConfig)
	if err != nil {
		log.Println("Failed to connect to broker:", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}
	defer netConn.Close()

	// Connect to the STOMP server over the TLS connection
	conn, err := stomp.Connect(netConn, stomp.ConnOpt.Login(brokerUsername, brokerPassword))
	if err != nil {
		log.Println("Failed to connect to the broker:", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}
	defer conn.Disconnect()

	// Further processing...
	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "Success"}, nil
}

func main() {
	lambda.Start(handler)
}
