package main

import (
	"context"
	// "errors"
	"fmt"
	// "os"
	// "time"
	"github.com/go-redis/redis/v8"
	"runtime"
	// "reflect"
)

func main() {

	opt := &redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	rdb := redis.NewClient(opt)
	fmt.Println(opt.MinIdleConns)
	// testScript := redis.NewScript(`
	// if redis.call("hexists", "finish", KEYS[1]) == 1  then	return 0 end return redis.call("hset","finish",KEYS[1],ARGV[1])
	// `)
	// n, err := testScript.Run(context.Background(), rdb, []string{"234"}, "ma").Result()
	// fmt.Println(n, err)
	// testScript := redis.NewScript(`
	// local infinish = redis.call("hexists", "finish", KEYS[1])
	// local infailed = redis.call("hexists", "failed", KEYS[1])
	// return {infinish, infailed}
	// `)
	// n, err := testScript.Run(context.Background(), rdb, []string{"ma"}).Result()
	// tmp := n.([]interface{})
	// fmt.Println(tmp[0].(int64))
	// fmt.Println(tmp[1].(int64))
	// fmt.Println(tmp)
	// fmt.Println(reflect.TypeOf(n))
	// fmt.Println(n, err)
	// fmt.Println()
	// testScript := redis.NewScript(`
	// redis.call("hsetnx", "finish", KEYS[1], ARGV[1])
	// redis.call("hdel", "failed", KEYS[1])
	// return 1
	// `)
	// n, err := testScript.Run(context.Background(), rdb, []string{"mapei"}, "yanis").Result()
	// fmt.Println(n, err)

	redisScript := redis.NewScript(`
	if redis.call("hexists", "finish", KEYS[1]) == 0 then
	redis.call("hsetnx", "failed", KEYS[1],ARGV[1])
	end
	return 1
	`)
	n, err := redisScript.Run(context.Background(), rdb, []string{"ma2"}, "yanis").Result()
	fmt.Println(n, err)
	fmt.Println(runtime.NumCPU())

}
