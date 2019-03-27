package main

import (
	"log"
	"os"

	"github.com/fletaio/citygame/explorer/city_explorer"

	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(cityexplorer.Assets, vfsgen.Options{
		PackageName:  "cityexplorer",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatal(err)
	}

	oldLocation := "./assets_vfsdata.go"
	newLocation := "../assets_vfsdata.go"
	err = os.Rename(oldLocation, newLocation)
	if err != nil {
		log.Fatal(err)
	}
}
