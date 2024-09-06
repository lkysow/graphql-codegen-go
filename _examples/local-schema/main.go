package main

import (
	"fmt"

	"github.com/lkysow/graphql-codegen-go/_examples/local-schema/internal/appsync"
)

func main() {
	newYear := appsync.EnumYearNEW
	fmt.Println(newYear)
	e := appsync.Entity1{Y: &newYear}
	fmt.Printf("%v\n", e)
}
