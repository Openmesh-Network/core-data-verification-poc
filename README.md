# C&C Server for Decentralized Verification experiment

We wrote this server to experiment with some of the distributed concensus ideas without investing much time in implementing a blockchain or P2P architecture from scratch.
It basically emulates the distributed part of the network and the concensus mechanism. 
It's more or less just a backdrop for us to experiment with SGX and different coordination/verification schemes.

## Future improvements
- Stop using JSON, it adds a lot of complexity and is very inefficient space and bandwidth-wise

## Notes for Harsh:
- All the APIs return JSON, with no adjustment just the raw data. Pretty sure that calling json.Unmarshal with the bytes should give you the same values
- Block time is: 
