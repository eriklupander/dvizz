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
package dvizz

import (
	"github.com/eriklupander/dvizz/internal/pkg/comms"
	"github.com/eriklupander/dvizz/internal/pkg/service"
	docker "github.com/fsouza/go-dockerclient"
	"log"
	"sync"
)

func main() {
	log.Println("Starting dvizz!")
	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		panic(err)
	}

	publisher := service.NewPublisher(&comms.EventServer{Client: dockerClient})

	go publisher.PublishTasks(dockerClient)
	log.Println("Initialized publishTasks")

	go publisher.PublishServices(dockerClient)
	log.Println("Initialized publishServices")

	go publisher.PublishNodes(dockerClient)
	log.Println("Initialized publishNodes")

	// Block...
	log.Println("Waiting at block...")

	wg := sync.WaitGroup{} // Use a WaitGroup to block main() exit
	wg.Add(1)
	wg.Wait()
}
