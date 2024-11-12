package main

import (
	"fmt"
	"os"
)

type TokenConfig struct {
	Stake      string `json:"stake"`
	Token      string `json:"token"`
	GentxStake string `json:"gentx_stake"`
}

type ValidatorConfig struct {
	Name     string      `json:"name"`
	Moniker  string      `json:"moniker"`
	KeyName  string      `json:"key_name"`
	Port     int         `json:"port"`
	RPCPort  int         `json:"rpc_port"`
	RESTPort int         `json:"rest_port"`
	GRPCPort int         `json:"grpc_port"`
	Tokens   TokenConfig `json:"tokens"`
}

func generateDockerCompose(configs []ValidatorConfig) string {
	var services string

	// Generate services for all validators
	for i, config := range configs {
		isFirst := i == 0
		services += fmt.Sprintf("  %s:\n%s\n", config.Name, generateValidatorScript(config, configs, isFirst))
	}

	// Complete docker-compose structure
	return fmt.Sprintf(`version: '3'

services:
%s
networks:
  default:
    name: blog-network`, services)
}

func main() {
	configs, err := loadValidatorConfigs("validators.json")
	if err != nil {
		fmt.Printf("Error loading configs: %v\n", err)
		return
	}

	dockerCompose := generateDockerCompose(configs)

	// Save to docker-compose.yml
	err = os.WriteFile("docker-compose.yml", []byte(dockerCompose), 0644)
	if err != nil {
		fmt.Printf("Error writing docker-compose.yml: %v\n", err)
		return
	}

	fmt.Println("Successfully generated docker-compose.yml")
}
