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
	"github.com/eriklupander/dvizz/cmd"
	"github.com/eriklupander/dvizz/internal/pkg/comms"
	"github.com/eriklupander/dvizz/internal/pkg/service"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/sirupsen/logrus"
	fmtlog "log"
	"strings"
	"sync"
)

func main() {
	configureLogging(cmd.DefaultConfiguration())
	logrus.Println("Starting dvizz!")
	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		panic(err)
	}

	eventServer := &comms.EventServer{Client: dockerClient}
	go eventServer.InitializeEventSystem()

	publisher := service.NewPublisher(eventServer)

	//	go publisher.PublishNetworks(dockerClient)

	go publisher.PublishTasks(dockerClient)
	logrus.Println("Initialized publishTasks")

	go publisher.PublishServices(dockerClient)
	logrus.Println("Initialized publishServices")

	go publisher.PublishNodes(dockerClient)
	logrus.Println("Initialized publishNodes")

	// Block...
	logrus.Println("Waiting at block...")

	wg := sync.WaitGroup{} // Use a WaitGroup to block main() exit
	wg.Add(1)
	wg.Wait()
}

// ConfigureLogging Configure logging for all cmd.
func configureLogging(configuration *cmd.GlobalConfiguration) {
	// configure default log flags
	fmtlog.SetFlags(fmtlog.Lshortfile | fmtlog.LstdFlags)
	// configure log level
	// an explicitly defined log level always has precedence. if none is
	// given and debug mode is disabled, the default is ERROR, and DEBUG
	// otherwise.
	levelStr := strings.ToLower(configuration.LogLevel)

	if levelStr == "" {
		levelStr = "error"
	}
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		fmtlog.Println("Error getting level", err)
	}
	logrus.SetLevel(level)

	ttyOK := false
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   ttyOK,
		DisableColors: false,
		FullTimestamp: true,
	})
}
