package gotoredis

import (
// "github.com/fzzy/radix/redis"
)

type StructMapper struct{}

func New(redisEndpoint string) *StructMapper {
	return &StructMapper{}
}

func (StructMapper) Save(obj interface{}) (int, error) {
	return -1, nil
}

func (StructMapper) Load(id int) (interface{}, error) {
	return nil, nil
}

func (StructMapper) Close() error {
	return nil
}
