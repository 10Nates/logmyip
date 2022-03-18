package main

import (
	"context"
	"fmt"
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
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	fmt.Println("Database client connected")
}

func pulldata(ip string) *storeipdata {
	valc := rdb.HGetAll(ctx, ip)
	val, err := valc.Result()
	if err != nil {
		return &storeipdata{e: false}
	}
	ulatp, err := strconv.ParseUint(val["ulat"], 10, 16)
	if err != nil {
		return &storeipdata{e: false}
	}
	ulat := uint16(ulatp)

	ulonp, err := strconv.ParseUint(val["ulon"], 10, 16)
	if err != nil {
		return &storeipdata{e: false}
	}
	ulon := uint16(ulonp)

	ts, err := strconv.ParseInt(val["ts"], 10, 64)
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
	ret := rdb.HSet(ctx, data.ip, "ulat", data.ulat, "ulon", data.ulon, "ts", data.ts)
	res, err := ret.Result()
	if err != nil {
		return false
	}
	if res != 1 {
		return false
	} else {
		return true
	}
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

// utility

func getTS() int64 {
	return time.Now().UnixMilli()
}
