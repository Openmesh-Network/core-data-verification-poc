package main

import "time"

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
			BytesProcessed uint16
			Hash           [32]byte
			Signature      [72]byte // maximum length of an ecdsa signature
		}
	}

	// not PoW so no nonce required, could include hash of signed block though?
	// nonce goes up by 1 every block
	Nonce uint64
	Time  uint64
}

type ChunkSubmission struct {
	// hash
	Hash      [32]byte
	Signature [72]byte
	// source id
	Sid uint64
	// time
	Until time.Time
	// bytes
	BytesProcessed uint16
}
