package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"encoding/xml"
	"expvar"
	"flag"
	"html/template"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync"
)

var (
	httpAddr = flag.String("http", "127.0.0.1:8080", "address to serve http")
	apiAddr  = flag.String("api", "127.0.0.1:8081", "address to serve api")
)

var (
	expConns      = expvar.NewInt("conns")
	expTotalConns = expvar.NewInt("total_conns")
	expMsgRecv    = expvar.NewInt("msg_recv")
	expMsgSend    = expvar.NewInt("msg_send")
)

func main() {
	runtime.MemProfileRate = 1
	runtime.SetBlockProfileRate(1)
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	http.HandleFunc("/", handleWelcome)
	http.HandleFunc("/chat", handleChat)
	http.Handle("/ws", websocket.Handler(handleWebsocket))
	go http.ListenAndServe(*httpAddr, nil)

	ln, err := net.Listen("tcp", *apiAddr)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				log.Fatal(err)
			}
			go handleApi(c)
		}
	}()

	log.Printf("serving http at %v, api at %v", *httpAddr, *apiAddr)
	select {}
}

type Message struct {
	From string
	Text string
}

type Client interface {
	Send(m Message)
}

var (
	clients      = make(map[Client]bool)
	clientsMutex sync.RWMutex
)

func registerClient(c Client) {
	expConns.Add(1)
	expTotalConns.Add(1)

	clientsMutex.Lock()
	clients[c] = true
	clientsMutex.Unlock()
}

func unregisterClient(c Client) {
	expConns.Add(-1)

	clientsMutex.Lock()
	delete(clients, c)
	clientsMutex.Unlock()
}

func broadcastMessage(m Message) {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()
	for c := range clients {
		c.Send(m)
	}
	expMsgRecv.Add(1)
	expMsgSend.Add(int64(len(clients)))
}

func handleWelcome(w http.ResponseWriter, r *http.Request) {
	log.Printf("serving welcome page")
	welcomePageTempl.Execute(w, nil)
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	log.Printf("serving chat page for '%v'", name)
	type Params struct {
		Name string
	}
	chatPageTempl.Execute(w, &Params{name})
}

type WSClient struct {
	conn *websocket.Conn
	enc  *json.Encoder
}

func (c *WSClient) Send(m Message) {
	c.enc.Encode(m)
}

func handleWebsocket(ws *websocket.Conn) {
	log.Printf("accepted websocket connection")
	c := &WSClient{ws, json.NewEncoder(ws)}
	registerClient(c)
	defer unregisterClient(c)

	dec := json.NewDecoder(ws)
	for {
		var m Message
		if err := dec.Decode(&m); err != nil {
			log.Printf("error reading from websocket: %v", err)
			return
		}
		broadcastMessage(m)
	}
}

type APIClient struct {
	conn net.Conn
	enc  *xml.Encoder
}

func (c *APIClient) Send(m Message) {
	c.enc.Encode(m)
}

func handleApi(conn net.Conn) {
	log.Printf("accepted api connection")
	c := &APIClient{conn, xml.NewEncoder(conn)}
	registerClient(c)
	defer unregisterClient(c)

	dec := xml.NewDecoder(conn)
	for {
		var m Message
		if err := dec.Decode(&m); err != nil {
			log.Printf("error reading from socket: %v", err)
			return
		}
		broadcastMessage(m)
	}
}

var welcomePageTempl = template.Must(template.New("").Parse(`
<html>
	<body>
		<b>Welcome to Devconf chat!</b>
		<form action="/chat">
		Your nickname: <input type="text" name="name" />
		<input class="button" type="submit" value="Go!" />
		</form>
	</body>
</html>
`))

var chatPageTempl = template.Must(template.New("").Parse(`
<html>
	<head>
	<script>
	var websocket = new WebSocket('ws://' + window.location.host + '/ws');
	websocket.onmessage = function(e) {
		var m = JSON.parse(e.data)
		ct = document.getElementById("alltext")
		ct.value += m.From + ": " + m.Text + "\n"
	}
	function sendMessage() {
		ct = document.getElementById("chattext")
		m = {From: "{{.Name}}", Text: ct.value}
		websocket.send(JSON.stringify(m));
		ct.value = ""
		ct.focus()
	}
	</script>
	</head>
	<body>
		<b>Hi, {{.Name}}!</b>
		<form>
		<input type="text" id="chattext"></input>
		<input type="button" value="Say" onclick="sendMessage()"></input>
		</form>
		<textarea readonly=1 rows=20 id="alltext"></textarea>
	</body>
</html>
`))
