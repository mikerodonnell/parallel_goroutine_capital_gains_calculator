package main

import (
	"./httpServe"
	"./math"
	"dba"
	"fileio"

	"fmt"

	"github.com/davecgh/go-spew/spew"
	"log"
	"readInput"
)

func main() {
	fmt.Println("~~~~~~~~~~~~~~ math: ", math.DoMath(1, 4))

	authRepId := "mike5@asapp.com"

	//insertedRepId, err := dba.CreateRep(authRepId)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println("~~~~~~~~~~~~~~~ inserted rep: ", insertedRepId)

	readRep, err := dba.GetRep(authRepId)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("~~~~~~~~~~~~~~~ read rep: ", readRep)

	fileContents, err := fileio.ReadCSVDataRows("resources/users.csv")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("~~~~~~~~~~~~~~ file: ")
	spew.Dump(fileContents)

	stringInput, err := readInput.ReadStringInput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("~~~~~~~~~~~~~~ string input [%s] ", stringInput))

	numInput, err := readInput.ReadNumberInput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("~~~~~~~~~~~~~~ number input [%d] ", numInput))

	httpServe.Serve()
}
