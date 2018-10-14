package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)


type Block struct{
	Index int
	Timestamp string
	BPM int
	Hash string
	Prehash string
}

//block is sesies of blocks
var Blockchain []Block
var bcServer chan []Block
var mutex=&sync.Mutex{}

func main() {
	err:=godotenv.Load()
	if err!=nil {
		log.Fatal(err)
	}
	//chan管道
	bcServer=make(chan []Block)
	//create genesis block
	t:=time.Now()
	genesisBlock:=Block{0,t.String(),0,"",""}
	Blockchain=append(Blockchain,genesisBlock)
	httpPort:=os.Getenv("PORT")

	//start tcp
	server,err:=net.Listen("tcp",":"+httpPort)
	if err!=nil {
		log.Fatal(err)
	}
	log.Println("port:",httpPort)
	defer server.Close()

	for   {
		conn,err:=server.Accept()
		if err!=nil {
			log.Fatal(err)
		}
		go handleConn(conn)

	}
}

func handleConn(conn net.Conn)  {
	defer conn.Close()
	io.WriteString(conn,"输入BPM")
	scanner:=bufio.NewScanner(conn)
	//get bpm add blockchain
	go func() {
		for scanner.Scan() {
			bpm, err := strconv.Atoi(scanner.Text())
			if err != nil {
				log.Printf("%v not a number: %v", scanner.Text(), err)
				continue
			}
			newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], bpm)
			if err != nil {
				log.Println(err)
				continue
			}
			if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
				newBlockchain := append(Blockchain, newBlock)
				replaceChain(newBlockchain)
			}

			bcServer <- Blockchain
			io.WriteString(conn, "\nEnter a new BPM:")
		}
	}()
	//receive broadcast
	go func() {
		for   {
			time.Sleep(30*time.Second)
			mutex.Lock()
			output,err:=json.Marshal(Blockchain)
			if err!=nil {
						log.Fatal(err)
			}
			mutex.Unlock()
			io.WriteString(conn,string(output))
	}
	}()
	for _=range bcServer{
		spew.Dump(Blockchain)
	}
}

//is valid blockchain
func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.Prehash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

//ensure longest blockchain
func replaceChain(newBlocks []Block) {
	mutex.Lock()
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
	mutex.Unlock()
}
// SHA256 hasing
func calculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.BPM) + block.Prehash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}


// create a new block using previous block's hash
func generateBlock(oldBlock Block, BPM int) (Block, error) {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.Prehash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
}
