# ValtheraDB SQL Proxy Server

A Go-based MySQL proxy that forwards SQL queries via HTTP POST to a ValtheraDB backend.  
Supports Bearer token authentication and configurable database name and listening port.

## Configuration

Create a `config.json` file in the same directory as the executable with the following structure:

```json
{
  "server_url": "http://localhost:14785",
  "auth_token": "your-bearer-token",
  "db_name": "your_database_name",
  "port": 4000
}
```

* `server_url`: HTTP endpoint of the ValtheraDB server
* `auth_token`: Bearer token for HTTP authentication
* `db_name`: The default database name to be used in MySQL connections
* `port`: TCP port the proxy will listen on for MySQL clients

## Building

Requires Go 1.20+ installed.

```bash
go build -o valtheradb-sql-proxy main.go
```

## Running

Simply run the compiled binary:

```bash
./valtheradb-sql-proxy
```

The proxy will listen on `127.0.0.1:<port>` (default 4000) and forward SQL queries to ValtheraDB.

## Usage

Connect your MySQL client to the proxy:

```bash
mysql -h 127.0.0.1 -P <port> -u root -p
```

Queries will be forwarded to ValtheraDB via HTTP.

## Limitations

* No TLS/SSL support (add if needed)
* Simple in-memory MySQL user authentication
* Supports only one HTTP backend endpoint
* Expects JSON response with columns and rows fields as arrays

## License

MIT License