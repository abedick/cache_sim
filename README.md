# cache_sim

Simulates an 8 way associative 32KB L1 Cache in Google Go. (64 Lines of 8 Blocks @ 64KB ea.)

## Usage 
`go build cache_sim.go`

`./cache_sim <FileName>`

Specify as many files as needed, will run in multithreaded mode.

## Generating Binary files

Binary Files can be generated using hexdump.

### Address Trace format: 
* Tag: 20b
* Line: 6b
* Block: 6b

Example:
`hexdump -n 64 -e '1/4 "%08_ax" " | “ ' -e '4/4 "%08X " "\n"’  AddressTrace_FirstIndex.bin`


Credit: Cache Simulator by Gary J Minden at the University of Kansas. <gminden@ittc.ku.edu>