package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-stomp/stomp"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	brokerEndpointIP := os.Getenv("MQ_ENDPOINT_IP")
	brokerUsername := os.Getenv("BROKER_USERNAME")
	brokerPassword := os.Getenv("BROKER_PASSWORD")

	// Remove "ssl://" prefix if present
	if strings.HasPrefix(brokerEndpointIP, "ssl://") {
		brokerEndpointIP = strings.TrimPrefix(brokerEndpointIP, "ssl://")
	}

	brokerEndpointIP = "b-b1847bf5-ab6a-4ca4-af8b-1874261411ac-1.mq.us-west-1.amazonaws.com:61614"

	// Load system CAs and add any custom CA if required
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		log.Println("Failed to load system CAs:", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	// If you have custom CA, uncomment the following lines and add the CA certificate
	/*
	   customCA := []byte(`-----BEGIN CERTIFICATE-----
	   ...
	   -----END CERTIFICATE-----`)
	   if ok := rootCAs.AppendCertsFromPEM(customCA); !ok {
	       log.Println("Failed to append custom CA")
	       return events.APIGatewayProxyResponse{StatusCode: 500}, errors.New("failed to append custom CA")
	   }
	*/

	// Create a tls.Config with proper settings
	tlsConfig := &tls.Config{
		RootCAs:    rootCAs,
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		},
		InsecureSkipVerify: true, // Use this only for testing purposes; not recommended for production
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
