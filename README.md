# C&C Server for Decentralized Verification experiment

We wrote this server to experiment with some of the distributed concensus ideas without investing much time in implementing a blockchain or P2P architecture from scratch.
It basically emulates the distributed part of the network and the concensus mechanism. 
It's more or less just a backdrop for us to experiment with SGX and different coordination/verification schemes.

## Future improvements
- Stop using JSON, it adds a lot of complexity and is very inefficient space and bandwidth-wise

## Notes for Harsh:
- types.go has all the relevant types you might care about
- All the APIs return JSON, with no adjustment just the raw data. Pretty sure that calling json.Unmarshal with the bytes should give you the same values
- TLS can be disabled if it's causing you problems it's in one of those environment variables
- There are a bunch of settings variables at the top of the main file to tweak the cnc server
- Many apologies, but the cleanest way to get this working was to serve the "live data" as a json array of objects with timestamps. This means you'll have to code the verifiers to discard any data they've already seen. I beg for your forgiveness
    - Not sure if that's clear enough. Basically everytime you fetch data from '/getdata/bananance', you'll get the 10 latest transactions with timestamps no matter how times you call it.
- I haven't tested the adding parts of the code since that would mean implementing a dummy client. Best of luck you're the best
