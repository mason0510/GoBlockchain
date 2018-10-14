package main

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
	"net/http"
	"os"
	"log"
	"github.com/gorilla/mux"
	"gopkg.in/gin-gonic/gin.v1/json"
	"io"
	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)

//struct index timestamp hash prehash bpmn
type Block struct {
	Index int
	Timestamp string
	BPM int //jump times
	Hash string
	PreHash string
}

//the whole blockchain
var blockChain []Block

//cal hash value
func calculateHash(block Block)string  {
	record:=string(block.Index)+block.Timestamp+string(block.BPM)+block.PreHash
	h:=sha256.New()
	h.Write([]byte(record))
	hashed:=h.Sum(nil)
	//encode hex
	return hex.EncodeToString(hashed)
}

//generate calhash via last hash
func generateBlock(oldBlock Block,BPM int)(Block,error)  {
	var  newBlock  Block
	t:=time.Now()
	newBlock.Index=oldBlock.Index+1
	newBlock.Timestamp=t.String()
	newBlock.BPM=BPM
	newBlock.PreHash=oldBlock.Hash
	newBlock.Hash=calculateHash(newBlock)
	return newBlock, nil
}

//isBlockValid
func  isBlockValid(newBlock Block,oldBlock Block)bool  {
	if oldBlock.Index+1!=newBlock.Index {
		return false
	}
	if oldBlock.Hash!=newBlock.PreHash {
		return false
	}
	if calculateHash(newBlock)!=newBlock.Hash{
		return false
	}

	return 	true
}

//which is main blockChain ,when it is mainChain
func replaceChain(newBlocks []Block)  {
	if len(newBlocks)>len(blockChain) {
		
	}
}

//web server
func run() error  {
	mux:=makeMuxRouter()
	httpAddr:=os.Getenv("ADDR")
	log.Println("listening on",os.Getenv("ADDR"))
	s:=&http.Server{
		Addr:":"+httpAddr,
		Handler:mux,
		ReadTimeout:10*time.Second,
		WriteTimeout:10*time.Second,
		MaxHeaderBytes:1<<20,

	}
	if error:=s.ListenAndServe();error!=nil {
		return error
	}
	return nil
}
func makeMuxRouter() http.Handler {
	muxRouter:=mux.NewRouter()
	muxRouter.HandleFunc("/",handleGetBlockchain).Methods("Get")
	muxRouter.HandleFunc("/",handleWriteBlock).Methods("post")
	return muxRouter
	}
//get Blockchain handler
func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes,err:=json.MarshalIndent(Blockchain,"","")
	if err!=nil {
		http.Error(w,err.Error(),http.StatusInternalServerError)
	}
	io.WriteString(w,string(bytes))
	}


type Message struct {
	BPM int
}
func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	var m Message
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()

	newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], m.BPM)
	if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, m)
		return
	}
	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
		newBlockchain := append(Blockchain, newBlock)
		replaceChain(newBlockchain)
		spew.Dump(Blockchain)
	}
	respondWithJSON(w, r, http.StatusCreated, newBlock)
}
func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

var Blockchain  []Block
func main() {
	err:=godotenv.Load()
	if err!=nil {
		log.Fatal(err)
	}
	go func() {
		t:=time.Now()
		genesisBlock:=Block{0,t.String(),0,"",""}
		spew.Dump(genesisBlock)
		Blockchain=append(Blockchain,genesisBlock)
		}()
	log.Fatal(run())
}