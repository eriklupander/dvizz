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
package model

type Identifier interface {
	GetId() string
}

type DObject interface {
	Equals(other Identifier) bool
}

type DEvent struct {
	Action string `json:"action"` // create or stop or update
	Type   string `json:"type"`
	Dtask  DTask  `json:"dtask"`
}

type DTaskStateUpdate struct {
	Action string `json:"action"` // create or stop or update
	Type   string `json:"type"`   // typically task
	Id     string `json:"id"`
	State  string `json:"state"`
}

type DNode struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	State  string `json:"state"`
	Memory string `json:"memory"`
	CPUs   string `json:"cpus"`
}

type DNodeEvent struct {
	Action string `json:"action"` // create or stop or update
	Type   string `json:"type"`
	Dnode  DNode  `json:"dnode"`
}

func (d DNode) GetId() string {
	return d.Id
}

func (d DNode) Equals(d2 Identifier) bool {
	// log.Printf("About to compare DNode %v with %v: result %v", d.GetId(), d2.GetId(), d.GetId() == d2.GetId())

	return d.GetId() == d2.GetId()
}

type DTask struct {
	Id        string     `json:"id"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	ServiceId string     `json:"serviceId"`
	NodeId    string     `json:"nodeId"`
	Networks  []DNetwork `json:"networks"`
}

type DNetwork struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (d DTask) Equals(d2 Identifier) bool {
	// log.Printf("About to compare DTask %v with %v: result %v", d.GetId(), d2.GetId(), d.GetId() == d2.GetId())
	return d.GetId() == d2.GetId()
}

func (d DTask) GetId() string {
	return d.Id
}

type DServiceEvent struct {
	Action   string   `json:"action"` // create or stop or destroy
	Type     string   `json:"type"`
	DService DService `json:"dservice"`
}

type DService struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	//  Image string  `json:"image"`
}

func (d DService) Equals(d2 Identifier) bool {
	// log.Printf("About to compare DService %v with %v: result %v", d.GetId(), d2.GetId(), d.GetId() == d2.GetId())

	return d.GetId() == d2.GetId()
}

func (d DService) GetId() string {
	return d.Id
}

// Asserts
var _ Identifier = (*DService)(nil)
var _ Identifier = (*DTask)(nil)
var _ Identifier = (*DNode)(nil)
var _ DObject = (*DService)(nil)
var _ DObject = (*DTask)(nil)
var _ DObject = (*DNode)(nil)
