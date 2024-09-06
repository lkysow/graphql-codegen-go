package main

import (
	"fmt"

	"github.com/lkysow/graphql-codegen-go/_examples/config/internal"
)

func main() {
	person := internal.Person{
		FirstName: "",
		Lastname:  "",
		Age:       nil,
		Gender:    nil,
		Address:   nil,
	}
	fmt.Println(person)
}
