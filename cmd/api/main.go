package main

const (
	configFilePath = "configs/config.yaml"
)

func main() {
	app, cleanup, err := wiring.InitializeStandardServer(config.ConfigFilePath(configFilePath))
	if err != nil {
		panic("failed to initialize server: " + err.Error())
	}
	defer cleanup()
	app.Start()
}
