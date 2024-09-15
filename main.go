package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

// what the response from the Plausible API looks like
type PlausibleResponse struct {
	Results struct {
		Pageviews struct {
			Value int `json:"value"`
		} `json:"pageviews"`
	} `json:"results"`
}

// what the response from the proxy server looks like
type ProxyResponse struct {
	Pageviews int    `json:"pageviews"`
	Page      string `json:"page"`
	Cached    bool   `json:"cached,omitempty"`
}

func main() {

	apiKey := os.Getenv("PLAUSIBLE_API_KEY")
	siteID := os.Getenv("PLAUSIBLE_SITE_ID")
	plausibleURL := os.Getenv("PLAUSIBLE_URL")
	statPeriod := "6mo"
	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "7000"
	}

	if apiKey == "" {
		log.Fatalf("PLAUSIBLE_API_KEY is not set")
	}

	if siteID == "" {
		log.Fatalf("PLAUSIBLE_SITE_ID is not set")
	}

	if plausibleURL == "" {
		log.Fatalf("PLAUSIBLE_URL is not set")
	}

	// create an in memory cache to prevent overwhelming the Plausible API
	// 10 seconds TTL and 60 seconds cleanup interval
	cache := cache.New(10*time.Second, 60*time.Second)

	client := &http.Client{Timeout: 10 * time.Second}

	http.HandleFunc("/pageviews", func(w http.ResponseWriter, r *http.Request) {

		startTime := time.Now()

		page := r.URL.Query().Get("page")
		if page == "" {
			http.Error(w, "page query parameter is required", http.StatusBadRequest)
			return
		}
		if page != "/" && strings.HasSuffix(page, "/") {
			// strip trailing slash
			page = page[:len(page)-1]
		}

		// first check the cache
		if x, found := cache.Get(page); found {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			json.NewEncoder(w).Encode(ProxyResponse{
				Pageviews: x.(int),
				Page:      page,
				Cached:    true,
			})
			return
		}

		req, err := http.NewRequest("GET", plausibleURL+"/api/v1/stats/aggregate", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		q := req.URL.Query()
		q.Add("site_id", siteID)
		q.Add("period", statPeriod)
		q.Add("metrics", "pageviews")
		statFilter := "event:page==" + page + "/"
		q.Add("filters", statFilter)
		req.URL.RawQuery = q.Encode()
		req.Header.Add("Authorization", "Bearer "+apiKey)

		// Make the request to the Plausible API
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, "unexpected status code from Plausible API: "+resp.Status, http.StatusInternalServerError)
			return
		}

		// Decode the response from the Plausible API
		var plausibleResponse PlausibleResponse
		if err := json.NewDecoder(resp.Body).Decode(&plausibleResponse); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// cache the response
		cache.Set(page, plausibleResponse.Results.Pageviews.Value, 0)

		// send the response back to the client
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		json.NewEncoder(w).Encode(ProxyResponse{
			Pageviews: plausibleResponse.Results.Pageviews.Value,
			Page:      page,
		})

		clientIP := r.RemoteAddr
		method := r.Method
		uri := r.RequestURI
		elapsedTime := time.Since(startTime)
		log.Printf("%s - %s %s %s %s", time.Now().Format("2006-01-02 15:04:05"), clientIP, method, uri, elapsedTime)
	})
	// start the server
	log.Printf("Server listening on port %s", serverPort)
	log.Fatal(http.ListenAndServe(":"+serverPort, nil))
}
