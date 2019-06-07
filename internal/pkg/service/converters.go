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
package service

import (
	"github.com/ahl5esoft/golang-underscore"
	"github.com/docker/docker/api/types/swarm"
	. "github.com/eriklupander/dvizz/internal/pkg/model"
	"strconv"
	"strings"
)

func convNodes(nodes []swarm.Node) []DNode {
	if nodes == nil || len(nodes) == 0 {
		return make([]DNode, 0)
	}
	return underscore.Map(nodes, toDNode).([]DNode)
}

func toDNode(node swarm.Node, _ int) DNode {
	return DNode{Id: node.ID, State: string(node.Status.State), Name: node.Description.Hostname}
}

func convTasks(tasks []swarm.Task) []DTask {
	if tasks == nil || len(tasks) == 0 {
		return make([]DTask, 0)
	}
	dst := make([]swarm.Task, 0)
	underscore.Chain2(tasks).Filter(func(task swarm.Task, _ int) bool {
		// Make sure we only include items that has a nodeId assigned
		return task.NodeID != ""
	}).Value(&dst)

	u := underscore.Map(dst, func(task swarm.Task, _ int) DTask {
		return DTask{
			Id:        task.ID,
			Name:      sanitizeTaskName(task.Spec.ContainerSpec.Image) + "." + strconv.Itoa(task.Slot),
			Status:    string(task.Status.State),
			ServiceId: task.ServiceID,
			NodeId:    task.NodeID,
		}
	})
	dtasks, _ := u.([]DTask)
	return dtasks
}

func sanitizeTaskName(name string) string {
	index := strings.Index(name, ":latest")
	if index > -1 {
		return name[:index]
	}
	return name
}

func convServices(services []swarm.Service) []DService {
	if services == nil || len(services) == 0 {
		return make([]DService, 0)
	}
	u := underscore.Map(services, func(service swarm.Service, _ int) DService {
		return DService{
			Id:   service.ID,
			Name: service.Spec.Name,
		}
	})
	return u.([]DService)
}
