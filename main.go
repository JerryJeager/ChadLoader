package main

import (
	"log"

	"github.com/JerryJeager/ChadLoader/cmd"
)

func main() {
	log.Println("Starting ChadLoader Server...")

	cmd.ExecuteApiRoutes()
}
