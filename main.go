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
	"time"
)

// SETTINGS
var ADDRESS = "localhost:6969"
var USE_TLS = false
var MIN_RAND_LATENCY_MS = 0
var MAX_RAND_LATENCY_MS = 0
var EVENT_COOLDOWN = 10 // waiting time between events in ms, keep low
var EVENT_ODDS = .1     // odds an event happens every EVENT_COOLDOWN (1.0 = 100%, .5 = 50%, etc)

type LiveSource struct {
	Name string
	Id   uint64

	// method for getting source (for now just GET)
	Url string
}

type HistoricSource struct {
	Name string
	Id   uint64

	// Content ID for seeding
	Cid [32]byte
}

type Block struct {
	// "header"
	HashPrev [32]byte

	// "body"

	// NOTE(Tom): Hey Harsh add the attestation data here, not sure how big it needs to be
	// estimating like 64bytes + 72 byte signature or something like that
	// I'll reserve some space for it here though
	Attestations [][256]byte

	Sources []struct {
		Id     uint64 // source ID
		Chunks []struct {
			Hash      [32]byte
			Signature [72]byte // maximum length of an ecdsa signature
		}
	}

	// not PoW so no nonce required, could include hash of signed block though?
	// nonce goes up by 1 every block
	nonce uint64
}

// need a list of fake historic sources...
// banance, codbase,

type NetworkState struct {
	Blockchain      []Block
	PendingBlock    Block
	LiveSources     []LiveSource     // sources that have to be routinely checked
	HistoricSources []HistoricSource // sources that the "DAO" voted on (we stub it here for now)
	MockLiveData    map[uint64][]struct {
		Timestamp time.Time
		Event     string
	}
}

func addLiveSource(ns *NetworkState, name string) {
	newSource := LiveSource{}

	newSource.Name = name
	newSource.Id = uint64(len(ns.LiveSources))
	newSource.Url = ADDRESS + "/getdata/" + name

	ns.LiveSources = append(ns.LiveSources, newSource)
}

func jsonOrEmptyArray(v any) []byte {
	j, err := json.Marshal(v)

	if err != nil {
		fmt.Printf("err: %v\n", err)
		j = []byte("[]")
	}

	return j
}

func jsonOrEmptyObject(v any) []byte {
	j, err := json.Marshal(v)

	if err != nil {
		j = []byte("{}")
	}

	return j
}

func main() {

	var ns NetworkState

	ns.MockLiveData = make(map[uint64][]struct {
		Timestamp time.Time
		Event     string
	})

	// register all HTTP endpoints

	addLiveSource(&ns, "banance")
	addLiveSource(&ns, "codbase")

	fmt.Println(ns)

	// --- give data
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
		// randomize some data?
		// might have to track how much data every node in pipe has?
		// gosh darn it!!
		// give out data with timestamps???

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
						w.Write(jsonOrEmptyObject(stream))
					}
				}
			}
		}

		fmt.Println("getting data")
	})

	// --- read data

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
				ns.PendingBlock.Attestations = append(ns.PendingBlock.Attestations, buf)
			}
		} else {
			fmt.Fprintln(w, "Only POST is accepted")
		}
	})

	fmt.Println("Started listening on port 6969...")

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

						fmt.Println(event)

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

			delay := 1 + (rand.Float32()*2 - 1)
			time.Sleep(time.Duration(delay) * time.Second)

			// ns.PendingBlock.HashPrev[0] = 0xde
			// ns.PendingBlock.HashPrev[1] = 0xad
			// ns.PendingBlock.HashPrev[2] = 0xbe
			// ns.PendingBlock.HashPrev[3] = 0xef

			ns.Blockchain = append(ns.Blockchain, ns.PendingBlock)
			fmt.Println("Block \"\"\"mined\"\"\"")

			// Set up next block
			{
				ns.PendingBlock = Block{}

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
				ns.PendingBlock.nonce = rand.Uint64()
			}
		}
	}()

	if USE_TLS {
		http.ListenAndServeTLS(ADDRESS, "certs/cert.pem", "certs/key.pem", nil)
	} else {
		http.ListenAndServe(ADDRESS, nil)
	}
}
