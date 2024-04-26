package env

import "os"

var RedisPassWord = os.Getenv("REDIS_PASSWORD")
var MongoPassWord = os.Getenv("MONGO_PASSWORD")
