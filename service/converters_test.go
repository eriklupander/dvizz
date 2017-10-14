package service

import (
	"github.com/docker/docker/api/types/swarm"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSanitizeTaskNameHavingLatestSuffix(t *testing.T) {
	name := sanitizeTaskName("some/name:latest@sha256.1")
	if name != "some/name" {
		t.Error("Expected 'some/name', got " + name)
	}
}

func TestSanitizeTaskNameWithoutSuffix(t *testing.T) {
	name := sanitizeTaskName("some/name.1")
	if name != "some/name.1" {
		t.Error("Expected 'some/name.1', got " + name)
	}
}

func TestConvertTasks(t *testing.T) {
	task := swarm.Task{

		ID:        "1",
		NodeID:    "node-1",
		ServiceID: "service-1",
		Spec: swarm.TaskSpec{
			ContainerSpec: swarm.ContainerSpec{
				Image: "image/name",
			},
		},
		Status: swarm.TaskStatus{
			State: swarm.TaskStateRunning,
		},
		Slot: 2,
	}

	arr := []swarm.Task{}
	arr = append(arr, task)

	tasks := convTasks(arr)
	if tasks[0].Name != "image/name.2" {
		t.Error("Expected task name: 'image/name.2', got: " + tasks[0].Name)
	}
}

func TestConvertNodes(t *testing.T) {
	nodes := make([]swarm.Node, 0)
	nodes = append(nodes, swarm.Node{ID: "id1", Status: swarm.NodeStatus{State: "running"}, Description: swarm.NodeDescription{Hostname: "hostname"}})
	result := convNodes(nodes)

	Convey("Assert", t, func() {
		So(result, ShouldNotBeNil)
		So(len(result), ShouldEqual, 1)
		So(result[0].Name, ShouldEqual, "hostname")
	})

}

func TestConvertTasksEmpty(t *testing.T) {
	tasks := make([]swarm.Task, 0)
	result := convTasks(tasks)
	Convey("Assert", t, func() {
		So(result, ShouldNotBeNil)
		So(len(result), ShouldEqual, 0)
	})
}

func TestConvertServicesEmpty(t *testing.T) {
	services := make([]swarm.Service, 0)
	result := convServices(services)
	Convey("Assert", t, func() {
		So(result, ShouldNotBeNil)
		So(len(result), ShouldEqual, 0)
	})
}

func TestConvertServicesNil(t *testing.T) {
	var services []swarm.Service
	result := convServices(services)
	Convey("Assert", t, func() {
		So(result, ShouldNotBeNil)
		So(len(result), ShouldEqual, 0)
	})
}
