package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

var cli = redis.NewClient(&redis.Options{
	Addr: getEnv("REDIS_HOST", "localhost") + ":" + getEnv("REDIS_PORT", "6379"),
})
var rh = rejson.NewReJSONHandler()

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
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

	// broadcast <- Message{
	// 	Username: "Server",
	// 	Message:  "A new user has joined the chat",
	// }
	b, _ := json.Marshal(Message{Username: "Server", Message: "Welcome! Id: " + strconv.Itoa(clientId)})
	_ = cli.Publish("chat", b).Err()

	// Bucle infinito que espera continuamente que se escriba  un nuevo mensaje en el WebSocket, lo desserializa de JSON a un objeto Message y luego lo arroja al canal de difusión.
	for {
		var msg Message

		// Read in a new message as JSON and map it to a Message object
		// Si hay un error, registramos ese error y eliminamos ese cliente de nuestro mapa global de clients
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			b, _ = json.Marshal(Message{
				Username: "Server",
				Message:  "A user has left the chat",
			})
			_ = cli.Publish("chat", b).Err()
			break
		}

		// Send the newly received message to the broadcast channel
		b, _ := json.Marshal(msg)
		_ = cli.Publish("chat", b).Err()
	}
}

func handleMessages() {
	pubsub := cli.Subscribe("chat")
	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {
		var message Message
		err := json.Unmarshal([]byte(msg.Payload), &message)
		if err != nil {
			log.Printf("error: %v", err)
		}

		for client := range clients {
			err := client.WriteJSON(message)
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
	port := os.Getenv("PORT")
	rh.SetGoRedisClient(cli)

	testMessage := Message{
		Username: "Server",
		Message:  "Welcome! Id: " + strconv.Itoa(clientId),
	}

	_, _ = rh.JSONSet("test", "$", testMessage)

	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	log.Println("http server started on " + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
