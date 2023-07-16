package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"zenquote/api"
	"zenquote/internal/pow"

	"google.golang.org/protobuf/proto"
)

const (
	maxRequestSize = 1024
)

func main() {
	conn, err := connectToServer()
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}

	defer func() {
		_ = conn.Close()
	}()

	reader := bufio.NewReader(conn)

	// Request challenge from server
	challenge, err := getChallenge(reader, conn)
	if err != nil {
		log.Panicf("failed to get challenge: %v", err)
	}

	hashcash, err := pow.NewHashcashFromString(challenge)
	if err != nil {
		log.Panicf("failed to create hashcash from string: %v", err)
	}

	// Solve the challenge
	err = hashcash.SolveChallenge()
	if err != nil {
		log.Panicf("failed to solve challenge: %v", err)
	}

	// Send solved challenge
	err = sendSolution(reader, conn, hashcash.ToString())
	if err != nil {
		log.Panicf("failed to send solution: %v", err)
	}
}

func connectToServer() (net.Conn, error) {
	conn, err := net.Dial("tcp", "server:8080")
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %w", err)
	}

	return conn, nil
}

func getRequestBytes(cmd api.Command, data string) ([]byte, error) {
	reqBytes, err := proto.Marshal(&api.Request{
		Cmd:  cmd,
		Data: data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	return reqBytes, nil
}

func getChallenge(reader *bufio.Reader, conn io.Writer) (string, error) {
	reqBytes, err := getRequestBytes(api.Command_GET_CHALLENGE, "")
	if err != nil {
		return "", err
	}

	_, _ = conn.Write(append(reqBytes, '\n'))

	challengeResponse, err := reader.ReadBytes('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read challenge: %w", err)
	}

	challengeResponse = challengeResponse[:len(challengeResponse)-1] // remove newline

	return parseResponse(challengeResponse)
}

func sendSolution(reader *bufio.Reader, conn io.Writer, solution string) error {
	reqBytes, err := getRequestBytes(api.Command_CHECK_SOLUTION, solution)
	if err != nil {
		return err
	}

	if len(reqBytes) > maxRequestSize {
		return fmt.Errorf("the request is too large: %d bytes", len(reqBytes))
	}

	_, _ = conn.Write(append(reqBytes, '\n'))

	quoteResponse, err := reader.ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("failed to read server message: %w", err)
	}

	quoteResponse = quoteResponse[:len(quoteResponse)-1] // remove newline

	quote, err := parseResponse(quoteResponse)
	if err != nil {
		return err
	}

	fmt.Printf("Zen Quote: %s\n", quote)

	return nil
}

func parseResponse(responseBytes []byte) (string, error) {
	var resp api.Response
	if err := proto.Unmarshal(responseBytes, &resp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.GetStatus() != api.Response_SUCCESS {
		return "", fmt.Errorf("server returned failure status: %s", resp.GetError())
	}

	return resp.GetData(), nil
}
