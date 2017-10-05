package comms

import (
        "github.com/gorilla/websocket"
        "github.com/fsouza/go-dockerclient"
        "fmt"
        "net/http"
        "encoding/json"
        "time"
        "log"
        "strconv"
)

type IEventServer interface {
        AddEventToSendQueue(data []byte)
        InitializeEventSystem()
}

type EventServer struct {
        upgrader           websocket.Upgrader
        // Create unbuffered channel
        eventQueue         chan []byte
        // Pointer to docker client
        Client             *docker.Client
        // Web Socket connection registry (in case we have > 1 dashboards driven by this backend)
        connectionRegistry []*websocket.Conn
}

func (server *EventServer) init() {
        server.upgrader = websocket.Upgrader{} // use default options
        server.eventQueue = make(chan []byte)
        server.connectionRegistry = make([]*websocket.Conn, 0, 10)
}

func (server *EventServer) AddEventToSendQueue(data []byte) {
        server.eventQueue <- data
}

func (server *EventServer) InitializeEventSystem() {              // dclient *docker.Client
       // server.client = dclient               // TODO remove
        if server.Client == nil {
                panic("Cannot initialize event server, Docker client not assigned.")
        }

        fmt.Println("Starting WebSocket server at port 6969")

        http.HandleFunc("/start", server.registerChannel)
        http.HandleFunc("/nodes", server.getNodes)
        http.HandleFunc("/services", server.getServices)
        http.HandleFunc("/tasks", server.getTasks)
        http.HandleFunc("/containers", server.getContainers)
        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
                http.ServeFile(w, r, "static/" + r.URL.Path[1:])
        })
        go server.startEventSender()

        fmt.Println("Starting WebSocket server")
        err := http.ListenAndServe(":6969", nil)
        if err != nil {
                panic("ListenAndServe: " + err.Error())
        }
}

func (server *EventServer) getNodes(w http.ResponseWriter, r *http.Request) {
        nodes, err := server.Client.ListNodes(docker.ListNodesOptions{})
        if err != nil {
                panic(err)
        }
        json, _ := json.Marshal(&nodes)
        writeResponse(w, json)
}

func (server *EventServer) getServices(w http.ResponseWriter, r *http.Request) {
        services, err := server.Client.ListServices(docker.ListServicesOptions{})
        if err != nil {
                panic(err)
        }
        json, _ := json.Marshal(&services)
        writeResponse(w, json)
}

func (server *EventServer) getTasks(w http.ResponseWriter, r *http.Request) {
        tasks, err := server.Client.ListTasks(docker.ListTasksOptions{})
        if err != nil {
                panic(err)
        }
        json, _ := json.Marshal(&tasks)
        writeResponse(w, json)
}

func (server *EventServer) getContainers(w http.ResponseWriter, r *http.Request) {
        containers, err := server.Client.ListContainers(docker.ListContainersOptions{All: false})
        if err != nil {
                panic(err)
        }
        json, _ := json.Marshal(&containers)
        writeResponse(w, json)
}

func (server *EventServer) startEventSender() {
        fmt.Println("Starting event sender goroutine...")
        for {
                data := <-server.eventQueue
                server.broadcastDEvent(data)
                time.Sleep(time.Millisecond * 50)
        }
}

func (server *EventServer) broadcastDEvent(data []byte) {
        for index, wsConn := range server.connectionRegistry {
                err := wsConn.WriteMessage(1, data)
                if err != nil {
                        // Detected disconnected channel. Need to clean up.
                        fmt.Printf("Could not write to channel: %v", err)
                        wsConn.Close()
                        server.remove(index)
                }
        }
}

func (server *EventServer) remove(item int) {
        server.connectionRegistry = append(server.connectionRegistry[:item], server.connectionRegistry[item + 1:]...)
}

func (server *EventServer) registerChannel(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/start" {
                http.Error(w, "Not found", 404)
                return
        }
        if r.Method != "GET" {
                http.Error(w, "Method not allowed", 405)
                return
        }
        header := make(map[string][]string)

        header["Access-Control-Allow-Origin"] = []string{"*"}
        c, err := server.upgrader.Upgrade(w, r, header)
        if err != nil {
                log.Print("upgrade:", err)
                return
        }
        server.connectionRegistry = append(server.connectionRegistry, c)
}

func writeResponse(w http.ResponseWriter, json []byte) {
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("Content-Length", strconv.Itoa(len(json)))
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.WriteHeader(http.StatusOK)
        w.Write(json)
}

