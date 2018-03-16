package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
)

/* Cache Size Exponenet */
const CacheSize_Exp = 15
const CacheSize_Nbr = (1 << CacheSize_Exp)

/* Address Size Exponenet */
const AddressSize_Exp = 32

/* Cache Associativity */
const CacheAssociativity_Exp = 3
const CacheAssociativity = (1 << CacheAssociativity_Exp)

/* Cache Block Size */
const BlockSize_Exp = 6
const BlockSize_Nbr = (1 << BlockSize_Exp)
const BlockSize_Mask = (BlockSize_Nbr - 1)

/* Cache Lines Size */
const Lines_Exp = ((CacheSize_Exp) - (CacheAssociativity_Exp + BlockSize_Exp))
const Lines_Nbr = (1 << Lines_Exp)
const Lines_Mask = (Lines_Nbr - 1)

/* Cache Tag Size */
const Tag_Exp = (AddressSize_Exp - BlockSize_Exp - Lines_Exp)
const Tag_Nbr = (1 << Tag_Exp)
const Tag_Mask = (Tag_Nbr - 1)

/* Program Structure */
const Number_Files = 3

/* Read Address Trace File Variables */
const Address_Bytes = 4
const Address_Block_Bits = 6
const Address_Line_Bits = 6
const Accesses_Max = 67108864

/* Wait Group */
var wg sync.WaitGroup

/* Reports the parameters specified above */
func ReportParam() {
	fmt.Println("-----------------------------------------------------------")
	fmt.Println("Cache Simulator")
	fmt.Println("Based off of Gary J Minden's Cache Simulation in C.")
	fmt.Println("-----------------------------------------------------------")
	fmt.Println("Cache Information")
	fmt.Println("Cache Size: ", CacheSize_Nbr, "KB")
	fmt.Println("Number Lines: ", Lines_Nbr, ", Number Blocks: ", CacheAssociativity,
		", Block Size: ", BlockSize_Nbr, "KB")
	fmt.Println("-----------------------------------------------------------")
	fmt.Println("Address Trace File Information")
	fmt.Println("Address Trace Size: 20 Bits")
	fmt.Println("Address Tag Length: 20 Bits, Address Lines Length: 6 Bits, Address Blocks Length: 6 Bits")
}

func ProcessFile(filename string) {

	/* Local Variables */
	type Block struct {
		BlockValid bool
		Tag        uint32
	}

	type Line struct {
		Block             [CacheAssociativity]Block
		Round_Robin_Count int
	}

	var Cache [Lines_Nbr]Line

	Hit_Count := 0
	Accesses_Nbr := 0

	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("-----------------------------------------------------------")
		fmt.Println("Could not read: ", filename)
		wg.Done()
		return
	}

	/* Flush the Cache */
	for i := 0; i < Lines_Nbr; i++ {
		Cache[i].Round_Robin_Count = 0

		for j := 0; j < CacheAssociativity; j++ {
			Cache[i].Block[j].BlockValid = false
		}
	}

	for Accesses_Nbr < Accesses_Max {
		Accesses_Nbr++

		Address_Trace_Raw := make([]byte, Address_Bytes)
		n, err := f.Read(Address_Trace_Raw)

		Address_Trace := binary.BigEndian.Uint32(Address_Trace_Raw)

		if err != nil || n != Address_Bytes {
			break
		}

		/* Discard the Block */
		Trace_Block := Address_Trace & BlockSize_Mask
		_ = Trace_Block
		Address_Trace = Address_Trace >> Address_Block_Bits

		/* Get the Line Number */
		Trace_Line := Address_Trace & Address_Line_Bits

		/* Get the tag */
		Address_Trace = Address_Trace >> Address_Line_Bits
		Trace_Tag := Address_Trace & Tag_Mask

		/*
		* Check if hit at the trace_line given by the address trace
		* -> Use the trace_line extracted from the read in address and check for each block in the correct
		*    line of cash for hit
		 */
		hit := false

	Hit_Loop:
		for i := 0; i < CacheAssociativity; i++ {
			if Cache[Trace_Line].Block[i].BlockValid && Cache[Trace_Line].Block[i].Tag == Trace_Tag {
				Hit_Count++
				hit = true
				break Hit_Loop
			}
		}

		/*
		* Look for invalid (empty) block
		* -> Go through each block at the trace_line of Cache,
		*    If there is an invalid block, add the item else go on to round robin state
		 */
		if !hit {
			invalid := false

		Valid_Loop:
			for i := 0; i < CacheAssociativity; i++ {
				if !Cache[Trace_Line].Block[i].BlockValid {

					Cache[Trace_Line].Block[i].Tag = Trace_Tag
					Cache[Trace_Line].Block[i].BlockValid = true

					invalid = true
					break Valid_Loop
				}
			}

			/*
			*	If no invalid block, then we must kick out a block
			*		For now, we'll use a common round robin state
			*      Round_robin_count is a a counter from 0 to 7 to indicate
			*      which block is next to kick out
			 */
			if !invalid {
				Round_Robin := Cache[Trace_Line].Round_Robin_Count
				Cache[Trace_Line].Block[Round_Robin].Tag = Trace_Tag
				Cache[Trace_Line].Round_Robin_Count = (Cache[Trace_Line].Round_Robin_Count + 1) % 8
			}
		}
	}

	f.Close()
	ReturnResults(filename, Hit_Count)
}

func ReturnResults(filename string, Hits int) {
	fmt.Println("-----------------------------------------------------------")
	fmt.Println("Processed Results : ", filename)
	fmt.Println("Hits: ", Hits, "Hit Ratio: ", (float32(Hits) / Accesses_Max))
	wg.Done()
}

func main() {

	files := os.Args[1:]

	// files := [Number_Files]string{"AddressTrace_FirstIndex.bin",
	// 	"AddressTrace_LastIndex.bin", "AddressTrace_RandomIndex.bin"}

	if len(files) == 0 {
		fmt.Println("Incorrect usage!")
		fmt.Println("./cache_sim <filename>")
	} else {
		ReportParam()
		for i := 0; i < len(files); i++ {
			wg.Add(1)
			go ProcessFile(files[i])
		}
		wg.Wait()
	}
}
