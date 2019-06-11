package comms

import (
	"encoding/json"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"syscall"
	"time"
)

type IEventServer interface {
	AddEventToSendQueue(data []byte)
	InitializeEventSystem()
	Close()
}

type EventServer struct {
	upgrader websocket.Upgrader
	// Create unbuffered channel
	eventQueue chan []byte
	// Pointer to docker client
	Client *docker.Client
	// Web Socket connection registry (in case we have > 1 dashboards driven by this backend)
	connectionRegistry []*websocket.Conn
}

func (server *EventServer) init() {
	server.upgrader = websocket.Upgrader{} // use default options
	server.connectionRegistry = make([]*websocket.Conn, 0)
}

func (server *EventServer) AddEventToSendQueue(data []byte) {
	server.eventQueue <- data
}

func (server *EventServer) InitializeEventSystem() {
	if server.Client == nil {
		panic("Cannot initialize event server, Docker client not assigned.")
	}

	logrus.Info("Starting WebSocket server at port 6969")

	http.HandleFunc("/start", server.registerChannel)
	http.HandleFunc("/nodes", server.getNodes)
	http.HandleFunc("/services", server.getServices)
	http.HandleFunc("/tasks", server.getTasks)
	http.HandleFunc("/networks", server.getNetworks)
	http.HandleFunc("/networkreport", server.getNetworkReport)
	http.HandleFunc("/servicereport", server.getServiceReport)
	http.HandleFunc("/containers", server.getContainers)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/"+r.URL.Path[1:])
	})

	server.eventQueue = make(chan []byte, 100)
	go server.startEventSender()
	go server.pinger()

	handleSigterm(func() {
		server.Close()
	})

	logrus.Info("Starting WebSocket server")
	err := http.ListenAndServe(":6969", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func (server *EventServer) Close() {
	deletes := make([]int, 0)
	for index, wsConn := range server.connectionRegistry {
		adr := wsConn.RemoteAddr().String()
		wsConn.Close()
		deletes = append(deletes, index)
		logrus.Info("Gracefully shut down websocket connection to " + adr)
	}
	for _, deleteMe := range deletes {
		server.connectionRegistry = remove(server.connectionRegistry, deleteMe)
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

func (server *EventServer) getNetworks(w http.ResponseWriter, r *http.Request) {
	opts := docker.NetworkFilterOpts{
		"scope": map[string]bool{
			"swarm": true,
		},
	}
	networks, err := server.Client.FilteredListNetworks(opts)
	if err != nil {
		panic(err)
	}
	json, _ := json.Marshal(&networks)
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

func (server *EventServer) getNetworkReport(w http.ResponseWriter, r *http.Request) {

	opts := docker.NetworkFilterOpts{
		"scope": map[string]bool{
			"swarm": true,
		},
	}
	networks, err := server.Client.FilteredListNetworks(opts)
	if err != nil {
		panic(err)
	}

	services, err := server.Client.ListServices(docker.ListServicesOptions{})
	if err != nil {
		panic(err)
	}

	sm := make(map[string]string)
	for _, s := range services {
		sm[s.ID] = s.Spec.Name
	}

	report := make([]NetworkReport, 0)

	for _, n := range networks {
		nr := NetworkReport{ID: n.ID, Name: n.Name, Services: make([]string, 0)}
		for _, s := range services {
			for _, na := range s.Spec.TaskTemplate.Networks {
				if na.Target == n.ID {
					nr.Services = append(nr.Services, sm[s.ID])
				}
			}
		}
		report = append(report, nr)
	}

	data, _ := json.Marshal(report)

	writeResponse(w, data)
}

func (server *EventServer) getServiceReport(w http.ResponseWriter, r *http.Request) {

	opts := docker.NetworkFilterOpts{
		"scope": map[string]bool{
			"swarm": true,
		},
	}
	networks, err := server.Client.FilteredListNetworks(opts)
	if err != nil {
		panic(err)
	}

	services, err := server.Client.ListServices(docker.ListServicesOptions{})
	if err != nil {
		panic(err)
	}

	sm := make(map[string]string)
	for _, s := range services {
		sm[s.ID] = s.Spec.Name
	}

	report := make([]NetworkReport, 0)

	for _, n := range networks {
		nr := NetworkReport{ID: n.ID, Name: n.Name, Services: make([]string, 0)}
		for _, s := range services {
			for _, na := range s.Spec.TaskTemplate.Networks {
				if na.Target == n.ID {
					nr.Services = append(nr.Services, sm[s.ID])
				}
			}
		}
		report = append(report, nr)
	}

	sReport := make([]ServiceReport, 0)

	for _, s := range services {
		sr := ServiceReport{Name: s.Spec.Name, Networks: make([]*NetworkReport, 0)}

		for _, na := range s.Spec.TaskTemplate.Networks {
			sr.Networks = append(sr.Networks, find(na.Target, report))
		}
		sReport = append(sReport, sr)
	}

	data, _ := json.Marshal(sReport)

	writeResponse(w, data)
}

func find(networkId string, reports []NetworkReport) *NetworkReport {
	for _, nr := range reports {
		if nr.ID == networkId {
			return &nr
		}
	}
	return nil
}

type ServiceReport struct {
	Name     string
	Networks []*NetworkReport
}

type NetworkReport struct {
	ID       string
	Name     string
	Services []string
}

func (server *EventServer) startEventSender() {
	logrus.Infof("Starting event sender goroutine...")
	for {
		data := <-server.eventQueue
		logrus.Debugf("About to send event: " + string(data))
		server.broadcastDEvent(data)
		time.Sleep(time.Millisecond * 50)
	}
}

func (server *EventServer) pinger() {
	for {
		time.Sleep(time.Second * 5)
		deletes := make([]int, 0)
		for index, wsConn := range server.connectionRegistry {
			err := wsConn.WriteMessage(1, []byte("PING"))
			if err != nil {
				// Detected disconnected channel. Need to clean up.
				err := wsConn.Close()
				if err != nil {
					logrus.Warnf("problem closing connection: %v", err)
				}
				deletes = append(deletes, index)
			}
		}

		// to mitigate problems with indicies not being updated when deleting multiple entries, we sort deleteMe
		// DESC so the highest is deleted first.
		sort.Slice(deletes, func(i, j int) bool {
			return deletes[i] > deletes[j]
		})

		for _, deleteMe := range deletes {
			server.connectionRegistry = remove(server.connectionRegistry, deleteMe)
			logrus.Infof("Removed stale connection, new count is %v", len(server.connectionRegistry))
		}
	}
}

func (server *EventServer) broadcastDEvent(data []byte) {
	deletes := make([]int, 0)
	for index, wsConn := range server.connectionRegistry {
		err := wsConn.WriteMessage(1, data)
		if err != nil {
			// Detected disconnected channel. Need to clean up.
			logrus.Errorf("Could not write to channel: %v", err)
			wsConn.Close()
			deletes = append(deletes, index)
		}
	}

	// to mitigate problems with indicies not being updated when deleting multiple entries, we sort deleteMe
	// DESC so the highest is deleted first.
	sort.Slice(deletes, func(i, j int) bool {
		return deletes[i] > deletes[j]
	})

	for _, deleteMe := range deletes {
		server.connectionRegistry = remove(server.connectionRegistry, deleteMe)
	}
}

func remove(s []*websocket.Conn, i int) []*websocket.Conn {
	s[i] = s[len(s)-1]
	// We do not need to put s[i] at the end, as it will be discarded anyway
	return s[:len(s)-1]
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
		logrus.Errorf("upgrade: %v", err)
		return
	}
	server.connectionRegistry = append(server.connectionRegistry, c)
	logrus.Infof("A new subscriber connected from %v. Current number of subscribers are: %v", c.RemoteAddr().String(), len(server.connectionRegistry))
}

func writeResponse(w http.ResponseWriter, json []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(json)))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

// Handles Ctrl+C or most other means of "controlled" shutdown gracefully. Invokes the supplied func before exiting.
func handleSigterm(handleExit func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		handleExit()
		os.Exit(1)
	}()
}
