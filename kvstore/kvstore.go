package kvstore

import (
	"errors"
	"regexp"
	"sync"
)

type KVStore struct {
	data map[string]string
	mu   sync.RWMutex
}

func New() *KVStore {
	return &KVStore{
		data: make(map[string]string),
	}
}

func (kv *KVStore) Set(key, value string) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.data[key] = value
	return nil
}

func (kv *KVStore) Get(key string) (string, error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	if value, ok := kv.data[key]; ok {
		return value, nil
	}
	return "", errors.New("key not found")
}

func (kv *KVStore) Del(key string) bool {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if _, ok := kv.data[key]; ok {
		delete(kv.data, key)
		return true
	}
	return false
}

func (kv *KVStore) Keys() []string {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	keys := make([]string, 0, len(kv.data))
	for k := range kv.data {
		keys = append(keys, k)
	}
	return keys
}

func (kv *KVStore) Exists(key string) bool {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	_, ok := kv.data[key]
	return ok
}

func (kv *KVStore) Scan(cursor int, match string, count int, keyType string) (int, []string) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	keys := make([]string, 0, len(kv.data))
	for k := range kv.data {
		keys = append(keys, k)
	}

	if cursor >= len(keys) {
		return 0, []string{}
	}

	var matchRegex *regexp.Regexp
	if match != "" {
		matchRegex = regexp.MustCompile(match)
	}

	result := []string{}
	for i := cursor; i < len(keys) && (count == 0 || len(result) < count); i++ {
		key := keys[i]
		if matchRegex != nil && !matchRegex.MatchString(key) {
			continue
		}
		// In this simple KV store, all values are strings, so we'll always match "string" type
		if keyType != "" && keyType != "string" {
			continue
		}
		result = append(result, key)
	}

	nextCursor := cursor + len(result)
	if nextCursor >= len(keys) {
		nextCursor = 0
	}

	return nextCursor, result
}