package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"time"
)

func doRequests(url string, request string, resultCh chan<- bool) {
	client := &http.Client{}

	for {
		resp, err := client.Post(url, "application/json", strings.NewReader(request))
		if err != nil {
			resultCh <- false
			fmt.Printf("error: %v\n", err)
		} else if resp.StatusCode != 200 {
			resultCh <- false
			fmt.Printf("status: %v\n", resp.Status)
		} else {
			resultCh <- true
		}

		if resp != nil {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}
	}
}

func main() {
	var url = flag.String("url", "", "Target URL")
	var parallel = flag.Int("parallel", 1, "parallel closed-loop operations")
	var request = flag.String("request", "", "Request content")
	flag.Parse()

	reportPeriod := 1 * time.Second

	results := make(chan bool, *parallel*10)
	for p := 0; p < *parallel; p++ {
		go doRequests(*url, *request, results)
	}

	start := time.Now()
	success := 0
	fail := 0
	expavg := 0.0
	wperiod := 60.0
	for {
		res := <-results
		if res {
			success++
		} else {
			fail++
		}

		elapsed := time.Since(start)
		if elapsed >= reportPeriod {
			alpha := 1 - math.Exp(-elapsed.Seconds()/wperiod)
			expavg = alpha*float64(success) + (1-alpha)*expavg
			fmt.Printf("%.1fs expavg: %.1f success: %.1f fail: %.1f\n",
				wperiod, expavg,
				float64(success)/elapsed.Seconds(), float64(fail)/elapsed.Seconds())
			success = 0
			fail = 0
			start = time.Now()
		}
	}
}
