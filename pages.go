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

type cachedip struct {
	data ipdata
	tc   int64
}

type mapcached struct {
	valid bool
	cache string
}

var ipcache = []cachedip{}
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

	var info = ipinfo(r)
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
		fmt.Println(err2)
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

//var svgre = regexp.MustCompile("(<svg[\\s\\S]+?)<!--(.+?)-->([\\s\\S]+?svg>)") //$1 is head, $2 is circle template, $3 is foot
const circlesize = 2.5

// path /rendermap
func rendermapw(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	fmt.Println("Received request " + r.URL.String())

	w.WriteHeader(200)
	w.Header().Set("content-type", "image/svg+xml")

	var mapstring string
	// already rendered
	if mapcache.valid {
		mapstring = mapcache.cache
	} else {
		// get template
		contentb, err := ioutil.ReadFile("src/maptemplate.svg")
		if err != nil {
			fmt.Println("Error with request:", r)
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
		content := strings.Split(string(contentb), "<!--template-->")

		// pull from database
		alldata := *pullall()

		mapstring = content[0] // header
		for i := 0; i < len(alldata); i++ {
			newcircle := content[1]
			//									   string from uint64 (uint64 from uint16)
			strings.Replace(newcircle, "{ulat}", strconv.FormatUint(uint64(alldata[i].Ulat), 10), 1)
			strings.Replace(newcircle, "{ulon}", strconv.FormatUint(uint64(alldata[i].Ulon), 10), 1)
			strings.Replace(newcircle, "{size}", strconv.FormatFloat(circlesize, 'f', 4, 64), 1)
			mapstring += newcircle
		}
		mapstring += content[2]
	}

	_, err := fmt.Fprint(w, mapstring)
	if err != nil {
		fmt.Println("Error with request:", r)
		fmt.Println(err)
		w.WriteHeader(500)
	}
}

// internal - taken from https://golangcode.com/get-the-request-ip-addr/
func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	if strings.Contains(r.RemoteAddr, "[::1]") {
		return "1.1.1.1" // testing from localhost
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}
