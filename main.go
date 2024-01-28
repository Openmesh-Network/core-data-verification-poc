package main

import (
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SETTINGS
var PORT = "6962"
var ADDRESS = "localhost:" + PORT
var USE_TLS = false
var PRINT_MOCK_EVENTS = false
var PRINT_BLOCK_MINED = true
var MIN_RAND_LATENCY_MS = 0
var MAX_RAND_LATENCY_MS = 0
var EVENT_COOLDOWN = 10           // waiting time between events in ms, keep low
var EVENT_ODDS = .1               // odds an event happens every EVENT_COOLDOWN (1.0 = 100%, .5 = 50%, etc)
var BLOCK_MINE_TIME = 8           // time in seconds in between blocks
var BLOCK_TIME_RANDOM_FACTOR = .1 // randomizes the block lengths, it's (float between 0-1) X this factor in seconds

type NetworkState struct {
	Blockchain      []Block
	PendingBlock    Block
	LiveSources     []LiveSource     // sources that have to be routinely checked
	HistoricSources []HistoricSource // sources that the "DAO" voted on (we stub it here for now)
	MockLiveData    map[uint64][]struct {
		Timestamp time.Time
		Event     string
	}
	lock sync.Mutex
}

func addLiveSource(ns *NetworkState, name string) {
	newSource := LiveSource{}

	newSource.Name = name
	newSource.Id = uint64(len(ns.LiveSources))
	newSource.Url = ADDRESS + "/getdata/" + name

	ns.LiveSources = append(ns.LiveSources, newSource)
}

func main() {
	var ns NetworkState

	ns.MockLiveData = make(map[uint64][]struct {
		Timestamp time.Time
		Event     string
	})

	// register all HTTP endpoints

	addLiveSource(&ns, "bananance")
	addLiveSource(&ns, "codbase")
	addLiveSource(&ns, "bybit")
	addLiveSource(&ns, "dydx")
	addLiveSource(&ns, "kraken")
	addLiveSource(&ns, "bitfinex")

	// --- data for reading
	http.HandleFunc("/getlivesources", func(w http.ResponseWriter, req *http.Request) {
		s := jsonOrEmptyArray(ns.LiveSources)
		w.Write(s)
	})

	http.HandleFunc("/gethistoricsources", func(w http.ResponseWriter, req *http.Request) {
		s := jsonOrEmptyArray(ns.HistoricSources)
		w.Write(s)
	})

	http.HandleFunc("/getblocksrecent", func(w http.ResponseWriter, req *http.Request) {
		i := max(len(ns.Blockchain)-5, 0)

		s := []byte("[]")

		if len(ns.Blockchain) > 0 {
			// only show last 5 blocks
			s = jsonOrEmptyArray(ns.Blockchain[i:])
		} else {
			fmt.Println("Blockchain still size 0")
		}
		w.Write(s)
	})

	http.HandleFunc("/getblocks", func(w http.ResponseWriter, req *http.Request) {
		s := []byte("[]")

		if len(ns.Blockchain) > 0 {
			s = jsonOrEmptyArray(ns.Blockchain)
		} else {
			fmt.Println("Blockchain still size 0")
		}
		w.Write(s)
	})

	http.HandleFunc("/getdata/", func(w http.ResponseWriter, req *http.Request) {
		liveSourceName := strings.Split(req.URL.Path, "/")[2]

		// find live source name in list of sources

		for i := 0; i < len(ns.LiveSources); i++ {
			if strings.Compare(liveSourceName, ns.LiveSources[i].Name) == 0 {
				// fetch live data
				id := ns.LiveSources[i].Id
				stream := ns.MockLiveData[id]

				// send list of 10 latest events!!!

				i := max(0, len(stream)-10)

				if len(stream) > 0 {
					// send json
					// {
					//		"timestamp": time
					//		"event": "Richard -> boshy 100 BTC"
					// }
					fmt.Println("Stream is length: ", len(stream))
					for i = i; i < len(stream); i++ {
						w.Write(jsonOrEmptyObject(stream[i]))
					}
				}
			}
		}

		fmt.Println("getting data")
	})

	// --- read data

	// TODO: Should work, but it's UNTESTED!!!
	http.HandleFunc("/submitchunk", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			req.ParseForm()
			var submission ChunkSubmission
			err := json.NewDecoder(req.Body).Decode(&submission)

			if err != nil {
				fmt.Fprintln(w, "Has to be formatted as ChunkSubmission struct in go file, here's the parsin error if you need it.", err.Error())
			}

			// TODO: add sanity checks (but for now we assume that all nodes are benevolent 😇)

			found := false
			index := 0
			for i := 0; i < len(ns.PendingBlock.Sources); i++ {
				if ns.PendingBlock.Sources[i].Id == submission.Sid {
					index = i
					found = true
				}
			}

			ns.lock.Lock()
			defer ns.lock.Unlock()
			if !found {
				ns.PendingBlock.Sources = append(ns.PendingBlock.Sources, struct {
					Id     uint64
					Chunks []struct {
						BytesProcessed uint16
						Hash           [32]byte
						Signature      [72]byte
					}
				}{Id: submission.Sid, Chunks: nil})
				index = len(ns.PendingBlock.Sources)
				// ns.PendingBlock.Sources[index].Chunks
			}

			ns.PendingBlock.Sources[index].Chunks = append(ns.PendingBlock.Sources[index].Chunks, struct {
				BytesProcessed uint16
				Hash           [32]byte
				Signature      [72]byte
			}{BytesProcessed: submission.BytesProcessed, Hash: submission.Hash, Signature: submission.Signature})
		} else {
			fmt.Fprintln(w, "Only POST is accepted")
		}
	})

	http.HandleFunc("/attest", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			req.ParseForm()
			// add to latest block
			var buf [256]byte
			n, err := req.Body.Read(buf[:])

			if err != nil {
				fmt.Fprintln(w, "Error reading attestation, server got", n, "bytes")
				return
			} else {
				ns.lock.Lock()
				defer ns.lock.Unlock()
				ns.PendingBlock.Attestations = append(ns.PendingBlock.Attestations, buf)
			}
		} else {
			fmt.Fprintln(w, "Only POST is accepted")
		}
	})

	fmt.Println("Started listening on port", PORT)

	// loop to add mock data
	go func() {
		for {
			// randomly add data to each live source
			for i := 0; i < len(ns.LiveSources); i++ {
				id := ns.LiveSources[i].Id

				if rand.Float32() < float32(EVENT_ODDS) {
					stream := ns.MockLiveData[id]

					if len(stream) > 100 {
						// downsize, delete first 10 members and move everything else down
						ns.MockLiveData[id] = append(stream[:0], stream[11:]...)
						stream = ns.MockLiveData[id]
					}

					// generate events randomly
					// A() <sent>|<received> X <to>|<from> B()
					{
						event := ""
						names := []string{"Richard", "Tom", "Harry", "James", "John", "Lindsey", "Luke", "Jesseppe", "Lenny", "Juan", "Bob", "Mike", "Linda", "Annie", "Caesar", "Brutus"}
						actions := []string{"sent", "received from"}
						coins := []string{"BTC", "BTC", "XMR", "XMR", "ETH", "ETH", "ETH", "BTC", "XMR", "🪙", "🍌", "GOLD", "USDT", "USDC", "POLY", "ETH"}

						// 50 / 50 chance of wallet or name

						getName := func() string {
							if rand.Float32() > 0.9 {
								return names[rand.Intn(len(names))]
							} else {
								name := "0x"
								hex := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}
								for i := 0; i < 40; i++ {
									name += hex[rand.Intn(16)]
								}
								return name
							}
						}

						event += getName()
						event += " "
						event += actions[rand.Intn(len(actions))]
						event += " "
						event += getName()
						event += " "
						event += strconv.Itoa(rand.Intn(100))
						event += " "
						event += coins[rand.Intn(len(coins))]

						if PRINT_MOCK_EVENTS {
							fmt.Println(event)
						}

						// only looks complicated bc I used anonymous struct earlier, this just adds a timestamp + event name to an array
						ns.MockLiveData[id] = append(ns.MockLiveData[id], struct {
							Timestamp time.Time
							Event     string
						}{Timestamp: time.Now(), Event: event})
					}
				}
			}

			time.Sleep(time.Duration(EVENT_COOLDOWN) * time.Millisecond)
		}
	}()

	// loop to """"MINE"""" blocks
	go func() {
		for {
			delay := float32(BLOCK_MINE_TIME) + (rand.Float32()*2-1)*float32(BLOCK_TIME_RANDOM_FACTOR)
			time.Sleep(time.Duration(delay) * time.Second)

			ns.lock.Lock()
			ns.PendingBlock.Time = uint64(time.Now().Unix())
			ns.Blockchain = append(ns.Blockchain, ns.PendingBlock)

			if PRINT_BLOCK_MINED {
				fmt.Println("Block \"\"\"mined\"\"\"")
			}

			// Set up next block
			{
				ns.PendingBlock = Block{}

				{ // turns the header to a hash
					hash := sha256.New()
					enc := gob.NewEncoder(hash)
					err := enc.Encode(ns.Blockchain[:len(ns.Blockchain)-1])
					if err != nil {
						fmt.Printf("err: %v\n", err)
					}
					res := hash.Sum(nil)
					for i := 0; i < 32; i++ {
						ns.PendingBlock.HashPrev[i] = res[i]
					}
				}

				ns.PendingBlock.Nonce = rand.Uint64()
			}
			ns.lock.Unlock()
		}
	}()

	if USE_TLS {
		http.ListenAndServeTLS(ADDRESS, "certs/cert.pem", "certs/key.pem", nil)
	} else {
		http.ListenAndServe(ADDRESS, nil)
	}
}
