package cmd

type GlobalConfiguration struct {
	LogLevel string `short:"l" description:"Log level"`
}

func DefaultConfiguration() *GlobalConfiguration {

	return &GlobalConfiguration{
		LogLevel: "debug",
	}
}
