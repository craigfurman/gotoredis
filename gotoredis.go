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
	return id, mapper.persist(id, obj, false)
}

func (mapper StructMapper) Update(id string, obj interface{}) error {
	return mapper.persist(id, obj, true)
}

func (mapper StructMapper) persist(id string, obj interface{}, isUpdate bool) error {
	valueToPersist := reflect.ValueOf(obj)
	structType := valueToPersist.Type()
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := valueToPersist.FieldByName(field.Name)
		err := mapper.insertFieldIntoRedis(id, field.Name, fieldValue, isUpdate)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mapper StructMapper) insertFieldIntoRedis(id, fieldName string, fieldValue reflect.Value, isUpdate bool) error {
	fieldValueAsString, err := convertFieldValueToString(fieldValue)
	if err != nil {
		return err
	}

	reply := mapper.client.Cmd("HSET", id, fieldName, fieldValueAsString)
	insertCount, err := reply.Int()
	if err != nil {
		return err
	}

	var expectedNewRows int = -1
	if isUpdate {
		expectedNewRows = 0
	} else {
		expectedNewRows = 1
	}
	if insertCount != expectedNewRows {
		return errors.New(fmt.Sprintf("Insert count should have been 1 but was %d", insertCount))
	}
	return nil
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

	case reflect.Bool:
		return strconv.FormatBool(value.Bool()), nil

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

func (mapper StructMapper) getHashFromRedis(id string) (map[string]string, error) {
	reply := mapper.client.Cmd("HGETALL", id)
	return reply.Hash()
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
