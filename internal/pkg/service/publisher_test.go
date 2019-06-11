package service

import (
	"fmt"
	"github.com/eriklupander/dvizz/cmd"
	"github.com/eriklupander/dvizz/internal/pkg/comms/mock_comms"
	. "github.com/eriklupander/dvizz/internal/pkg/model"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestProcessOneNodeAdded(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEventServer := mock_comms.NewMockIEventServer(ctrl)
	mockEventServer.EXPECT().AddEventToSendQueue(gomock.Any()).Times(1)

	p := NewPublisher(mockEventServer, cmd.DefaultConfiguration())

	Convey("Given", t, func() {
		// Start state, start with two nodes.
		p.lastNodes = buildDNodes([]string{"node1", "node2"})

		Convey("When", func() {
			nextNodes := buildDNodes([]string{"node1", "node2", "node3"})
			p.processNodeListing(nextNodes)
			Convey("Then", func() {
				So(len(p.lastNodes), ShouldEqual, 3)
			})
		})
	})
}

func TestProcessOneNodeRemoved(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEventServer := mock_comms.NewMockIEventServer(ctrl)
	mockEventServer.EXPECT().AddEventToSendQueue(gomock.Any()).Times(1)

	p := NewPublisher(mockEventServer, cmd.DefaultConfiguration())

	Convey("Given", t, func() {
		// Start state, start with two nodes.
		p.lastNodes = buildDNodes([]string{"node1", "node2"})
		Convey("When", func() {
			nextNodes := buildDNodes([]string{"node2"})
			p.processNodeListing(nextNodes)
			Convey("Then", func() {
				So(len(p.lastNodes), ShouldEqual, 1)
			})
		})
	})
}

func TestProcessOneNodeRemovedTwoAdded(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEventServer := mock_comms.NewMockIEventServer(ctrl)
	mockEventServer.EXPECT().AddEventToSendQueue(gomock.Any()).Times(3)

	p := NewPublisher(mockEventServer, cmd.DefaultConfiguration())

	Convey("Given", t, func() {

		// Start state, start with two nodes.
		p.lastNodes = buildDNodes([]string{"node1", "node2"})
		Convey("When", func() {
			nextNodes := buildDNodes([]string{"node2", "node3", "node4"})
			p.processNodeListing(nextNodes)
			Convey("Then", func() {
				So(len(p.lastNodes), ShouldEqual, 3)
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
