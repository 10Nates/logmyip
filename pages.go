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
	ts   int64
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

	go countVisits()

	content, err := ioutil.ReadFile("plate/index.html")
	if err != nil {
		fmt.Println("Error with request:", r)
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}
	contentpip := strings.Replace(string(content), "{{userip}}", getIP(r), 1) // show ip
	w.Header().Set("content-type", "text/html")
	w.WriteHeader(200)
	_, err2 := fmt.Fprint(w, contentpip)
	if err2 != nil {
		fmt.Println("Error with request:", r)
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}
}

// path /logip
func logip(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}

	err := r.ParseForm()
	if err != nil {
		fmt.Println("Error with request:", r)
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}
	conf := r.FormValue("confirm")
	if conf != "yes" || !goodRefer(r) {
		w.WriteHeader(406)
		return
	}
	data, ts := cachedipinfo(r)
	sdat := IPDatatoStoreData(data, ts)
	success := storedata(sdat)
	if success {
		mapcache = mapcached{valid: false}
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(200)
		fmt.Println("Logged IP - " + data.IP)
		_, err := fmt.Fprint(w, "IP successfully stored")
		if err != nil {
			fmt.Println("Error with request:", r)
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
	} else {
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(500)
		fmt.Println("Error with request:", r)
		fmt.Println("Error storing IP address to database")
		_, err := fmt.Fprint(w, "Error storing IP address to database")
		if err != nil {
			fmt.Println("Error with request:", r)
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
	}
}

// path /ipinfo
func ipinfow(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}

	info, _ := cachedipinfo(r)

	jinfo, err := json.Marshal(info)
	if err != nil {
		fmt.Println("Error with request:", r)
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)
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
		ulatf := latf + 90 // -90/90 -> 0/180

		//longitude
		lonf, err := strconv.ParseFloat(data[1], 64)
		if err != nil {
			fmt.Println(err)
			return &ipdata{OK: false}
		}
		ulonf := lonf + 180 // -180/180 -> 0/360

		//convert to uint
		ulat := uint16(ulatf)
		ulon := uint16(ulonf)

		retdata := &ipdata{
			OK:   true,
			IP:   userIP,
			Ulat: ulat,
			Ulon: ulon,
		}

		ipcache = append(ipcache, cachedip{
			data: *retdata,
			ts:   getTS(),
		})

		return retdata
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
		ulatf := latf + 90 // -90/90 -> 0/180

		//longitude
		lonf, err := strconv.ParseFloat(data[2], 64)
		if err != nil {
			fmt.Println(err)
			return &ipdata{OK: false}
		}
		ulonf := lonf + 180 // -180/180 -> 0/360

		//convert to uint
		ulat := uint16(ulatf)
		ulon := uint16(ulonf)

		retdata := &ipdata{
			OK:   true,
			IP:   userIP,
			Ulat: ulat,
			Ulon: ulon,
		}

		ipcache = append(ipcache, cachedip{
			data: *retdata,
			ts:   getTS(),
		})

		return retdata
	}
}

// internal
func cachedipinfo(r *http.Request) (*ipdata, int64) {
	cache := cachedip{}
	cleanup := false
	for _, e := range ipcache {
		if e.data.IP == getIP(r) {
			if e.ts > (getTS() - 86400000) { // 24 hours as ms
				cache = e
			} else {
				cleanup = true
			}
		}
	}

	if cleanup {
		go deleteOldCache()
	}

	if cache.data.OK {
		return &cache.data, cache.ts
	} else {
		return ipinfo(r), getTS()
	}
}

//var svgre = regexp.MustCompile("(<svg[\\s\\S]+?)<!--(.+?)-->([\\s\\S]+?svg>)") //$1 is head, $2 is circle template, $3 is foot
const circlesize = 2

// path /rendermap.svg
func rendermapw(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}

	var mapstring string
	// already rendered
	if mapcache.valid {
		mapstring = mapcache.cache
	} else {
		// get template
		contentb, err := ioutil.ReadFile("plate/maptemplate.svg")
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
		for _, e := range alldata {
			newcircle := content[1]
			//									   string from uint64 (uint64 from uint16)
			newcircle = strings.Replace(newcircle, "{ulat}", strconv.FormatInt(180-int64(e.Ulat), 10), 1)
			newcircle = strings.Replace(newcircle, "{ulon}", strconv.FormatInt(int64(e.Ulon), 10), 1)
			newcircle = strings.Replace(newcircle, "{size}", strconv.FormatFloat(circlesize, 'f', 4, 64), 1)
			mapstring += newcircle
		}
		mapstring += content[2]

		// save cache & send
		mapcache = mapcached{
			valid: true,
			cache: mapstring,
		}
	}

	w.Header().Set("content-type", "image/svg+xml")
	w.WriteHeader(200)

	_, err := fmt.Fprint(w, mapstring)
	if err != nil {
		fmt.Println("Error with request:", r)
		fmt.Println(err)
		w.WriteHeader(500)
		return
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

// internal
func deleteOldCache() {
	newipcache := []cachedip{}
	for _, e := range ipcache {
		if e.ts < (getTS() - 86400000) { // 24 hours as ms {
			newipcache = append(newipcache, e)
		}
	}
	ipcache = newipcache
}

//internal
func goodRefer(r *http.Request) bool {
	return strings.HasPrefix(r.Referer(), "https://logmyip.com")
}

// path /unlog (frontend)
func unlogpage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}

	content, err := ioutil.ReadFile("plate/unlog.html")
	if err != nil {
		fmt.Println("Error with request:", r)
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}
	userip := getIP(r)

	var isLogged string
	if pulldata(userip).e { // if exists
		isLogged = "yes"
	} else {
		isLogged = "no"
	}

	contentpip := strings.Replace(string(content), "{{userip}}", userip, 1)       // show ip
	contentpip = strings.Replace(string(contentpip), "{{islogged}}", isLogged, 1) // let user know if ip is logged
	w.Header().Set("content-type", "text/html")
	w.WriteHeader(200)
	_, err2 := fmt.Fprint(w, contentpip)
	if err2 != nil {
		fmt.Println("Error with request:", r)
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}
}

// path /unlogip (api)
func unlogip(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}

	err := r.ParseForm()
	if err != nil {
		fmt.Println("Error with request:", r)
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}
	conf := r.FormValue("confirmunlog") // differentiate from logip
	if conf != "yes" || !goodRefer(r) {
		w.WriteHeader(406)
		return
	}
	userip := getIP(r)
	success := removeip(userip)
	if success {
		mapcache = mapcached{valid: false}
		w.WriteHeader(200)
		fmt.Println("Removed IP - " + userip)
		_, err := fmt.Fprint(w, "IP successfully removed")
		if err != nil {
			fmt.Println("Error with request:", r)
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
	} else {
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(500)
		fmt.Println("Error with request:", r)
		fmt.Println("Error removing IP address from database")
		_, err := fmt.Fprint(w, "Error removing IP address from database")
		if err != nil {
			fmt.Println("Error with request:", r)
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
	}
}
