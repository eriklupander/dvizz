package main

import (
	"github.com/docker/docker/api/types/swarm"
	"strconv"
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

func TestConvertTasksEmpty(t *testing.T) {
	tasks := make([]swarm.Task, 0)
	result := convTasks(tasks)
	if result == nil {
		t.Error("Expected non-nill result")
	}
	if len(result) != 0 {
		t.Error("Expecte 0 length result, got: " + strconv.Itoa(len(result)))
	}
}

func TestConvertServicesEmpty(t *testing.T) {
	services := make([]swarm.Service, 0)
	result := convServices(services)
	if result == nil {
		t.Error("Expected non-nill result")
	}
	if len(result) != 0 {
		t.Error("Expecte 0 length result, got: " + strconv.Itoa(len(result)))
	}
}

func TestConvertServicesNil(t *testing.T) {
	var services []swarm.Service
	result := convServices(services)
	if result == nil {
		t.Error("Expected non-nill result")
	}
	if len(result) != 0 {
		t.Error("Expecte 0 length result, got: " + strconv.Itoa(len(result)))
	}
}
