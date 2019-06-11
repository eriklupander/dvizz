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
	"fmt"
	"github.com/containous/flaeg"
	"github.com/containous/flaeg/parse"
	"github.com/eriklupander/dvizz/cmd"
	"github.com/eriklupander/dvizz/internal/pkg/comms"
	"github.com/eriklupander/dvizz/internal/pkg/service"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/ogier/pflag"
	"github.com/sirupsen/logrus"
	fmtlog "log"
	"os"
	"reflect"
	"strings"
	"sync"
)

type dvizzConfiguration struct {
	cmd.GlobalConfiguration
}

func defaultDvizzPointersConfiguration() *dvizzConfiguration {
	return &dvizzConfiguration{}
}

func defaultDvizzConfiguration() *dvizzConfiguration {
	return &dvizzConfiguration{
		GlobalConfiguration: *cmd.DefaultConfiguration(),
	}
}

func main() {
	defaultConfiguration := defaultDvizzConfiguration()
	defaultPointersConfiguration := defaultDvizzPointersConfiguration()

	mainCommand := &flaeg.Command{
		Name:                  "dvizz",
		Description:           "dvizz main process. Set DOCKER_HOST env var if connecting to a non-local Docker Swarm cluster",
		Config:                defaultConfiguration,
		DefaultPointersConfig: defaultPointersConfiguration,
		Run: func() error {
			run(defaultConfiguration)
			return nil
		},
	}

	f := flaeg.New(mainCommand, os.Args[1:])
	f.AddParser(reflect.TypeOf([]string{}), &parse.SliceStrings{})

	usedCmd, err := f.GetCommand()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if _, err := f.Parse(usedCmd); err != nil {
		if err == pflag.ErrHelp {
			os.Exit(0)
		}
		fmt.Printf("Error parsing command: %s\n", err)
		os.Exit(1)
	}

	if err := f.Run(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	os.Exit(0)
}

func run(cfg *dvizzConfiguration) {
	configureLogging(cfg)
	logrus.Println("Starting dvizz!")
	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		panic(err)
	}

	eventServer := &comms.EventServer{Client: dockerClient}
	go eventServer.InitializeEventSystem()

	publisher := service.NewPublisher(eventServer, &cfg.GlobalConfiguration)

	go publisher.PublishTasks(dockerClient)
	logrus.Infof("Initialized publishTasks, will poll every %v seconds", cfg.TaskPoll)

	go publisher.PublishServices(dockerClient)
	logrus.Infof("Initialized publishServices, will poll every %v seconds", cfg.ServicePoll)

	go publisher.PublishNodes(dockerClient)
	logrus.Infof("Initialized publishNodes, will poll every %v seconds", cfg.NodePoll)

	// Block...
	logrus.Println("Waiting at block...")

	wg := sync.WaitGroup{} // Use a WaitGroup to block main() exit
	wg.Add(1)
	wg.Wait()
}

// ConfigureLogging Configure logging for all cmd.
func configureLogging(configuration *dvizzConfiguration) {
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
