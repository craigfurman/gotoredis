package gotoredis

type StructMapper struct{}

func New(redisEndpoint string) *StructMapper {
	return &StructMapper{}
}

func (StructMapper) Save(obj interface{}) error {
	return nil
}
