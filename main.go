package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SETTINGS
var ADDRESS = "localhost:6969"
var USE_TLS = false

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
			hash      [32]byte
			signature [72]byte // maximum length of an ecdsa signature
		}
	}

	// not PoW so no nonce required, could include hash of signed block though?
}

// need a list of fake historic sources...
// banance, codbase,

type NetworkState struct {
	Blockchain      []Block
	PendingBlock    Block
	LiveSources     []LiveSource     // sources that have to be routinely checked
	HistoricSources []HistoricSource // sources that the "DAO" voted on (we stub it here for now)
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
	// register all HTTP endpoints

	addLiveSource(&ns, "banance")
	addLiveSource(&ns, "codbase")

	fmt.Println(ns)

	// --- give data
	http.HandleFunc("/getlivesources", func(w http.ResponseWriter, req *http.Request) {
		s := jsonOrEmptyArray(ns.LiveSources)
		w.Write(s)
	})

	http.HandleFunc("/getblocks", func(w http.ResponseWriter, req *http.Request) {
		i := min(len(ns.Blockchain)-10, len(ns.Blockchain))
		s := []byte("[]")

		if len(ns.Blockchain) > 0 {
			s = jsonOrEmptyArray(ns.Blockchain[i:])
		} else {
			fmt.Println("Blockchain still size 0")
		}
		w.Write(s)
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

	if USE_TLS {
		http.ListenAndServeTLS(ADDRESS, "certs/cert.pem", "certs/key.pem", nil)
	} else {
		http.ListenAndServe(ADDRESS, nil)
	}
}
