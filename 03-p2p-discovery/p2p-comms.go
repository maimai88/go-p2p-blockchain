package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"os/signal"
    "syscall"

	"github.com/davecgh/go-spew/spew"
	net "github.com/libp2p/go-libp2p-net"

)

func handleStream(s net.Stream) {

	log.Println("Got a new stream!")
	log.Println("New list of peers =", ha.Peerstore().Peers())

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go p2pReadData(rw)
	go p2pWriteData(rw)

	// stream 's' will stay open until you close it (or the other side closes it).

	ch := make(chan os.Signal)
    signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-ch
        log.Println("Received Interrupt. Exiting now.")
        cleanup(rw)
        os.Exit(1)
    }()
}

func p2pReadData(rw *bufio.ReadWriter) {

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			//log.Fatal(err)
			log.Println(err)
		}

		if str == "" {
			return
		}
		if str != "Exit\n" { // Old condition was if str != "\n"

			chain := make([]Block, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}

			mutex.Lock()
			if len(chain) > len(Blockchain) {
				Blockchain = chain
				bytes, err := json.MarshalIndent(Blockchain, "", "  ")
				if err != nil {

					log.Fatal(err)
				}
				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			}
			mutex.Unlock()
		}
	}
}

func p2pWriteData(rw *bufio.ReadWriter) {

	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			bytes, err := json.Marshal(Blockchain)
			if err != nil {
				log.Println(err)
			}
			mutex.Unlock()

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()

		}
	}()

	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if sendData != "\n" {
			sendData = strings.Replace(sendData, "\n", "", -1)
			bpm, err := strconv.Atoi(sendData)
			if err != nil {
				log.Fatal(err)
			}
			newBlock := generateBlock(Blockchain[len(Blockchain)-1], bpm)

			if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
				mutex.Lock()
				Blockchain = append(Blockchain, newBlock)
				mutex.Unlock()
			}

			bytes, err := json.Marshal(Blockchain)
			if err != nil {
				log.Println(err)
			}

			spew.Dump(Blockchain)

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()
		}
	}
}