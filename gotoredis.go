package gotoredis

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"

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

func (mapper StructMapper) Save(key string, obj interface{}) error {
	structToPersist := reflect.ValueOf(obj)
	structType := structToPersist.Type()
	objectHash := toHash(structType, structToPersist)
	return mapper.saveInRedis(key, objectHash)
}

func (mapper StructMapper) saveInRedis(key string, objectHash map[string]interface{}) error {
	redisCmdArgs := toRedisHmsetArgs(key, objectHash)
	reply := mapper.client.Cmd("HMSET", redisCmdArgs)
	_, err := reply.Str()
	return err
}

func toHash(typeToPersist reflect.Type, valueToPersist reflect.Value) map[string]interface{} {
	objectHash := make(map[string]interface{})
	for i := 0; i < typeToPersist.NumField(); i++ {
		fieldName := typeToPersist.Field(i).Name
		fieldValue := valueToPersist.FieldByName(fieldName).Interface()
		objectHash[fieldName] = fieldValue
	}
	return objectHash
}

func toRedisHmsetArgs(key string, hash map[string]interface{}) []interface{} {
	redisCmdArgs := []interface{}{key}
	for key, value := range hash {
		redisCmdArgs = append(redisCmdArgs, key, value)
	}
	return redisCmdArgs
}

func (mapper StructMapper) Load(id string, structPointer interface{}) error {
	structAsHash, err := mapper.getHashFromRedis(id)
	if err != nil {
		return err
	}

	pointerToFill := reflect.ValueOf(structPointer)
	toFill := reflect.Indirect(pointerToFill)
	structType := toFill.Type()
	loadFields(structType, toFill, structAsHash)

	return nil
}

func loadFields(structType reflect.Type, toFill reflect.Value, structAsHash map[string]string) {
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldToFill := toFill.FieldByName(field.Name)
		fieldValue := structAsHash[field.Name]
		setValueOnStruct(field.Type.Kind(), fieldToFill, fieldValue)
	}
}

func setValueOnStruct(kind reflect.Kind, fieldValue reflect.Value, valueToSet string) error {
	switch kind {
	case reflect.String:
		fieldValue.SetString(valueToSet)

	case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint, reflect.Uintptr:
		valueAsUint, err := strconv.ParseUint(valueToSet, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetUint(valueAsUint)

	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
		valueAsInt, err := strconv.ParseInt(valueToSet, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetInt(valueAsInt)

	case reflect.Float32, reflect.Float64:
		valueAsFloat, err := strconv.ParseFloat(valueToSet, 64)
		if err != nil {
			return err
		}
		fieldValue.SetFloat(valueAsFloat)

	case reflect.Complex64, reflect.Complex128:
		regex, err := regexp.Compile("[\\d\\.]+")
		if err != nil {
			return err
		}
		components := regex.FindAllString(valueToSet, -1)
		if len(components) != 2 {
			return errors.New(fmt.Sprintf("%s is not a complex number", valueToSet))
		}
		r, err := strconv.ParseFloat(components[0], 64)
		if err != nil {
			return err
		}
		i, err := strconv.ParseFloat(components[1], 64)
		if err != nil {
			return err
		}
		fieldValue.SetComplex(complex(r, i))

	case reflect.Bool:
		boolValue, err := strconv.ParseBool(valueToSet)
		if err != nil {
			return err
		}
		fieldValue.SetBool(boolValue)

	default:
		return errors.New("Unsupported Type")
	}
	return nil
}

func (mapper StructMapper) Delete(id string) error {
	reply := mapper.client.Cmd("DEL", id)
	valuesDeleted, err := reply.Int()
	if err != nil {
		return err
	}
	if valuesDeleted != 1 {
		return errors.New(fmt.Sprintf("Delete count should have been 1 but was %d", valuesDeleted))
	}
	return nil
}

func (mapper StructMapper) Close() error {
	return mapper.client.Close()
}

func (mapper StructMapper) getHashFromRedis(id string) (map[string]string, error) {
	reply := mapper.client.Cmd("HGETALL", id)
	hash, err := reply.Hash()
	if err != nil {
		return nil, err
	}
	if len(hash) < 1 {
		return nil, errors.New(fmt.Sprintf("No Redis hash found for key %s", id))
	}
	return hash, nil
}
