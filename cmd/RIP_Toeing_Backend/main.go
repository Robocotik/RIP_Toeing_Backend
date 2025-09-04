package main

import (
	"backend/internal/api"
	"fmt"
)

func main() {
	fmt.Println("Application starts")
	defer fmt.Println("Application finished")
	api.StartServer()
}
