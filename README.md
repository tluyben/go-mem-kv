# 🚀 go-mem-kv

An experimental in-memory key-value store with Redis-compatible wire protocol 🧪

[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](https://choosealicense.com/licenses/mit/)
[![Twitter](https://img.shields.io/twitter/url/https/twitter.com/luyben.svg?style=social&label=Follow%20%40luyben)](https://twitter.com/luyben)

## 🚨 Disclaimer

This project is purely experimental and educational. It's designed to help learn about databases and should not be used in any production environment or for any real-world applications. Seriously, don't do it! 😅

## 🌟 Features

- In-memory key-value store
- Redis-compatible wire protocol
- Basic Redis commands support (GET, SET, DEL, KEYS, EXISTS, SCAN)
- Configurable port

## 🛠️ Installation

```bash
go get github.com/tluyben/go-mem-kv
```

## 🚀 Usage

### As a standalone server

1. Clone the repository:

   ```bash
   git clone https://github.com/tluyben/go-mem-kv.git
   cd go-mem-kv
   ```

2. Run the server:

   ```bash
   go run main.go
   ```

   By default, the server runs on port 6379. You can specify a different port using the `-port` flag:

   ```bash
   go run main.go -port 6380
   ```

### As a library

You can use `go-mem-kv` as a library in your Go projects:

```go
package main

import (
    "fmt"
    "github.com/tluyben/go-mem-kv/kvstore"
)

func main() {
    store := kvstore.New()
    server := kvstore.NewRedisServer(store)

    fmt.Println("Starting Redis-compatible server on port 6379")
    err := server.Start()
    if err != nil {
        fmt.Printf("Failed to start server: %v\n", err)
    }
}
```

## 🔧 Redis Examples

You can interact with the server using any Redis client. Here are some examples using the `redis-cli`:

```bash
$ redis-cli
127.0.0.1:6379> SET mykey "Hello, go-mem-kv!"
OK
127.0.0.1:6379> GET mykey
"Hello, go-mem-kv!"
127.0.0.1:6379> EXISTS mykey
(integer) 1
127.0.0.1:6379> DEL mykey
(integer) 1
127.0.0.1:6379> EXISTS mykey
(integer) 0
127.0.0.1:6379> KEYS *
(empty list or set)
```

## 🧪 Embed Examples

You can embed `go-mem-kv` in your Go applications:

```go
package main

import (
    "fmt"
    "github.com/tluyben/go-mem-kv/kvstore"
)

func main() {
    store := kvstore.New()

    // Set a key
    err := store.Set("greeting", "Hello, embedded go-mem-kv!")
    if err != nil {
        fmt.Printf("Error setting key: %v\n", err)
        return
    }

    // Get a key
    value, err := store.Get("greeting")
    if err != nil {
        fmt.Printf("Error getting key: %v\n", err)
        return
    }
    fmt.Printf("Value: %s\n", value)

    // Check if a key exists
    exists := store.Exists("greeting")
    fmt.Printf("Key exists: %v\n", exists)

    // Delete a key
    deleted := store.Del("greeting")
    fmt.Printf("Key deleted: %v\n", deleted)

    // List all keys
    keys := store.Keys()
    fmt.Printf("All keys: %v\n", keys)
}
```

## 🧑‍💻 Contributing

Contributions are welcome! This is an experimental project, so feel free to experiment, learn, and share your ideas. Just remember, this isn't meant for production use!

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 👨‍💼 Author

Created by tluyben ([@luyben](https://twitter.com/luyben) on Twitter, [tluyben](https://github.com/tluyben) on GitHub)

## 🙏 Acknowledgements

Thanks to the Go community and the Redis protocol for inspiring this experimental project. Remember, folks, this is for learning – don't use it for your billion-dollar startup! 😉
