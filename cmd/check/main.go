package main

import (
	"fmt"
	"log"

	"github.com/DestroctorGT/bark-sentinel/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuracion invalida: %v", err)
	}

	fmt.Printf("%+v\n", cfg)
}
