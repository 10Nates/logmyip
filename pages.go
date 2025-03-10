package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
var numips = -1 // Initalization only number

// path /
func home(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	// path / is a catchall in net/http so this has to be sorted out
	if r.RequestURI != "/" {
		w.WriteHeader(404)
		return
	}

	if numips == -1 { // Only on first visit
		alldata := *pullall()

		numips = len(alldata) - 1 // There is only 1 non-IP item in the dataset
	}

	go countVisits()

	content, err := os.ReadFile("plate/index.html")
	if err != nil {
		fmt.Println("Error with request:", r)
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}
	contentpip := strings.Replace(string(content), "{{userip}}", getIP(r), 1)           // show ip
	contentwnip := strings.Replace(contentpip, "{{numips}}", strconv.Itoa(numips-1), 1) // number of users (-1 because it says "OVER x IPs")
	w.Header().Set("content-type", "text/html")
	w.WriteHeader(200)
	_, err2 := fmt.Fprint(w, contentwnip)
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
func parseLatLon(latstr string, lonstr string) (uint16, uint16, error) {
	//latitude
	latf, err := strconv.ParseFloat(latstr, 64)
	if err != nil {
		fmt.Println(err)
		return 0, 0, err
	}
	ulatf := latf + 90 // -90/90 -> 0/180

	//longitude
	lonf, err := strconv.ParseFloat(lonstr, 64)
	if err != nil {
		fmt.Println(err)
		return 0, 0, err
	}
	ulonf := lonf + 180 // -180/180 -> 0/360

	//convert to uint
	return uint16(ulatf), uint16(ulonf), nil
}

// internal
func ipinfo(r *http.Request) *ipdata {
	userIP := getIP(r)

	// first service - 50K req/month
	res, err := http.Get("https://ipinfo.io/" + userIP + "/loc")
	if err != nil {
		fmt.Println(err)
		return &ipdata{OK: false}
	}
	if res.StatusCode == 200 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return &ipdata{OK: false}
		}
		trimbody := strings.TrimSpace(string(body))
		data := strings.Split(trimbody, ",")
		// something is wrong
		if data[0] == "" || data[1] == "" {
			fmt.Println("Something is wrong ->", res)
			return &ipdata{OK: false}
		}

		ulat, ulon, err := parseLatLon(data[0], data[1])
		if err != nil {
			fmt.Println(err)
			return &ipdata{OK: false}
		}

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

	// second service - 1000 req/day
	res, err = http.Get("https://ipapi.co/" + userIP + "/latlong")
	if err != nil {
		fmt.Println(err)
		return &ipdata{OK: false}
	}
	if res.StatusCode == 200 {
		body, err := io.ReadAll(res.Body)
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

		ulat, ulon, err := parseLatLon(data[0], data[1])
		if err != nil {
			fmt.Println(err)
			return &ipdata{OK: false}
		}

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

	// third service - 45 req/min - no https
	res, err = http.Get("http://ip-api.com/line/" + userIP + "?fields=16576")
	if err != nil {
		fmt.Println(err)
		return &ipdata{OK: false}
	}
	if res.StatusCode == 200 {

		//read content
		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return &ipdata{OK: false}
		}
		data := strings.Split(string(body), "\n")
		if data[0] != "success" {
			fmt.Println("ip-api.com returned error: " + data[0])
			return &ipdata{OK: false}
		}

		ulat, ulon, err := parseLatLon(data[1], data[2])
		if err != nil {
			fmt.Println(err)
			return &ipdata{OK: false}
		}

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

	fmt.Println("All services returned non-200")
	return &ipdata{OK: false}
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

// var svgre = regexp.MustCompile("(<svg[\\s\\S]+?)<!--(.+?)-->([\\s\\S]+?svg>)") //$1 is head, $2 is circle template, $3 is foot
const pointsize = 1.5 // diameter 3

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
		contentb, err := os.ReadFile("plate/maptemplate.svg")
		if err != nil {
			fmt.Println("Error with request:", r)
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
		content := strings.Split(string(contentb), "<!--template-->")

		// pull from database
		alldata := *pullall()

		numips = len(alldata) - 1 // There is only 1 non-IP item in the dataset

		mapstring = content[0] // header
		for _, e := range alldata {
			if !e.OK { // non-IP items in dataset
				continue
			}

			newcircle := content[1]
			//									   string from uint64 (uint64 from uint16)
			newcircle = strings.Replace(newcircle, "{ulat}", strconv.FormatInt(180-int64(e.Ulat), 10), 1) // -1 to center rectangle
			newcircle = strings.Replace(newcircle, "{ulon}", strconv.FormatInt(int64(e.Ulon), 10), 1)     // -1 to center rectangle
			newcircle = strings.Replace(newcircle, "{size}", strconv.FormatFloat(pointsize, 'f', -1, 64), 1)
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
	if strings.Contains(r.RemoteAddr, "192.168.65.1") {
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

// internal
func goodRefer(r *http.Request) bool {
	return strings.HasPrefix(r.Referer(), "https://logmyip.com")
}

// path /unlog (frontend)
func unlogpage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}

	content, err := os.ReadFile("plate/unlog.html")
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
