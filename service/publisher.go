package service

import (
	"encoding/json"
	"github.com/ahl5esoft/golang-underscore"
	"github.com/fsouza/go-dockerclient"
	"time"
	"github.com/eriklupander/dvizz/comms"
        . "github.com/eriklupander/dvizz/model"
        "log"
)

var filters = make(map[string][]string)
var lastNodes []DNode
var eventServer comms.IEventServer

func init() {
	filters["desired-state"] = []string{"running"}
}

func SetEventServer(_eventServer comms.IEventServer) {
        eventServer = _eventServer
        go eventServer.InitializeEventSystem()
        log.Println("Initialized event system. Period.")
}
/**
 * Will poll for Swarm Nodes changes every 5 seconds.
 */
func PublishNodes(client *docker.Client) {
	tmp, _ := client.ListNodes(docker.ListNodesOptions{})
	lastNodes = convNodes(tmp)
	for {
		time.Sleep(time.Second * 5)
		tmp2, _ := client.ListNodes(docker.ListNodesOptions{})
		currentNodes := convNodes(tmp2)
		processNodeListing(currentNodes)
	}
}

// Unit-testable
func processNodeListing(currentNodes []DNode) {
	// Broadcasts stop events for nodes gone missing
	for _, lastNode := range lastNodes {
		isThere := underscore.Any(currentNodes, func(other DObject, _ int) bool {
			return other.Equals(lastNode)
		})
		if !isThere {
			eventServer.AddEventToSendQueue(marshal(DNodeEvent{Action: "stop", Type: "node", Dnode: lastNode}))
		}
	}

	// Broadcasts start events for nodes added
	for _, currentNode := range currentNodes {
		isThere := underscore.Any(lastNodes, func(other DObject, _ int) bool {
			return other.Equals(currentNode)
		})
		if !isThere {
			eventServer.AddEventToSendQueue(marshal(DNodeEvent{Action: "start", Type: "node", Dnode: currentNode}))
		}
	}

	// Broadcast status updates
	for _, currentNode := range currentNodes {
		for _, lastNode := range lastNodes {
			if currentNode.Id == lastNode.Id && currentNode.State != lastNode.State {
				eventServer.AddEventToSendQueue(marshal(DNodeEvent{Action: "update", Type: "node", Dnode: currentNode}))
			}
		}
	}

	lastNodes = currentNodes
}

/**
 * Will poll for Swarm service changes every second.
 */
func PublishServices(client *docker.Client) {
	services, _ := client.ListServices(docker.ListServicesOptions{})
	lastServices := convServices(services)
	for {
		time.Sleep(time.Second * 1)

		tmp, _ := client.ListServices(docker.ListServicesOptions{})

		currentServices := convServices(tmp)

		// First, check if there are any items in lastTasks NOT present in currentTasks. Keep those in temp list
		toDelete := []DService{}
		for _, lastService := range lastServices {
			isThere := underscore.Any(currentServices, func(other DObject, _ int) bool {
				return other.Equals(lastService)
			})
			if !isThere {
				toDelete = append(toDelete, lastService)
			}
		}

		// Then, perform the opposite and populate the toAdd list
		toAdd := []DService{}
		for _, currentService := range currentServices {
			isThere := underscore.Any(lastServices, func(other DObject, _ int) bool {
				return other.Equals(currentService)
			})
			if !isThere {
				toAdd = append(toAdd, currentService)
			}
		}

		// Finally, serialize to JSON and push as events
		go underscore.Each(toAdd, func(item DService, _ int) {
			eventServer.AddEventToSendQueue(marshal(&DServiceEvent{DService: item, Action: "start", Type: "service"}))
		})
		go underscore.Each(toDelete, func(item DService, _ int) {
			eventServer.AddEventToSendQueue(marshal(&DServiceEvent{DService: item, Action: "stop", Type: "service"}))
		})

		lastServices = currentServices // Assign current as last for next iteration.
	}
}

/** Polls for task changes once per second */
func PublishTasks(client *docker.Client) {
	tasks, _ := client.ListTasks(docker.ListTasksOptions{Filters: filters})
	lastTasks := convTasks(tasks)
	for {
		time.Sleep(time.Second * 1)

		tmp, _ := client.ListTasks(docker.ListTasksOptions{Filters: filters})

		currentTasks := convTasks(tmp)

		// First, check if there are any items in lastTasks NOT present in currentTasks. Keep those in temp list
		toDelete := []DTask{}
		for _, lastTask := range lastTasks {
			if !contains(currentTasks, lastTask) {
				toDelete = append(toDelete, lastTask)
			}
		}

		// Then, perform the opposite and populate the toAdd list
		toAdd := []DTask{}
		for _, currentTask := range currentTasks {
			if !contains(lastTasks, currentTask) {
				toAdd = append(toAdd, currentTask)
			}
		}

		// We also want state updates propagated to GUI (desiredState != actual state)
		// Do this by comparing id + state for all
		for _, currentTask := range currentTasks {
			for _, lastTask := range lastTasks {
				if currentTask.Id == lastTask.Id && currentTask.Status != lastTask.Status {
					// We have a status change for a task,
					go func(currentTask DTask) {
						// Wait about .5 second until sending status updates for state changes.
						eventServer.AddEventToSendQueue(marshal(&DTaskStateUpdate{Id: currentTask.Id, State: currentTask.Status, Action: "update", Type: "task"}))
					}(currentTask)
				}
			}
		}

		// Finally, serialize to JSON and push as events
		go underscore.Each(toAdd, func(item DTask, _ int) {
			eventServer.AddEventToSendQueue(marshal(&DEvent{Dtask: item, Action: "start", Type: "task"}))
		})
		go underscore.Each(toDelete, func(item DTask, _ int) {
			eventServer.AddEventToSendQueue(marshal(&DEvent{Dtask: item, Action: "stop", Type: "task"}))
		})

		lastTasks = currentTasks // Assign current as last for next iteration.
	}
}

func marshal(intf interface{}) []byte {
	data, _ := json.Marshal(intf)
	return data
}

func contains(arr []DTask, dstruct Identifier) bool {

	return underscore.Any(arr, func(other DObject, _ int) bool {
		return other.Equals(dstruct)
	})
}
