package internal

var Version = "0.0.1"

type Metadata struct {
    Version string
    Name    string
    Author  string
    License string
    Github  string
}

var AppMeta = Metadata{
    Version: Version,
    Name:    "Prasmoid",
    Author:  "PRAS",
    License: "MIT",
    Github:  "https://github.com/PRASSamin/prasmoid",
}
