package kvstore

import (
	"fmt"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
)

func SlowRedisTest(address string) {
	conn, err := redis.Dial("tcp", address)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer conn.Close()

	fmt.Println("Starting slow Redis compatibility test...")

	// Test SET and GET
	fmt.Println("\nTesting SET and GET:")
	_, err = conn.Do("SET", "mykey", "Hello, go-mem-kv!")
	if err != nil {
		fmt.Printf("SET failed: %v\n", err)
	} else {
		fmt.Println("SET mykey \"Hello, go-mem-kv!\"")
		fmt.Println("OK")
	}

	time.Sleep(time.Second) // Pause for readability

	value, err := redis.String(conn.Do("GET", "mykey"))
	if err != nil {
		fmt.Printf("GET failed: %v\n", err)
	} else {
		fmt.Println("GET mykey")
		fmt.Printf("\"%s\"\n", value)
	}

	time.Sleep(time.Second)

	// Test EXISTS
	fmt.Println("\nTesting EXISTS:")
	exists, err := redis.Int(conn.Do("EXISTS", "mykey"))
	if err != nil {
		fmt.Printf("EXISTS failed: %v\n", err)
	} else {
		fmt.Println("EXISTS mykey")
		fmt.Printf("(integer) %d\n", exists)
	}

	time.Sleep(time.Second)

	// Test DEL
	fmt.Println("\nTesting DEL:")
	deleted, err := redis.Int(conn.Do("DEL", "mykey"))
	if err != nil {
		fmt.Printf("DEL failed: %v\n", err)
	} else {
		fmt.Println("DEL mykey")
		fmt.Printf("(integer) %d\n", deleted)
	}

	time.Sleep(time.Second)

	// Test EXISTS again
	fmt.Println("\nTesting EXISTS after DEL:")
	exists, err = redis.Int(conn.Do("EXISTS", "mykey"))
	if err != nil {
		fmt.Printf("EXISTS failed: %v\n", err)
	} else {
		fmt.Println("EXISTS mykey")
		fmt.Printf("(integer) %d\n", exists)
	}

	time.Sleep(time.Second)

	// Test KEYS
	fmt.Println("\nTesting KEYS:")
	keys, err := redis.Strings(conn.Do("KEYS", "*"))
	if err != nil {
		fmt.Printf("KEYS failed: %v\n", err)
	} else {
		fmt.Println("KEYS *")
		if len(keys) == 0 {
			fmt.Println("(empty list or set)")
		} else {
			for _, key := range keys {
				fmt.Println(key)
			}
		}
	}

	time.Sleep(time.Second)

	// Test SCAN
	fmt.Println("\nTesting SCAN:")
	cursor := 0
	for {
		values, err := redis.Values(conn.Do("SCAN", cursor))
		if err != nil {
			fmt.Printf("SCAN failed: %v\n", err)
			break
		}
		cursor, _ = redis.Int(values[0], nil)
		keys, _ := redis.Strings(values[1], nil)
		
		fmt.Printf("SCAN %d\n", cursor)
		if len(keys) == 0 {
			fmt.Println("(empty list or set)")
		} else {
			for _, key := range keys {
				fmt.Println(key)
			}
		}
		
		if cursor == 0 {
			break
		}
		time.Sleep(time.Second)
	}

	fmt.Println("\nSlow Redis compatibility test completed.")
}