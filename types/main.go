package types

type ConfigCommands struct {
	Dir string `json:"dir"`
	DefaultRT string `json:"defaultRT"`
	Ignore []string `json:"ignore"`
}

type Config struct {
	Commands ConfigCommands `json:"commands"`
}