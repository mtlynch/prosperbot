package redis

type RedisQuitter interface {
	Quit() (string, error)
}

type RedisSetNXer interface {
	SetNX(key string, value interface{}) (bool, error)
}

type RedisSetter interface {
	Set(key string, value interface{}) (string, error)
}

type RedisGetterSetter interface {
	Get(string) (string, error)
	Set(key string, value interface{}) (string, error)
}

type RedisListPrepender interface {
	LRange(key string, start int64, stop int64) ([]string, error)
	LPush(key string, values ...interface{}) (int64, error)
}
