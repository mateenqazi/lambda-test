package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-stomp/stomp/v3"
)

func main() {
	lambda.Start(handler)
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// Get the broker endpoint
	brokerEndpointIP := os.Getenv("MQ_ENDPOINT_IP")
	brokerUsername := os.Getenv("BROKER_USERNAME")
	brokerPassword := os.Getenv("BROKER_PASSWORD")
	log.Println(">>>>>>>>", brokerEndpointIP, brokerUsername, brokerPassword, strings.TrimPrefix(brokerEndpointIP, "stomp+ssl://"))
	// Remove "ssl://" prefix if it exists
	if strings.HasPrefix(brokerEndpointIP, "ssl://") {
		brokerEndpointIP = strings.TrimPrefix(brokerEndpointIP, "ssl://")
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // for testing purposes only
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		},
		KeyLogWriter: os.Stdout, // logs TLS keys to stdout for debugging
	}

	// Create a tls dial and stomp connect to broker
	netConn, err := tls.Dial("tcp", brokerEndpointIP, tlsConfig)
	if err != nil {
		log.Fatalln(">>>>>>>>>>>>>>ERROR>>>>>>>>>", err.Error())
	}
	defer netConn.Close()

	conn, err := stomp.Connect(netConn,
		stomp.ConnOpt.Login(brokerUsername, brokerPassword))
	if err != nil {
		log.Printf("Failed to connect to the broker: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}
	defer conn.Disconnect()

	fmt.Print("connection established")

	// Send a message to a queue on the broker
	queueName := "Demo-Queue"
	message := request.Body
	err = conn.Send(
		queueName,
		"text/plain",
		[]byte(message),
		nil,
	)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	log.Printf("Message sent to the queue: %s", message)

	// Subscribe to a queue on the broker
	sub, err := conn.Subscribe(queueName, stomp.AckAuto)
	if err != nil {
		log.Printf("Failed to subscribe to the queue: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}
	defer sub.Unsubscribe()

	fmt.Print("Connection established, waiting for messages...\n")

	// Listen for and process incoming messages
	var messageBody string
	for {
		msg := <-sub.C
		if msg.Err != nil {
			log.Printf("Failed to receive message: %v", msg.Err)
			return events.APIGatewayProxyResponse{StatusCode: 500}, msg.Err
		}

		// Process the received message (you can modify this part as needed)
		messageBody = string(msg.Body)
		log.Printf("Received message from the queue: %s", messageBody)
		break
	}

	response := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("Message sent: %s and received also %s", "done", messageBody),
	}
	return response, nil
}
