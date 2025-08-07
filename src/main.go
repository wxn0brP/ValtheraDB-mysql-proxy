package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"

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

type Config struct {
	ServerURL string `json:"server_url"`
	AuthToken string `json:"auth_token"`
	DBName    string `json:"db_name"`
	Port      int    `json:"port"`
}

type SQLHandlerData struct {
	serverURL string
	authToken string
	dbName    string
	port      int
}

var ServerCfg *SQLHandlerData

func readConfig(configPath string) (*SQLHandlerData, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	if cfg.ServerURL == "" || cfg.AuthToken == "" || cfg.DBName == "" || cfg.Port == 0 {
		return nil, fmt.Errorf("config missing required fields")
	}

	return &SQLHandlerData{
		serverURL: cfg.ServerURL,
		authToken: cfg.AuthToken,
		dbName:    cfg.DBName,
		port:      cfg.Port,
	}, nil
}

func (h *SQLHandler) HandleQuery(query string) (*mysql.Result, error) {
	// Encode request body
	requestBody, err := json.Marshal(map[string]string{
		"query": query,
		"db":    string(ServerCfg.dbName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode request body: %w", err)
	}

	// Build request
	req, err := http.NewRequest("POST", ServerCfg.serverURL+"/q/sql-proxy", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", ServerCfg.authToken)

	// Send request using custom client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for unsupported status codes
	if resp.StatusCode == http.StatusBadRequest ||
		resp.StatusCode == http.StatusForbidden ||
		resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("unsupported query (HTTP %d)", resp.StatusCode)
	}

	// Handle other HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("SQL proxy server error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// Read and decode JSON response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var data struct {
		Columns []string        `json:"columns"`
		Rows    [][]interface{} `json:"rows"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w\nraw body:\n%s", err, body)
	}

	// Build mysql.Result
	resultSet, err := mysql.BuildSimpleTextResultset(data.Columns, data.Rows)
	if err != nil {
		return nil, fmt.Errorf("failed to build resultset: %w", err)
	}

	return &mysql.Result{Resultset: resultSet}, nil
}

func main() {
	var err error
	ServerCfg, err = readConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	// Create a new server
	srv := server.NewServer(
		"8.0.0",
		mysql.DEFAULT_COLLATION_ID,
		mysql.AUTH_CACHING_SHA2_PASSWORD,
		nil, // no public key
		nil, // disable tls
	)

	// Listen for connections on localhost port 4000
	l, err := net.Listen("tcp", "127.0.0.1:"+fmt.Sprint(ServerCfg.port))
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
