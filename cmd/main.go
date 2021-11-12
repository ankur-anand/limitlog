package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/ankur-anand/limitlog"
)

func main() {
	reader := os.Stdin
	writer := os.Stdout

	db := limitlog.ReaderInWriterOutLog{}
	_, err := db.ReadFrom(reader, writer)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println()
	//PrintMemUsage()
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
