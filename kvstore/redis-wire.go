package kvstore

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// KVStoreInterface defines the methods that our KV store must implement
type KVStoreInterface interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Del(key string) bool
	Keys() []string
	Exists(key string) bool
	Scan(cursor int, match string, count int, keyType string) (int, []string) 

}

// RedisServer represents our Redis-compatible server
type RedisServer struct {
	store KVStoreInterface
	port  int
}

// NewRedisServer creates a new RedisServer instance
func NewRedisServer(store KVStoreInterface) *RedisServer {
	port := 6379 // Default Redis port
	if envPort := os.Getenv("REDIS_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}
	return &RedisServer{store: store, port: port}
}

// Start begins listening for connections
func (s *RedisServer) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("Redis-compatible server listening on port %d\n", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *RedisServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// Send initial prompt
	conn.Write([]byte(fmt.Sprintf("%s:%d> ", conn.LocalAddr().(*net.TCPAddr).IP, s.port)))

	for {
		cmd, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		response := s.handleCommand(strings.TrimSpace(cmd))
		conn.Write([]byte(response))

		// Send prompt after each command
		conn.Write([]byte(fmt.Sprintf("%s:%d> ", conn.LocalAddr().(*net.TCPAddr).IP, s.port)))
	}
}
func parseCommand(cmd string) []string {
	var parts []string
	var current string
	inQuotes := false
	escapeNext := false

	for _, char := range cmd {
		if escapeNext {
			current += string(char)
			escapeNext = false
		} else if char == '\\' {
			escapeNext = true
		} else if char == '"' {
			inQuotes = !inQuotes
		} else if unicode.IsSpace(char) && !inQuotes {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

func (s *RedisServer) handleCommand(cmd string) string {
	parts := parseCommand(cmd)
	if len(parts) == 0 {
		return "-ERR empty command\r\n"
	}

	switch strings.ToLower(parts[0]) {
	case "ping":
		return "+PONG\r\n"
	case "get":
		if len(parts) != 2 {
			return "-ERR wrong number of arguments for 'get' command\r\n"
		}
		val, err := s.store.Get(parts[1])
		if err != nil {
			return "$-1\r\n"
		}
		return fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
	case "set":
		if len(parts) != 3 {
			return "-ERR wrong number of arguments for 'set' command\r\n"
		}
		err := s.store.Set(parts[1], parts[2])
		if err != nil {
			return "-ERR internal error\r\n"
		}
		return "+OK\r\n"
	case "del":
		if len(parts) != 2 {
			return "-ERR wrong number of arguments for 'del' command\r\n"
		}
		if s.store.Del(parts[1]) {
			return ":1\r\n"
		}
		return ":0\r\n"
	case "keys":
		if len(parts) != 2 || parts[1] != "*" {
			return "-ERR wrong number of arguments for 'keys' command\r\n"
		}
		keys := s.store.Keys()
		response := fmt.Sprintf("*%d\r\n", len(keys))
		for _, key := range keys {
			response += fmt.Sprintf("$%d\r\n%s\r\n", len(key), key)
		}
		return response
	case "exists":
		if len(parts) != 2 {
			return "-ERR wrong number of arguments for 'exists' command\r\n"
		}
		if s.store.Exists(parts[1]) {
			return ":1\r\n"
		}
		return ":0\r\n"
	case "scan":
		if len(parts) < 2 {
			return "-ERR wrong number of arguments for 'scan' command\r\n"
		}
		cursor, err := strconv.Atoi(parts[1])
		if err != nil {
			return "-ERR invalid cursor\r\n"
		}
		
		count := 10 // default count
		match := ""
		keyType := ""
		
		for i := 2; i < len(parts); i += 2 {
			if i+1 >= len(parts) {
				return "-ERR syntax error\r\n"
			}
			switch strings.ToLower(parts[i]) {
			case "count":
				count, err = strconv.Atoi(parts[i+1])
				if err != nil {
					return "-ERR invalid count\r\n"
				}
			case "match":
				match = parts[i+1]
			case "type":
				keyType = strings.ToLower(parts[i+1])
			default:
				return "-ERR syntax error\r\n"
			}
		}
		
		nextCursor, keys := s.store.Scan(cursor, match, count, keyType)
		response := fmt.Sprintf("*2\r\n$%d\r\n%d\r\n", len(strconv.Itoa(nextCursor)), nextCursor)
		response += fmt.Sprintf("*%d\r\n", len(keys))
		for _, key := range keys {
			response += fmt.Sprintf("$%d\r\n%s\r\n", len(key), key)
		}
		return response
	
	default:
		return "-ERR unknown command '" + parts[0] + "'\r\n"
	}
}