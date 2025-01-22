package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)
var clientConns = make(map[*websocket.Conn]string)
type Message struct {
	MessageType int
	Data []byte
}
type Signal struct {
	Symbol string `json:"symbol"`;
	Side string `json:"side"`;
	Strategy string `json:"strategy"`
	Exchange string `json:"exchange"`;
}
func broadcastSignal (clientConns map[*websocket.Conn]string, signal Signal) {
	for conn, _ := range clientConns{
		s, err := json.Marshal(signal)
		if err != nil{
			fmt.Println("Failed Json: ", err)
		}		
		conn.WriteMessage(1,s)
	}
}
func WSHandler (w http.ResponseWriter, r *http.Request) {
	// Upgrade incoming GET request into a Websocket connection
	upgrader := websocket.Upgrader{
		ReadBufferSize: 1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrate connection:", err)
		
	}
	if _, ok := clientConns[conn]; !ok{
		clientConns[conn] = conn.RemoteAddr().String()
		log.Printf("Connections Map: %v", clientConns)
	}

	
	// Close ws connection & unregister the client when they disconnect
	defer conn.Close()
	defer func() {
		delete(clientConns, conn)
		log.Println("Client disconnected!")
	}()

	// Register the new client to the symbol they're subscribing to
	for {
		_, message, err := conn.ReadMessage()
		
		log.Printf("Connections Map: %v", clientConns)
		log.Println(string(message))
		if err != nil {
			log.Println("Error reading from the client:", err)
			break
		}
		
	}
}
func main(){
	godotenv.Load(".env")
		
	fmt.Println("Running server on port: ", os.Getenv("PORT"))
	
	http.HandleFunc("/ws", WSHandler)
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Received a GET request\n"))
	})
	http.HandleFunc("/signal", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST"{
			s := new(Signal)
			if err := json.NewDecoder(r.Body).Decode(s); err != nil{
					fmt.Println(err)
			}
			fmt.Println(s)
			broadcastSignal(clientConns, *s)
			w.Write([]byte("Received a POST request\n"))

		} else{
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
		}
	})	
	http.ListenAndServe(":"+os.Getenv("PORT"),nil)
}


