package repository

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/vmihailenco/msgpack/v5"
)

type Redis struct {
	Client *redis.Client
	Ctx    context.Context
}

func NewRedisClient(addr string, password string, db int) *Redis {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &Redis{
		Client: client,
		Ctx:    context.Background(),
	}
}

func (r *Redis) Exists(key string) (bool, error) {
	result, err := r.Client.Exists(r.Ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (r *Redis) Set(key string, value string, expiration time.Duration) error {
	return r.Client.Set(r.Ctx, key, value, expiration).Err()
}

func (r *Redis) Get(key string) (string, error) {
	value, err := r.Client.Get(r.Ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return value, err
}

func (r *Redis) SetJSON(key string, data interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.Client.Set(r.Ctx, key, jsonData, expiration).Err()
}

func (r *Redis) GetJSON(key string, dest interface{}) error {
	value, err := r.Client.Get(r.Ctx, key).Result()
	if err == redis.Nil {
		return nil
	} else if err != nil {
		return err
	}
	return json.Unmarshal([]byte(value), dest)
}

func (r *Redis) Invalidate(key string) error {
	return r.Client.Del(r.Ctx, key).Err()
}

func (r *Redis) InvalidPrefix(prefix string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := r.Client.Scan(r.Ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return fmt.Errorf("error scanning keys: %w", err)
		}

		if len(keys) > 0 {
			if err := r.Client.Del(r.Ctx, keys...).Err(); err != nil {
				return fmt.Errorf("error deleting keys: %w", err)
			}
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}
	return nil
}

// external package
func (r *Redis) SetMsgPack(key string, data interface{}, expiration time.Duration) error {
	msgData, err := msgpack.Marshal(data)
	if err != nil {
		return err
	}
	return r.Client.Set(r.Ctx, key, msgData, expiration).Err()
}

// external package
func (r *Redis) GetMsgPack(key string, dest interface{}) error {
	value, err := r.Client.Get(r.Ctx, key).Bytes()
	if err == redis.Nil {
		return nil
	} else if err != nil {
		return err
	}
	return msgpack.Unmarshal(value, dest)
}

// Only for golang
func (r *Redis) SetGob(key string, data interface{}, expiration time.Duration) error {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(data); err != nil {
		return err
	}
	return r.Client.Set(r.Ctx, key, buffer.Bytes(), expiration).Err()
}

// Only for golang
func (r *Redis) GetGob(key string, dest interface{}) error {
	value, err := r.Client.Get(r.Ctx, key).Bytes()
	if err == redis.Nil {
		return nil
	} else if err != nil {
		return err
	}

	buffer := bytes.NewBuffer(value)
	dec := gob.NewDecoder(buffer)
	return dec.Decode(dest)
}

/*
JSON: 24MB (100.000 records)
MessagePack: 14MB (100.000 records)
Gob (Go-native): 10MB (100.000 records)

Which One Should You Use?
Format	Pros	Cons	Use Case
JSON	Human-readable, easy debugging	Larger size, slower	General usage
MessagePack	Small size, fast serialization	Needs external package	API caching, fast processing
Gob	Go-native, very efficient for structs	Not cross-language compatible	Internal Go-only data
🚀 Recommendation: Use MessagePack for speed and efficiency. If you only work within Go, Gob is a great alternative.
*/
