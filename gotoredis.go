package gotoredis

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"

	"code.google.com/p/go-uuid/uuid"
	"github.com/fzzy/radix/redis"
)

type StructMapper interface {
	Save(obj interface{}) (string, error)
	Load(id string, structPointer interface{}) error
	Update(id string, obj interface{}) error
	Delete(id string) error
	io.Closer
}

type redisStructMapper struct {
	client *redis.Client
}

func New(redisEndpoint string) (StructMapper, error) {
	redisClient, err := redis.Dial("tcp", redisEndpoint)
	if err != nil {
		return nil, err
	}

	return &redisStructMapper{
		client: redisClient,
	}, nil
}

func (mapper redisStructMapper) Save(obj interface{}) (string, error) {
	id := uuid.New()
	return id, mapper.persist(id, obj)
}

func (mapper redisStructMapper) Update(id string, obj interface{}) error {
	return mapper.persist(id, obj)
}

func (mapper redisStructMapper) persist(id string, obj interface{}) error {
	valueToPersist := reflect.ValueOf(obj)
	structType := valueToPersist.Type()

	redisCmdArgs := []string{id}
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := valueToPersist.FieldByName(field.Name)
		fieldValueAsString, err := convertFieldValueToString(fieldValue)
		if err != nil {
			return err
		}
		redisCmdArgs = append(redisCmdArgs, field.Name, fieldValueAsString)
	}

	reply := mapper.client.Cmd("HMSET", redisCmdArgs)
	_, err := reply.Str()
	return err
}

func convertFieldValueToString(value reflect.Value) (string, error) {
	switch value.Kind() {
	case reflect.String:
		return value.String(), nil

	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
		return fmt.Sprintf("%d", value.Int()), nil

	case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint, reflect.Uintptr:
		return fmt.Sprintf("%d", value.Uint()), nil

	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", value.Float()), nil

	case reflect.Complex64, reflect.Complex128:
		return fmt.Sprintf("%f", value.Complex()), nil

	case reflect.Bool:
		return strconv.FormatBool(value.Bool()), nil

	default:
		return "", errors.New("Unsupported Type")
	}
}

func (mapper redisStructMapper) Load(id string, structPointer interface{}) error {
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

func (mapper redisStructMapper) getHashFromRedis(id string) (map[string]string, error) {
	reply := mapper.client.Cmd("HGETALL", id)
	return reply.Hash()
}

func (mapper redisStructMapper) Delete(id string) error {
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

func (mapper redisStructMapper) Close() error {
	return mapper.client.Close()
}
