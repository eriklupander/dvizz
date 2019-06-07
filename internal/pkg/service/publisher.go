package service

import (
	"encoding/json"
	underscore "github.com/ahl5esoft/golang-underscore"
	"github.com/eriklupander/dvizz/internal/pkg/comms"
	"github.com/eriklupander/dvizz/internal/pkg/model"
	docker "github.com/fsouza/go-dockerclient"
	"time"
)

//var filters = make(map[string][]string)
//var lastNodes []model.DNode
//var eventServer comms.IEventServer

type Publisher struct {
	filters     map[string][]string
	lastNodes   []model.DNode
	eventServer comms.IEventServer
}

func NewPublisher(eventServer comms.IEventServer) *Publisher {
	f := make(map[string][]string)
	f["desired-state"] = []string{"running"}
	return &Publisher{filters: f, eventServer: eventServer}
}

func (p *Publisher) SetEventServer(_eventServer comms.IEventServer) {
	p.eventServer = _eventServer
	go p.eventServer.InitializeEventSystem()
}

/**
 * Will poll for Swarm Nodes changes every 5 seconds.
 */
func (p *Publisher) PublishNodes(client *docker.Client) {
	tmp, _ := client.ListNodes(docker.ListNodesOptions{})
	p.lastNodes = convNodes(tmp)
	for {
		time.Sleep(time.Second * 5)
		tmp2, _ := client.ListNodes(docker.ListNodesOptions{})
		currentNodes := convNodes(tmp2)
		p.processNodeListing(currentNodes)
	}
}

// Unit-testable
func (p *Publisher) processNodeListing(currentNodes []model.DNode) {
	// Broadcasts stop events for nodes gone missing
	for _, lastNode := range p.lastNodes {
		isThere := underscore.Chain2(currentNodes).Any(func(other model.DObject, _ int) bool {
			return other.Equals(lastNode)
		})
		if !isThere {
			p.eventServer.AddEventToSendQueue(marshal(model.DNodeEvent{Action: "stop", Type: "node", Dnode: lastNode}))
		}
	}

	// Broadcasts start events for nodes added
	for _, currentNode := range currentNodes {
		isThere := underscore.Chain2(p.lastNodes).Any(func(other model.DObject, _ int) bool {
			return other.Equals(currentNode)
		})
		if !isThere {
			p.eventServer.AddEventToSendQueue(marshal(model.DNodeEvent{Action: "start", Type: "node", Dnode: currentNode}))
		}
	}

	// Broadcast status updates
	for _, currentNode := range currentNodes {
		for _, lastNode := range p.lastNodes {
			if currentNode.Id == lastNode.Id && currentNode.State != lastNode.State {
				p.eventServer.AddEventToSendQueue(marshal(model.DNodeEvent{Action: "update", Type: "node", Dnode: currentNode}))
			}
		}
	}

	p.lastNodes = currentNodes
}

/**
 * Will poll for Swarm service changes every second.
 */
func (p *Publisher) PublishServices(client *docker.Client) {
	services, _ := client.ListServices(docker.ListServicesOptions{})
	lastServices := convServices(services)
	for {
		time.Sleep(time.Second * 1)

		tmp, _ := client.ListServices(docker.ListServicesOptions{})

		currentServices := convServices(tmp)

		// First, check if there are any items in lastTasks NOT present in currentTasks. Keep those in temp list
		toDelete := []model.DService{}
		for _, lastService := range lastServices {
			isThere := underscore.Chain2(currentServices).Any(func(other model.DObject, _ int) bool {
				return other.Equals(lastService)
			})
			if !isThere {
				toDelete = append(toDelete, lastService)
			}
		}

		// Then, perform the opposite and populate the toAdd list
		toAdd := []model.DService{}
		for _, currentService := range currentServices {
			isThere := underscore.Chain2(lastServices).Any(func(other model.DObject, _ int) bool {
				return other.Equals(currentService)
			})
			if !isThere {
				toAdd = append(toAdd, currentService)
			}
		}

		// Finally, serialize to JSON and push as events
		go underscore.Chain2(toAdd).Each(func(item model.DService, _ int) {
			p.eventServer.AddEventToSendQueue(marshal(&model.DServiceEvent{DService: item, Action: "start", Type: "service"}))
		})
		go underscore.Chain2(toDelete).Each(func(item model.DService, _ int) {
			p.eventServer.AddEventToSendQueue(marshal(&model.DServiceEvent{DService: item, Action: "stop", Type: "service"}))
		})

		lastServices = currentServices // Assign current as last for next iteration.
	}
}

/** Polls for task changes once per second */
func (p *Publisher) PublishTasks(client *docker.Client) {
	tasks, _ := client.ListTasks(docker.ListTasksOptions{Filters: p.filters})
	lastTasks := convTasks(tasks)
	for {
		time.Sleep(time.Second * 1)

		tmp, _ := client.ListTasks(docker.ListTasksOptions{Filters: p.filters})

		currentTasks := convTasks(tmp)

		// First, check if there are any items in lastTasks NOT present in currentTasks. Keep those in temp list
		toDelete := []model.DTask{}
		for _, lastTask := range lastTasks {
			if !contains(currentTasks, lastTask) {
				toDelete = append(toDelete, lastTask)
			}
		}

		// Then, perform the opposite and populate the toAdd list
		toAdd := []model.DTask{}
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
					go func(currentTask model.DTask) {
						// Wait about .5 second until sending status updates for state changes.
						p.eventServer.AddEventToSendQueue(marshal(&model.DTaskStateUpdate{Id: currentTask.Id, State: currentTask.Status, Action: "update", Type: "task"}))
					}(currentTask)
				}
			}
		}

		// Finally, serialize to JSON and push as events

		go underscore.Chain2(toAdd).Each(func(item model.DTask, _ int) {
			p.eventServer.AddEventToSendQueue(marshal(&model.DEvent{Dtask: item, Action: "start", Type: "task"}))
		})
		go underscore.Chain2(toDelete).Each(func(item model.DTask, _ int) {
			p.eventServer.AddEventToSendQueue(marshal(&model.DEvent{Dtask: item, Action: "stop", Type: "task"}))
		})

		lastTasks = currentTasks // Assign current as last for next iteration.
	}
}

func marshal(intf interface{}) []byte {
	data, _ := json.Marshal(intf)
	return data
}

func contains(arr []model.DTask, dstruct model.Identifier) bool {
	return underscore.Chain2(arr).Any(func(other model.DObject, _ int) bool {
		return other.Equals(dstruct)
	})
}
