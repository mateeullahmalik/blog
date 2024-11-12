package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type GlobalConfig struct {
	ChainID        string       `json:"chain_id"`
	KeyringBackend string       `json:"keyring_backend"`
	GasPrice       string       `json:"gas_price"`
	DataDir        string       `json:"data_dir"`
	Binary         BinaryConfig `json:"binary"`
}

type BinaryConfig struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

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

func loadGlobalConfig(filename string) (*GlobalConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading global config file: %v", err)
	}

	var config GlobalConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing global config file: %v", err)
	}

	return &config, nil
}

func generateDockerCompose(configs []ValidatorConfig, globalCfg *GlobalConfig) string {
	var services string

	// Generate services for all validators
	for i, config := range configs {
		isFirst := i == 0
		services += fmt.Sprintf("  %s:\n%s\n", config.Name, generateValidatorScript(config, configs, globalCfg, isFirst))
	}

	// Complete docker-compose structure with updated binary path
	return fmt.Sprintf(`version: '3'

services:
%s
networks:
  default:
    name: blog-network

volumes:
  shared:`, services)
}

func main() {
	globalCfg, err := loadGlobalConfig("config.json")
	if err != nil {
		fmt.Printf("Error loading global config: %v\n", err)
		return
	}

	configs, err := loadValidatorConfigs("validators.json")
	if err != nil {
		fmt.Printf("Error loading validator configs: %v\n", err)
		return
	}

	dockerCompose := generateDockerCompose(configs, globalCfg)

	// Save to docker-compose.yml
	err = os.WriteFile("docker-compose.yml", []byte(dockerCompose), 0644)
	if err != nil {
		fmt.Printf("Error writing docker-compose.yml: %v\n", err)
		return
	}

	fmt.Println("Successfully generated docker-compose.yml")
}
