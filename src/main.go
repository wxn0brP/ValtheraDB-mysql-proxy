package main

import (
	"log"
	"net"
	"strings"

	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/server"
)

type InMemCredProvider struct {
	users map[string]string
}

func (p *InMemCredProvider) CheckUsername(username string) (bool, error) {
	_, ok := p.users[username]
	return ok, nil
}

func (p *InMemCredProvider) GetCredential(username string) (string, bool, error) {
	pass, ok := p.users[username]
	return pass, ok, nil
}

type SQLHandler struct {
	server.EmptyHandler
}

func (h *SQLHandler) HandleQuery(query string) (*mysql.Result, error) {
	if strings.HasPrefix(strings.ToLower(query), "select") {
		// Create a mock result set
		resultSet, err := mysql.BuildSimpleTextResultset(
			[]string{"id", "name"},
			[][]interface{}{
				{1, "John Doe"},
				{2, "Jane Doe"},
			},
		)
		if err != nil {
			return nil, err
		}
		return &mysql.Result{Resultset: resultSet}, nil
	}
	return nil, nil
}

func main() {
	// Create a new server
	srv := server.NewServer(
		"8.0.0",
		mysql.DEFAULT_COLLATION_ID,
		mysql.AUTH_CACHING_SHA2_PASSWORD,
		nil, // no public key
		nil, // disable tls
	)

	// Listen for connections on localhost port 4000
	l, err := net.Listen("tcp", "127.0.0.1:4000")
	if err != nil {
		log.Fatal(err)
	}

	// Accept a new connection once
	c, err := l.Accept()
	if err != nil {
		log.Fatal(err)
	}

	conn, err := srv.NewCustomizedConn(c, &InMemCredProvider{users: map[string]string{"root": ""}}, &SQLHandler{})
	if err != nil {
		log.Fatal(err)
	}

	// as long as the client keeps sending commands, keep handling them
	for {
		if err := conn.HandleCommand(); err != nil {
			log.Fatal(err)
		}
	}
}
