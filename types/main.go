package types

type ConfigCommands struct {
	Dir    string   `json:"dir"`
	Ignore []string `json:"ignore"`
}

type ConfigI18n struct {
	Dir     string   `json:"dir"`
	Locales []string `json:"locales"`
}

type Config struct {
	Commands ConfigCommands `json:"commands"`
	I18n     ConfigI18n     `json:"i18n"`
}
