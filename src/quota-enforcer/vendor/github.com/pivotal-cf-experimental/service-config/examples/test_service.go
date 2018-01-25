package main

import (
	"github.com/pivotal-cf-experimental/service-config"

	"flag"
	"fmt"
	"os"
)

type ShipConfig struct {
	Name   string `yaml:"Name"`
	ID     int    `yaml:"ID"`
	Crew   Crew   `yaml:"Crew"`
	Active bool   `yaml:"Active"`
}

type Crew struct {
	Officers   []Officer   `yaml:"Officers"`
	Passengers []Passenger `yaml:"Passengers"`
}

type Officer struct {
	Name string `yaml:"Name"`
	Role string `yaml:"Role"`
}

type Passenger struct {
	Name  string `yaml:"Name"`
	Title string `yaml:"Title"`
}

func main() {
	serviceConfig := service_config.New()

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	serviceConfig.AddDefaults(ShipConfig{
		Active: true,
	})

	serviceConfig.AddFlags(flags)
	flags.Parse(os.Args[1:])

	var config ShipConfig
	err := serviceConfig.Read(&config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read config: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("Config: %#v\n", config)
}
