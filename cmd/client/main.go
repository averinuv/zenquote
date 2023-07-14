package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"

	"zenquote/internal/pow"
)

const (
	maxRequestSize   = 1024
	getChallengeCmd  = "get_challenge"
	checkSolutionCmd = "check_solution"
)

type Request struct {
	Cmd  string `json:"cmd"`
	Data string `json:"data,omitempty"`
}

func main() {
	conn, err := net.Dial("tcp", "server:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = conn.Close()
	}()

	reader := bufio.NewReader(conn)

	for i := 0; i < 3; i++ {
		// request challenge from server
		b, _ := json.Marshal(Request{Cmd: getChallengeCmd})
		_, _ = conn.Write([]byte(fmt.Sprintf("%s\n", b)))
		challenge, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Failed to read challenge: %v", err)
		}

		challenge = strings.TrimSpace(challenge)
		hc, err := pow.NewHashcashFromString(challenge)
		if err != nil {
			log.Fatalf("Failed NewHashcashFromString: %v", err)
		}

		// solve the challenge
		err = hc.SolveChallenge()
		if err != nil {
			log.Fatalf("Failed to solve challenge: %v", err)
		}

		// send solved challenge command to server
		solvedChBytes, err := json.Marshal(Request{Cmd: checkSolutionCmd, Data: hc.ToString()})
		if err != nil {
			log.Fatalf("marshal solverd cachecashe failed: %v", err)
		}

		// check the request size
		if len(solvedChBytes) > maxRequestSize {
			log.Fatalf("The request is too large: %d bytes", len(solvedChBytes))
		}

		_, _ = conn.Write([]byte(fmt.Sprintf("%s\n", solvedChBytes)))

		// read server zen quote message
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Failed to read server message: %v", err)
		}
		fmt.Printf("%s", msg)
	}
}
