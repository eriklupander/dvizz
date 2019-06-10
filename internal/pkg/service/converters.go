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
	"fmt"
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
	return DNode{Id: node.ID, State: string(node.Status.State), Name: node.Description.Hostname, CPUs: toCPU(node.Description.Resources.NanoCPUs), Memory: toMemory(node.Description.Resources.MemoryBytes)}
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
		networks := make([]DNetwork, len(task.NetworksAttachments))
		for idx, na := range task.NetworksAttachments {
			networks[idx] = DNetwork{Id: na.Network.ID, Name: na.Network.Spec.Name}
		}

		return DTask{
			Id:        task.ID,
			Name:      sanitizeTaskName(task.Spec.ContainerSpec.Image) + "." + strconv.Itoa(task.Slot),
			Status:    string(task.Status.State),
			ServiceId: task.ServiceID,
			NodeId:    task.NodeID,
			Networks:  networks,
		}
	})
	dtasks, _ := u.([]DTask)
	return dtasks
}

func sanitizeTaskName(name string) string {
	index := strings.Index(name, ":latest")
	if index > -1 {
		name = name[:index]
	}

	// Remove everything before any leading slash
	index = strings.Index(name, "/")
	if index > -1 && index != len(name)-1 {
		name = name[index+1:]
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

func toCPU(c int64) string {
	return fmt.Sprintf("%d CPU(s)", int(c/1000000000))
}

func toMemory(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
