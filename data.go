package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type storeipdata struct {
	e    bool // exists
	ip   string
	ulat uint16
	ulon uint16
	ts   int64
}

var rdb *redis.Client
var ctx = context.Background()

func initdb() {
	// addr := os.Getenv("RedisAddr")
	// if addr == "" {
	// 	addr = "localhost:6379" // Use default Redis address if not provided
	// }
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("RedisAddr"),
		Password: os.Getenv("RedisPass"),
		DB:       0, // use default DB
	})
	fmt.Println("Database client connected")
}

func pulldata(ip string) *storeipdata {
	valc := rdb.HGetAll(ctx, ip)
	val, err := valc.Result()
	if err != nil {
		return &storeipdata{e: false}
	}
	ulatp, err := strconv.ParseUint(val["ulat"], 36, 16)
	if err != nil {
		return &storeipdata{e: false}
	}
	ulat := uint16(ulatp)

	ulonp, err := strconv.ParseUint(val["ulon"], 36, 16)
	if err != nil {
		return &storeipdata{e: false}
	}
	ulon := uint16(ulonp)

	ts, err := strconv.ParseInt(val["ts"], 36, 64)
	if err != nil {
		return &storeipdata{e: false}
	}

	return &storeipdata{
		e:    true,
		ip:   ip,
		ulat: ulat,
		ulon: ulon,
		ts:   ts,
	}

}

func pullall() *[]ipdata {
	kc := rdb.Keys(ctx, "*")
	k, err := kc.Result()
	if err != nil {
		return &[]ipdata{}
	}
	ret := []ipdata{}
	for i := 0; i < len(k); i++ {
		data := pulldata(k[i])

		ret = append(ret, *storeDataToIPData(data))
	}

	return &ret
}

func storedata(data *storeipdata) bool {
	if !data.e {
		return false
	}

	// convert to strings
	ulat := strconv.FormatUint(uint64(data.ulat), 36)
	ulon := strconv.FormatUint(uint64(data.ulon), 36)
	ts := strconv.FormatInt(data.ts, 36)

	ret := rdb.HSet(ctx, data.ip, map[string]interface{}{"ulat": ulat, "ulon": ulon, "ts": ts})
	_, err := ret.Result()

	return err == nil // return true if no error
}

func removeip(ip string) bool {
	res, err := rdb.Del(ctx, ip).Result()
	if err != nil || res != 1 {
		return false
	}
	return true
}

func storeDataToIPData(storedata *storeipdata) *ipdata {
	return &ipdata{
		OK:   storedata.e,
		IP:   storedata.ip,
		Ulat: storedata.ulat,
		Ulon: storedata.ulon,
	}
}

func IPDatatoStoreData(data *ipdata, timestamp int64) *storeipdata {
	return &storeipdata{
		e:    data.OK,
		ip:   data.IP,
		ulat: data.Ulat,
		ulon: data.Ulon,
		ts:   timestamp,
	}
}

func countVisits() {
	rdb.Incr(ctx, "visits")
}

// utility

func getTS() int64 {
	return time.Now().UnixMilli()
}
