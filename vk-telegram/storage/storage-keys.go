package storage

import (
	"vk-telegram/lib/e"

	"github.com/go-redis/redis"
)

type StorageClient struct {
	client *redis.Client
}

func NewRedisClient(host, pw string) *StorageClient {
	return &StorageClient{
		client: redis.NewClient(&redis.Options{
			Addr:     host,
			Password: pw,
			DB:       0,
		}),
	}
}

func (c *StorageClient) AddKey(id, key string) (err error) {
	defer func() { err = e.WrapIfErr("can't add key", err) }()
	err = c.client.Set(id, key, 0).Err()
	return err
}

func (c *StorageClient) GetKey(id string) (res string, err error) {
	defer func() { err = e.WrapIfErr("can't get key", err) }()
	res, err = c.client.Get(id).Result()
	return res, err
}
