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
package main

import (
	"github.com/eriklupander/dvizz/comms"
	"github.com/eriklupander/dvizz/service"
	"github.com/fsouza/go-dockerclient"
	"log"
	"sync"
)

func main() {

	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}

	service.SetEventServer(&comms.EventServer{Client: client})

	go service.PublishTasks(client)
	log.Println("Initialized publishTasks")

	go service.PublishServices(client)
	log.Println("Initialized publishServices")

	go service.PublishNodes(client)
	log.Println("Initialized publishNodes")

	// Block...
	log.Println("Waiting at block...")

	wg := sync.WaitGroup{} // Use a WaitGroup to block main() exit
	wg.Add(1)
	wg.Wait()
}
