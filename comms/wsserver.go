/**
The MIT License (MIT)

Copyright (c) 2016 ErikL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package comms

import (
	"encoding/json"
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"time"
)

var upgrader = websocket.Upgrader{} // use default options

// Create unbuffered channel
var eventQueue = make(chan []byte)

// Pointer to docker client
var client *docker.Client

// Web Socket connection registry (in case we have > 1 dashboards driven by this backend)
var connectionRegistry = make([]*websocket.Conn, 0, 10)


func AddEventToSendQueue(data []byte) {
      eventQueue <- data
}

func startEventSender() {
	fmt.Println("Starting event sender goroutine...")
	for {
		data := <- eventQueue
		broadcastDEvent(data)
		time.Sleep(time.Millisecond * 50)
	}
}

func broadcastDEvent(data []byte) {
	for index, wsConn := range connectionRegistry {
		err := wsConn.WriteMessage(1, data)
		if err != nil {
			// Detected disconnected channel. Need to clean up.
			fmt.Printf("Could not write to channel: %v", err)
			wsConn.Close()
			remove(index)
		}
	}
}


func remove(item int) {
	connectionRegistry = append(connectionRegistry[:item], connectionRegistry[item+1:]...)
}

func registerChannel(w http.ResponseWriter, r *http.Request) {
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
	c, err := upgrader.Upgrade(w, r, header)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	connectionRegistry = append(connectionRegistry, c)

}

func InitializeEventSystem(dclient *docker.Client) {
	client = dclient

	fmt.Println("Starting WebSocket server at port 6969")

	http.HandleFunc("/start", registerChannel)
	http.HandleFunc("/nodes", getNodes)
	http.HandleFunc("/services", getServices)
	http.HandleFunc("/tasks", getTasks)
	http.HandleFunc("/containers", getContainers)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/"+r.URL.Path[1:])
	})
	go startEventSender()

	fmt.Println("Starting WebSocket server")
	err := http.ListenAndServe(":6969", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func getNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := client.ListNodes(docker.ListNodesOptions{})
	if err != nil {
		panic(err)
	}
	json, _ := json.Marshal(&nodes)
	writeResponse(w, json)
}

func getServices(w http.ResponseWriter, r *http.Request) {
	services, err := client.ListServices(docker.ListServicesOptions{})
	if err != nil {
		panic(err)
	}
	json, _ := json.Marshal(&services)
	writeResponse(w, json)
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := client.ListTasks(docker.ListTasksOptions{})
	if err != nil {
		panic(err)
	}
	json, _ := json.Marshal(&tasks)
	writeResponse(w, json)
}

func getContainers(w http.ResponseWriter, r *http.Request) {
	containers, err := client.ListContainers(docker.ListContainersOptions{All: false})
	if err != nil {
		panic(err)
	}
	json, _ := json.Marshal(&containers)
	writeResponse(w, json)
}

func writeResponse(w http.ResponseWriter, json []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(json)))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}
