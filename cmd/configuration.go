package cmd

type GlobalConfiguration struct {
	PollConfig
	LogLevel string `short:"l" description:"Log level"`
}

type PollConfig struct {
	NodePoll    int `short:"n" description:"Node poll interval, seconds"`
	ServicePoll int `short:"s" description:"Service poll interval, seconds"`
	TaskPoll    int `short:"t" description:"Task poll interval, seconds"`
}

func DefaultConfiguration() *GlobalConfiguration {

	return &GlobalConfiguration{
		LogLevel: "info",
		PollConfig: PollConfig{
			NodePoll:    60,
			ServicePoll: 30,
			TaskPoll:    10,
		},
	}
}
