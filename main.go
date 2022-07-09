package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/nitishm/go-rejson"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

var clientId int = 0

func handleConnections(w http.ResponseWriter, r *http.Request) {
	log.Println("handleConnections")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Para cerrar la conexión una vez termina la función
	defer ws.Close()

	// Registramos nuestro nuevo cliente al agregarlo al mapa global de "clients" que fue creado anteriormente.
	clients[ws] = true

	ws.WriteJSON(Message{Username: "Server", Message: "Welcome! Id: " + strconv.Itoa(clientId)})
	clientId++

	broadcast <- Message{
		Username: "Server",
		Message:  "A new user has joined the chat",
	}

	// Bucle infinito que espera continuamente que se escriba  un nuevo mensaje en el WebSocket, lo desserializa de JSON a un objeto Message y luego lo arroja al canal de difusión.
	for {
		var msg Message

		// Read in a new message as JSON and map it to a Message object
		// Si hay un error, registramos ese error y eliminamos ese cliente de nuestro mapa global de clients
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			broadcast <- Message{Username: "system", Message: "User disconnected"}
			break
		}

		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	log.Println("handleRoot")
	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}

func main() {
	cli := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	rh := rejson.NewReJSONHandler()
	rh.SetGoRedisClient(cli)

	testMessage := Message{
		Username: "Server",
		Message:  "Welcome! Id: " + strconv.Itoa(clientId),
	}

	_, _ = rh.JSONSet("test", "$", testMessage)

	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	log.Println("http server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
