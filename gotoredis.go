package gotoredis

import (
	"errors"
	"fmt"
	"reflect"

	"code.google.com/p/go-uuid/uuid"
	"github.com/fzzy/radix/redis"
)

type StructMapper struct {
	client *redis.Client
}

func New(redisEndpoint string) (*StructMapper, error) {
	redisClient, err := redis.Dial("tcp", redisEndpoint)
	if err != nil {
		return nil, err
	}

	return &StructMapper{
		client: redisClient,
	}, nil
}

func (mapper StructMapper) Save(obj interface{}) (string, error) {
	id := uuid.New()

	toPersist := reflect.ValueOf(obj)
	structType := toPersist.Type()
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		value := toPersist.FieldByName(field.Name)
		reply := mapper.client.Cmd("HSET", id, field.Name, value)
		insertCount, err := reply.Int()
		if err != nil {
			return "", err
		}
		if insertCount != 1 {
			return "", errors.New(fmt.Sprint("Insert count should have been 1 but was %d", insertCount))
		}
	}

	return id, nil
}

func (mapper StructMapper) Load(id string, obj interface{}) error {
	reply := mapper.client.Cmd("HGETALL", id)
	responseParts, err := reply.Hash()
	if err != nil {
		return err
	}

	pointerToInflate := reflect.ValueOf(obj)
	structToInflate := reflect.Indirect(pointerToInflate)
	typeToInflate := structToInflate.Type()
	for i := 0; i < typeToInflate.NumField(); i++ {
		field := typeToInflate.Field(i)
		value := responseParts[field.Name]
		structToInflate.FieldByName(field.Name).SetString(value)
	}

	return nil
}

func (StructMapper) Close() error {
	return nil
}
