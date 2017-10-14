package service

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	. "github.com/eriklupander/dvizz/model"
	"fmt"
	"github.com/eriklupander/dvizz/comms"
	"testing"
)

// mocks for boltdb and messsaging

func TestProcessOneNodeAdded(t *testing.T) {
	mockEventServer := setupMock()

	Convey("Given", t, func() {
		// Start state, start with two nodes.
		lastNodes = buildDNodes([]string{"node1", "node2"})

		Convey("When", func() {
			nextNodes := buildDNodes([]string{"node1", "node2", "node3"})
			processNodeListing(nextNodes)
			Convey("Then", func() {
				So(len(lastNodes), ShouldEqual, 3)
				So(mockEventServer.AssertNumberOfCalls(t, "AddEventToSendQueue", 1), ShouldBeTrue)
			})
		})
	})
}

func TestProcessOneNodeRemoved(t *testing.T) {
	mockEventServer := setupMock()

	Convey("Given", t, func() {
		// Start state, start with two nodes.
		lastNodes = buildDNodes([]string{"node1", "node2"})
		Convey("When", func() {
			nextNodes := buildDNodes([]string{"node2"})
			processNodeListing(nextNodes)
			Convey("Then", func() {
				So(len(lastNodes), ShouldEqual, 1)
				So(mockEventServer.AssertNumberOfCalls(t, "AddEventToSendQueue", 1), ShouldBeTrue)
			})
		})
	})
}

func TestProcessOneNodeRemovedTwoAdded(t *testing.T) {
	mockEventServer := setupMock()

	Convey("Given", t, func() {

		// Start state, start with two nodes.
		lastNodes = buildDNodes([]string{"node1", "node2"})
		Convey("When", func() {
			nextNodes := buildDNodes([]string{"node2", "node3", "node4"})
			processNodeListing(nextNodes)
			Convey("Then", func() {
				So(len(lastNodes), ShouldEqual, 3)
				So(mockEventServer.AssertNumberOfCalls(t, "AddEventToSendQueue", 3), ShouldBeTrue)
			})
		})
	})
}

func buildDNodes(ids []string) []DNode {
	nodes := make([]DNode, 0)
	fmt.Printf("Before iterating %v nodes.\n", len(nodes))
	for index, id := range ids {
		fmt.Printf("Iterating %v\n", index)
		nodes = append(nodes, buildDNode(id))
	}
	fmt.Printf("Returning %v nodes.\n", len(nodes))
	return nodes
}
func buildDNode(nodeId string) DNode {
	return DNode{Id: nodeId, Name: nodeId + "-name", State: "running"}
}

func setupMock() *comms.MockEventServer {
	mockEventServer := &comms.MockEventServer{}
	mockEventServer.Mock.On("AddEventToSendQueue", mock.AnythingOfType("[]uint8"))
	eventServer = mockEventServer
	return mockEventServer
}
