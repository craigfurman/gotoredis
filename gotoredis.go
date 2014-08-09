package gotoredis

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

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
		fieldValue := toPersist.FieldByName(field.Name)
		err := mapper.insertFieldIntoRedis(id, field.Name, fieldValue)
		if err != nil {
			return "", err
		}
	}
	return id, nil
}

func (mapper StructMapper) insertFieldIntoRedis(id, fieldName string, fieldValue reflect.Value) error {
	fieldValueAsString, err := convertFieldValueToString(fieldValue)
	if err != nil {
		return err
	}
	reply := mapper.client.Cmd("HSET", id, fieldName, fieldValueAsString)
	insertCount, err := reply.Int()
	if err != nil {
		return err
	}
	if insertCount != 1 {
		return errors.New(fmt.Sprint("Insert count should have been 1 but was %d", insertCount))
	}
	return nil
}

func convertFieldValueToString(value reflect.Value) (string, error) {
	switch value.Kind() {
	case reflect.String:
		return value.String(), nil

	case reflect.Uint64:
		return fmt.Sprintf("%d", value.Uint()), nil

	default:
		return "", errors.New("Unsupported Type")
	}
}

func (mapper StructMapper) Load(id string, structPointer interface{}) error {
	structAsHash, err := mapper.getHashFromRedis(id)
	if err != nil {
		return err
	}

	pointerToInflate := reflect.ValueOf(structPointer)
	structToInflate := reflect.Indirect(pointerToInflate)
	typeToInflate := structToInflate.Type()
	for i := 0; i < typeToInflate.NumField(); i++ {
		field := typeToInflate.Field(i)
		structValue := structToInflate.FieldByName(field.Name)
		valueToSet := structAsHash[field.Name]
		setValueOnStruct(field.Type.Kind(), structValue, valueToSet)
	}

	return nil
}

func setValueOnStruct(kind reflect.Kind, fieldValue reflect.Value, valueToSet string) error {
	switch kind {
	case reflect.String:
		fieldValue.SetString(valueToSet)

	case reflect.Uint64:
		valueAsUint, err := strconv.ParseUint(valueToSet, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetUint(valueAsUint)

	default:
		return errors.New("Unsupported Type")
	}
	return nil
}

func (mapper StructMapper) getHashFromRedis(id string) (map[string]string, error) {
	reply := mapper.client.Cmd("HGETALL", id)
	return reply.Hash()
}

func (mapper StructMapper) Close() error {
	return mapper.client.Close()
}
