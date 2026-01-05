package main

//import packages : install module -  go get github.com/gorilla/websocket
import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// struct json message response model
type JsonResponse struct {
	MsgType string `json:"msg_type"`
	Message string `json:"message"`
	From    string `json:"from"`
	To      string `json:"to"`
}

var All_clients = make(map[string]*websocket.Conn) //map to store the websocket clients

// mutex for avoiding same threads using All_clients at same time (by locking when using and unlocked when finsihed)
var clientsMutex sync.RWMutex

// upgrader for websocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024, //size of buffer -  incominng data (1KB)
	WriteBufferSize: 1024, //size of buffer - outgoing data (1KB)
	CheckOrigin: func(r *http.Request) bool { //like CORS in nodejs
		return true //return true means allowing any user
	},
}

// parse and return from string message as JSOn
func Json_response_maker(message string) (string, string, string, string) {
	//json parse message
	var jsonParsed JsonResponse
	json_parse_error := json.Unmarshal([]byte(message), &jsonParsed)
	if json_parse_error != nil {
		fmt.Println("JSON FAILED TO EXTRACT (wrong json format)") //--- LOG
		return "_", "_", "_", "_"                                 //fake values for error happened
	}
	return jsonParsed.MsgType, jsonParsed.Message, jsonParsed.From, jsonParsed.To
}

// send message to client
func Send_message(ws *websocket.Conn, msg_type string, message string, to string, from string) bool {
	json_string := fmt.Sprintf(`{"msg_type":"%s","message":"%s","from":"%s","to":"%s"}`, msg_type, message, from, to)
	if err := ws.WriteMessage(1, []byte(json_string)); err != nil {
		fmt.Println("SEND MESSAGE FAILED error : ", err) //--- LOG
		return false
	}
	return true //no error
}

// receive message from client forward it to other client
func Receive_and_forward(ws *websocket.Conn, from string) bool {
	_, message, err := ws.ReadMessage() //receive message from client, note :  _ is message type

	//receive message error
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			fmt.Printf("ERROR RECEIVING MESSAGE - error : %v \n", err) //--- LOG
		}
		//break loop to trigger the ws.Close() - close event
		return false
	}

	message_type, message_response, _, to := Json_response_maker(string(message)) // _ is from address

	//for ping requests - skip forwarding
	if message_type == "PING" {
		return true
	}

	//if the json conversion is failed and fake values are returned ("_")
	if to == "_" {
		fmt.Println("JSON FAILED TO EXTRACT DATA FROM STRING")
		return true
	}

	clientsMutex.RLock() //lock the All_clients map for this thread (RLock for map reading)
	targetClient, exists := All_clients[to]
	clientsMutex.RUnlock() //unlock the All_clients map for other threads (RUnlock for map reading)

	if exists {
		Send_message(targetClient, message_type, message_response, to, from)
	} else {
		fmt.Println("DOESNT EXIST USER : ", to)
	}

	return true
}

// manage socket
func HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	connectionID := strings.TrimPrefix(r.URL.Path, "/usr/") //id after the /usr/ of connection url
	ws, err := upgrader.Upgrade(w, r, nil)                  //the connection socket

	//connection error
	if err != nil {
		fmt.Printf("Upgrade Error for %s : \n", connectionID)
		return
	}

	//WRITE
	clientsMutex.Lock()            //lock the All_clients map for this thread
	All_clients[connectionID] = ws //add socket into map
	//new client print
	fmt.Printf("NEW CONNECTION : %s | TOTAL: %d\n", connectionID, len(All_clients))
	clientsMutex.Unlock() //unlock the All_clients map for other threads

	//socket close event
	defer func() {
		ws.Close()

		//WRITE
		clientsMutex.Lock()               //lock the All_clients map for this thread
		delete(All_clients, connectionID) // remove the connection details from map
		fmt.Printf("DISCONNECTED %s | REMAINING: %d\n", connectionID, len(All_clients))
		clientsMutex.Unlock() //unlock the All_clients map for other threads

	}()

	//receive message
	for {

		success := Receive_and_forward(ws, connectionID)
		if success {
			//fmt.Println("forwarded")
		} else {
			//fmt.Println("error while forwarding")
			break
		}

	}
}

// show home page index.html
func HomePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "page/index.html")
}

// show ws test page ws_test.html
func Ws_test(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "page/ws_test.html")
}

// show webrtc+ws test page webrtc_test.html
func WebRtc_test(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "page/webrtc_test.html")
}

// give the count of websocket clients as json
func TotalConnections(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	total_connections := len(All_clients)
	data := map[string]int{"no": total_connections}
	json.NewEncoder(w).Encode(data)
}

func main() {
	port_number := flag.String("p", "8080", "Port number to run server : default - 8080")                      //argparse port number; default value 8080
	enable_website := flag.Bool("w", true, "Enable/Disable the html pages : default - true (website enabled)") //argparse port number; default value 8080 (for low bandwidth mode)
	flag.Parse()

	//if website mode is enabled (-w=true)
	if *enable_website {
		static_files := http.FileServer(http.Dir("./page"))                 //static file directory (adding as static)
		http.Handle("/static/", http.StripPrefix("/static/", static_files)) //make the files in static folder (pages) ready to be accessed
		http.HandleFunc("/", HomePage)                                      //homepage route
		http.HandleFunc("/testws", Ws_test)                                 //ws test page route
		http.HandleFunc("/testwsrtc", WebRtc_test)                          //ws+webrtc test page route
		http.HandleFunc("/count", TotalConnections)                         //get total clietn count route
	}

	http.HandleFunc("/usr/", HandleWebsocket) //handle websocket connection

	fmt.Printf("\nserver started at port : %s \n", *port_number)
	fmt.Printf("Website mode : %v \n", *enable_website)
	http.ListenAndServe(":"+*port_number, nil)
}
