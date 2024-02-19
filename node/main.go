package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"net/http"
	"sync"
	"time"
	"unsafe"
	"encoding/json"
	"github.com/edgelesssys/ego/enclave"
	"crypto/sha256"
)
// EventData represents the structure of each event in the array
type EventData struct {
	Timestamp time.Time `json:"Timestamp"`
	Event     string    `json:"Event"`
}

type ChunkSubmission struct {
	Hash            [32]byte `json:"Hash"`
	Signature       [4616]byte `json:"Signature"`
	Sid             uint64 `json:"Sid"` 
	Until           time.Time `json:"Until"`
	BytesProcessed uint16 `json:"BytesProcessed"`
}

const (
	
	base_url = "http://localhost:6963"
	chunk_submit_url = base_url+"/submitchunk"
	get_live_source_url = base_url + "/getlivesources"
)

var events []EventData

func fetchData(url string, delay int, ch chan<- string) {
	for {
		resp, err := http.Get(url)
		if err != nil {
			ch <- fmt.Sprintf("Error fetching data from %s: %v", url, err)
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			ch <- fmt.Sprintf("Error reading response body from %s: %v", url, err)
			continue
		}

		ch <- string(body)

		// Sleep for the specified delay before making the next request
		time.Sleep(time.Duration(delay) * time.Second)
	}
}

func main() {
	url := flag.String("url", "", "URL to call")
	delay := flag.Int("delay", 0, "Delay in seconds")
	id := flag.Uint64("sourceid",0,"Id of the source")
	flag.Parse()

	if *url == "" || *delay <= 0 {
		fmt.Println("Usage: go run main.go -url <URL> -delay <seconds> -sourceid <ID>")
		return
	}

	fmt.Printf("Will call %s indefinitely with a delay of %d seconds\n", *url, *delay)

	dataChannel := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1) // Increment the WaitGroup counter before starting the goroutine

	// Start a goroutine to fetch data indefinitely with the specified delay
	go func() {
		defer wg.Done() // Decrement the WaitGroup counter when the goroutine completes
		fetchData(*url, *delay, dataChannel)
	}()

	// Process data received from the goroutine
	for data := range dataChannel {
		var event []EventData
		
		err := json.Unmarshal([]byte(data), &event)
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			return
		}
		events = append(events,event...)
		
		sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&events))
		totalSize := unsafe.Sizeof(*sliceHeader) + uintptr(len(events))*unsafe.Sizeof(EventData{})
	
		sizeInKB := float64(totalSize) / (1024)
		fmt.Printf("Size of events is %.2f KB\n",sizeInKB)
		fmt.Printf("Size of events ", len(events))
		if(sizeInKB > 5){
			hash := sha256.New()
			jsonData, err := json.Marshal(events)
			if err != nil {
				fmt.Println("Error encoding JSON:", err)
				return
			}
			hash.Write(jsonData)

			// Get the final SHA-256 hash
			hashInBytes := hash.Sum(nil)
			events = events[:0]
			report, err := enclave.GetRemoteReport(hashInBytes[:])
			var hashArray [32]byte
			copy(hashArray[:], hashInBytes)
			var reportArray [4616]byte
			copy(reportArray[:], report)
			fmt.Println("the hash array is", hashArray)
			submission := ChunkSubmission{
				Hash:            hashArray,        // Replace with actual hash values
				Signature:       reportArray,      // Replace with actual signature values
				Sid:             *id,                      // Replace with actual Sid value
				Until:           time.Now(),   // Replace with actual Until time
				BytesProcessed:  uint16(totalSize),                         // Replace with actual BytesProcessed value
			}
			jsonData, err = json.Marshal(submission)
			if err != nil {
				fmt.Println("Error encoding JSON:", err)
				return
			}
			resp, err := http.Post(chunk_submit_url, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Println("Error making POST request:", err)
				return
			}
			defer resp.Body.Close()
		
			// Check the response status
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Error: Unexpected response status code %d\n", resp.StatusCode)
				return
			}
		
			// Read and print the response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading response body:", err)
				return
			}
		
			jsonData, err = json.Marshal(submission)
			if err != nil {
				fmt.Println("Error encoding JSON:", err)
				return
			}
			fmt.Println(body)
		}
		//fmt.Printf("Data received: %s\n", data)
	}

	// This point will be reached only if the dataChannel is closed
	wg.Wait() // Wait for the fetchData goroutine to complete
}
