package internal

var Version = "1.0.3"

type AppMetadata struct {
    Version string
    Name    string
    Author  string
    License string
    Github  string
}

var AppMetaData = AppMetadata{
    Version: Version,
    Name:    "Prasmoid",
    Author:  "PRAS",
    License: "MIT",
    Github:  "https://github.com/PRASSamin/prasmoid",
}