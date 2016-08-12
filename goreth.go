// Package goreth initializes a connection to Redis instance
// to store verification codes and other temp authentication data
package goreth

import (
	//	"fmt"
	"time"
	"errors"
	"strconv"
	"gopkg.in/redis.v4"
)

// Client is a Redis client representing a pool of zero or more underlying connections. It's safe for concurrent use by multiple goroutines.
var pool *redis.Client

func InitPool(host string,
	port string,
	db int,
	pwd string) error {
	
	// NewClient returns a client to the Redis Server
	pool = redis.NewClient(&redis.Options{
		Addr:     host + ":" + port, // "reds:6379",
		DB:       db,
		Password: pwd,
		PoolSize: 10,  // default
	})

	_, err := pool.Ping().Result()
	
	return err
}

var (
	ErrLgnExists error = errors.New("LgnExists: please retry after expiration")
	ErrLgnInvalid error = errors.New("LgnInvalid: min 3")
	ErrVcodeInvalid error = errors.New("VcodeInvalid: min 3")
	ErrSecondsInvalid error = errors.New("SecondsInvalid: min 5")
	ErrLgnNotFound error = errors.New("LgnNotFound: no such key in db")
)

// SetLoginAndVcode: insert a key with expiration time
// If exists already - error
func SetLoginAndVcode(lgn string,
	vcode string,
	seconds time.Duration) (
		error) {

	// validation on code level (no db level in Redis)
	if len(lgn) < 3 { return ErrLgnInvalid }

	if len(vcode) < 3 { return ErrVcodeInvalid	}

	if seconds < 1 { return ErrSecondsInvalid }

	//fmt.Println("saving", lgn, vcode, seconds)
	// NX = not exists
	// http://redis.io/commands/hsetnx
	// *StatusCmd = http://redis.io/topics/protocol#simple-string-reply
	// return OK or other string
	//isSaved, err := pool.SetNX(lgn, vcode, seconds * time.Second).Result()
	isSaved, err := pool.HSetNX(lgn, "vcode", vcode).Result()

	// 127.0.0.1:6379> HGETALL mylogin
	// 1) "vcode"
	// 2) "85060"

	// 127.0.0.1:6379> HGET mylogin "vcode"
	// "41298"	
	
	if err != nil {	return err }

	if isSaved == false {
		// a "lgn" field already exists
		// a key exists only if a "lgn" field exists
		return ErrLgnExists
	}

	// Zero expiration means the key has no expiration time
	_, err = pool.Expire(lgn, seconds * time.Second).Result()

	return err
}

// Add retry to check verification code
// http://redis.io/commands/HINCRBY
// By default HINCRBY:
// - If key does not exist, a new key holding a hash is created.
// - If field does not exist the value is set to 0 before the operation is performed.
// Using EVAL:
// - Verify key existence
// - Increments only if a key exists
func AddRetry(lgn string) (error) {

	scr := `
local exists = redis.call('EXISTS', KEYS[1]);
if (exists == 1) then
return redis.call('HINCRBY', KEYS[1], 'retry', 1);
else 
return 0;
end;
exists = nil;`
	
	r, err := pool.Eval(scr, []string{lgn}).Result()

	// *IntCmd
	// the value at field after the increment operation
	// for a first try: 0 + 1 = 1
	// 2nd try: "2"
	// 3rd try: "3"
	//_, err := pool.HIncrBy(lgn, "retry", 1).Result()
	
	if err != nil { return err }

	if r.(int64) == 0 {
		return ErrLgnNotFound
	}

	return nil
}

// GetVcode by email/phone
// Returns empty string if not found or empty string
func GetVcode(lgn string) (string, int, error) {
	vcode := ""
	retry := 0

	// *SliceCmd
	// []interface{}
	arr, err := pool.HMGet(lgn, "vcode", "retry").Result()

	// 127.0.0.1:6379> HMGET asdf "vcode" "retry"
	// 1) "41576"
	// 2) (nil)

	if err != nil{
		return vcode, retry, err
		// if err == redis.Nil {
		// 	// return empty string			
		// 	return "", nil

		// 	// need to check later: can not be nil or empty string
		// }

		// return "", err
	}

	// For every field that does not exist in the hash, a nil value is returned. Because a non-existing keys are treated as empty hashes, running HMGET against a non-existing key will return a list of nil values

	if arr[0] == nil {
		return vcode, retry, ErrLgnNotFound
	}
	
	vcode = arr[0].(string)

	var retryStr string
	if arr[1] == nil {
		retryStr = "0"
	} else {
		retryStr = arr[1].(string)
	}

	retry, errConv := strconv.Atoi(retryStr)

	if errConv != nil { return vcode, retry, errConv }
	
	//fmt.Println("vcode", vcode, retry)
	
	return vcode, retry, nil
}

// Del deletes a one key from a storage
// No error if a record not exists
func DelKey(key string) (error){
	// *IntCmd
	// contains filtered or unexported fields
	_, err := pool.Del(key).Result()

	if err != nil {
		return err
	}

	// result: The number of keys that were removed
	//if result != 1 {
	// no items removed
	// return errors.New("No items removed")
	//}

	return nil
}
