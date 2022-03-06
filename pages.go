package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type ipdata struct {
	OK   bool   `json:"ok"`
	IP   string `json:"ip"`
	Ulat uint16 `json:"absolute_latitude"`
	Ulon uint16 `json:"absolute_longitude"`
}

type iplast struct {
	ip string
	tc uint64
}

type mapcached struct {
	valid bool
	cache string
}

var recents = []iplast{}
var mapcache = mapcached{valid: false}

// path /
func home(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	fmt.Println("Received request " + r.URL.String())

	content, err := ioutil.ReadFile("src/index.html")
	if err != nil {
		fmt.Println("Error with request:", r)
		fmt.Println(err)
	}
	w.WriteHeader(200)
	w.Header().Set("content-type", "text/html")
	_, err2 := w.Write(content)
	if err2 != nil {
		fmt.Println("Error with request:", r)
		fmt.Println(err)
	}
}

// path /logip
func logip(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	fmt.Println("Received request " + r.URL.String())
}

// path /ipinfo
func ipinfow(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	fmt.Println("Received request " + r.URL.String())

	info := ipinfo(r)
	jinfo, err := json.Marshal(info)
	if err != nil {
		fmt.Println("Error with request:", r)
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Header().Set("content-type", "application/json")
	_, err2 := w.Write(jinfo)
	if err2 != nil {
		fmt.Println("Error with request:", r)
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}
}

// internal
func ipinfo(r *http.Request) *ipdata {
	userIP := getIP(r)

	// first service - 45 req/min
	res, err := http.Get("http://ip-api.com/line/" + userIP + "?fields=16576")
	if err != nil {
		fmt.Println(err)
		return &ipdata{OK: false}
	}

	// too many requests
	if res.StatusCode != 200 {
		// second service - 1000 req/day
		res2, err := http.Get("https://ipapi.co/" + userIP + "/latlong")
		if err != nil {
			fmt.Println(err)
			return &ipdata{OK: false}
		}
		if res2.StatusCode != 200 {
			fmt.Println("Both services returned non-200")
			return &ipdata{OK: false}
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return &ipdata{OK: false}
		}
		data := strings.Split(string(body), ",")
		// something is wrong
		if data[0] == "" || data[1] == "" {
			fmt.Println("Something is wrong ->", res)
			return &ipdata{OK: false}
		}

		//latitude
		latf, err := strconv.ParseFloat(data[0], 64)
		if err != nil {
			fmt.Println(err)
			return &ipdata{OK: false}
		}
		ulatf := latf + 180 // -180/180 -> 0/360

		//longitude
		lonf, err := strconv.ParseFloat(data[1], 64)
		if err != nil {
			fmt.Println(err)
			return &ipdata{OK: false}
		}
		ulonf := lonf + 90 // -90/90 -> 0/180

		//convert to uint
		ulat := uint16(ulatf + 180)
		ulon := uint16(ulonf + 180)

		return &ipdata{
			OK:   true,
			IP:   userIP,
			Ulat: ulat,
			Ulon: ulon,
		}
	} else {

		//read content
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return &ipdata{OK: false}
		}
		data := strings.Split(string(body), "\n")
		if data[0] != "success" {
			fmt.Println("ip-api.com returned error: " + data[0])
			return &ipdata{OK: false}
		}

		//latitude
		latf, err := strconv.ParseFloat(data[1], 64)
		if err != nil {
			fmt.Println(err)
			return &ipdata{OK: false}
		}
		ulatf := latf + 180 // -180/180 -> 0/360

		//longitude
		lonf, err := strconv.ParseFloat(data[2], 64)
		if err != nil {
			fmt.Println(err)
			return &ipdata{OK: false}
		}
		ulonf := lonf + 90 // -90/90 -> 0/180

		//convert to uint
		ulat := uint16(ulatf + 180)
		ulon := uint16(ulonf + 180)

		return &ipdata{
			OK:   true,
			IP:   userIP,
			Ulat: ulat,
			Ulon: ulon,
		}
	}
}

// path /rendermap
func rendermapw(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	fmt.Println("Received request " + r.URL.String())
}

// internal - taken from https://golangcode.com/get-the-request-ip-addr/
func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}
