package main

import (
	"modelgenerator/db"
	"modelgenerator/generate"
)

func main() {
	db.Init()
	generate.Generate()
}
