# Data Verification Proof of Concept

A proof of concept mock-blockchain to experiment with SGX attestations for data verification.

We wrote this to experiment with some of the distributed concensus ideas without investing much time in implementing a blockchain or P2P architecture from scratch.
It basically emulates the distributed part of the network and the concensus mechanism. 

## Dependencies
- Go
- Npm (for frontend visualization)

## Libraries
- [EGo](https://www.edgeless.systems/products/ego/) for SGX attestations + verifications
- React for visualization

## Project Layout
- `frontend` all the react js code for displaying info
- `node` the go program that will use SGX to validate source data

<!--
## How to run
go run .

## Getting up to speed
There is a central HTTP server which receives all the requests from the nodes and marshals them.
This is meant to emulate blockchain consensus.
Given we know all the .

### Central server


### SGX
SGX or Software Guard Extensions, is a cpu-level security scheme that can more or less verify computation (meaning some specific set of instructions was run) through PKC.
The constraint is that 

#### SGX in this project
The validator
!-->

## Future improvements
- Stop using JSON, it adds a lot of complexity and is very inefficient space and bandwidth-wise
    - Switch to protocolbuffer

