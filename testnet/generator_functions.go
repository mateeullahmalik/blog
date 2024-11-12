package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func loadValidatorConfigs(filename string) ([]ValidatorConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var configs []ValidatorConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return configs, nil
}

func generateWaitConditions(configs []ValidatorConfig, currentValidator string) string {
	var conditions []string

	for _, config := range configs {
		if config.Name != currentValidator {
			conditions = append(conditions, fmt.Sprintf("! -f /shared/%s_address", config.Name))
		}
	}

	return fmt.Sprintf("while [[ %s ]]; do\n          echo \"Waiting for other validators to initialize...\"\n          sleep 1\n        done", strings.Join(conditions, " || "))
}

func generateGenesisAccounts(configs []ValidatorConfig, currentValidator string, globalCfg *GlobalConfig) string {
	var commands []string

	var currentConfig ValidatorConfig
	for _, cfg := range configs {
		if cfg.Name == currentValidator {
			currentConfig = cfg
			break
		}
	}

	commands = append(commands, fmt.Sprintf("%s genesis add-genesis-account $$ADDR %s,%s",
		globalCfg.Binary.Name,
		currentConfig.Tokens.Stake,
		currentConfig.Tokens.Token))

	for _, config := range configs {
		if config.Name != currentValidator {
			commands = append(commands,
				fmt.Sprintf("VAL_%s_ADDR=$$(cat /shared/%s_address)",
					strings.ToUpper(config.Name), config.Name))
			commands = append(commands,
				fmt.Sprintf("%s genesis add-genesis-account $$VAL_%s_ADDR %s,%s",
					globalCfg.Binary.Name,
					strings.ToUpper(config.Name),
					config.Tokens.Stake,
					config.Tokens.Token))
		}
	}

	return strings.Join(commands, "\n        ")
}

func generateGentxWaitAndCollection(configs []ValidatorConfig, currentValidator string, globalCfg *GlobalConfig) string {
	var waitConditions []string
	var copyCommands []string

	for _, config := range configs {
		if config.Name != currentValidator {
			waitConditions = append(waitConditions, fmt.Sprintf("! -f /shared/%s_gentx.json", config.Name))
			copyCommands = append(copyCommands, fmt.Sprintf("cp /shared/%s_gentx.json %s/config/gentx/", config.Name, globalCfg.DataDir))
		}
	}

	return fmt.Sprintf(`# Wait for other validators gentxs
        while [[ %s ]]; do
          echo "Waiting for other validators gentxs..."
          sleep 1
        done
        
        # Collect gentxs and create final genesis
        mkdir -p %s/config/gentx
        %s
        %s genesis collect-gentxs
        cp %s/config/genesis.json /shared/final_genesis.json
        echo "true" > /shared/setup_complete`,
		strings.Join(waitConditions, " || "),
		globalCfg.DataDir,
		strings.Join(copyCommands, "\n        "),
		globalCfg.Binary.Name,
		globalCfg.DataDir)
}

func generateValidatorCommand(config ValidatorConfig, configs []ValidatorConfig, globalCfg *GlobalConfig, isFirst bool) string {
	if isFirst {
		return fmt.Sprintf(`    command: |
      bash -c '
      if [[ ! -f %s/config/genesis.json ]] || [[ ! -f %s/config/priv_validator_key.json ]]; then
        echo "First time initialization for %s..."
        
        # First time initialization
        %s init %s --chain-id %s --overwrite
        %s keys add %s --keyring-backend %s
        ADDR=$(%s keys show %s -a --keyring-backend %s)
        echo $$ADDR > /shared/%s_address
        
        %s
        
        %s
        
        # Share genesis and create gentx
        cp %s/config/genesis.json /shared/genesis.json
        echo "true" > /shared/genesis_accounts_ready
        %s genesis gentx %s %s --chain-id %s --keyring-backend %s
        cp %s/config/gentx/*.json /shared/%s_gentx.json
        
        %s
      else
        echo "%s already initialized, starting chain..."
      fi
      
      # Get node ID and share it
      nodeid=$(%s tendermint show-node-id)
      echo $$nodeid > /shared/%s_nodeid

      # Wait for other node IDs
      while [[ %s ]]; do
        echo "Waiting for other node IDs..."
        sleep 1
      done

      # Create persistent peers string
      %s

      # Update persistent peers
      sed -i "s/^persistent_peers *=.*/persistent_peers = \"$$PEERS\"/" %s/config/config.toml
      
      # Set gas prices and start chain
      sed -i "s/minimum-gas-prices = \"\"/minimum-gas-prices = \"%s\"/" %s/config/app.toml
      %s start --minimum-gas-prices=%s'`,
			globalCfg.DataDir, globalCfg.DataDir,
			config.Name,
			globalCfg.Binary.Name, config.Moniker, globalCfg.ChainID,
			globalCfg.Binary.Name, config.KeyName, globalCfg.KeyringBackend,
			globalCfg.Binary.Name, config.KeyName, globalCfg.KeyringBackend,
			config.Name,
			generateWaitConditions(configs, config.Name),
			generateGenesisAccounts(configs, config.Name, globalCfg),
			globalCfg.DataDir,
			globalCfg.Binary.Name, config.KeyName, config.Tokens.GentxStake, globalCfg.ChainID, globalCfg.KeyringBackend,
			globalCfg.DataDir, config.Name,
			generateGentxWaitAndCollection(configs, config.Name, globalCfg),
			config.Name,
			globalCfg.Binary.Name,
			config.Name,
			generatePeerWaitConditions(configs, config.Name),
			generatePeersString(configs, config.Name),
			globalCfg.DataDir,
			globalCfg.GasPrice, globalCfg.DataDir,
			globalCfg.Binary.Name, globalCfg.GasPrice)
	}

	return fmt.Sprintf(`    command: |
      bash -c '
      if [[ ! -f %s/config/genesis.json ]] || [[ ! -f %s/config/priv_validator_key.json ]]; then
        echo "First time initialization for %s..."
        
        # First time initialization
        %s init %s --chain-id %s --overwrite
        %s keys add %s --keyring-backend %s
        ADDR=$(%s keys show %s -a --keyring-backend %s)
        echo $$ADDR > /shared/%s_address
        
        # Wait for genesis file
        while [[ ! -f /shared/genesis_accounts_ready ]]; do
          echo "Waiting for genesis accounts..."
          sleep 1
        done
        
        # Copy genesis and create gentx
        cp /shared/genesis.json %s/config/genesis.json
        %s genesis gentx %s %s --chain-id %s --keyring-backend %s
        cp %s/config/gentx/*.json /shared/%s_gentx.json
        
        # Wait for final genesis
        while [[ ! -f /shared/final_genesis.json ]]; do
          echo "Waiting for final genesis..."
          sleep 1
        done
        cp /shared/final_genesis.json %s/config/genesis.json
      else
        echo "%s already initialized, starting chain..."
      fi

      # Get node ID and share it
      nodeid=$(%s tendermint show-node-id)
      echo $$nodeid > /shared/%s_nodeid

      # Wait for other node IDs
      while [[ %s ]]; do
        echo "Waiting for other node IDs..."
        sleep 1
      done

      # Create persistent peers string
      %s

      # Update persistent peers
      sed -i "s/^persistent_peers *=.*/persistent_peers = \"$$PEERS\"/" %s/config/config.toml
      
      # Set gas prices and start chain
      sed -i "s/minimum-gas-prices = \"\"/minimum-gas-prices = \"%s\"/" %s/config/app.toml
      while [[ ! -f /shared/setup_complete ]]; do
        echo "Waiting for setup to complete..."
        sleep 1
      done
      %s start --minimum-gas-prices=%s'`,
		globalCfg.DataDir, globalCfg.DataDir,
		config.Name,
		globalCfg.Binary.Name, config.Moniker, globalCfg.ChainID,
		globalCfg.Binary.Name, config.KeyName, globalCfg.KeyringBackend,
		globalCfg.Binary.Name, config.KeyName, globalCfg.KeyringBackend,
		config.Name,
		globalCfg.DataDir,
		globalCfg.Binary.Name, config.KeyName, config.Tokens.GentxStake, globalCfg.ChainID, globalCfg.KeyringBackend,
		globalCfg.DataDir, config.Name,
		globalCfg.DataDir,
		config.Name,
		globalCfg.Binary.Name,
		config.Name,
		generatePeerWaitConditions(configs, config.Name),
		generatePeersString(configs, config.Name),
		globalCfg.DataDir,
		globalCfg.GasPrice, globalCfg.DataDir,
		globalCfg.Binary.Name, globalCfg.GasPrice)
}

func generateValidatorScript(config ValidatorConfig, configs []ValidatorConfig, globalCfg *GlobalConfig, isFirst bool) string {
	script := fmt.Sprintf(`    build: .
    container_name: blog-%s
    ports:
      - "%d:%d"  # P2P
      - "%d:%d"  # RPC
      - "%d:%d"    # REST
      - "%d:%d"    # GRPC
    volumes:
      - ./%s-data:%s
      - ./shared:/shared
    environment:
      - MONIKER=%s`,
		config.Name,
		config.Port, 26656,
		config.RPCPort, 26657,
		config.RESTPort, 1317,
		config.GRPCPort, 9090,
		config.Name, globalCfg.DataDir,
		config.Moniker)

	if !isFirst {
		script += "\n    depends_on:\n      - validator1"
	}

	// Append command with proper indentation
	script += "\n"
	script += generateValidatorCommand(config, configs, globalCfg, isFirst)
	return script
}

func generatePeerWaitConditions(configs []ValidatorConfig, currentValidator string) string {
	var conditions []string
	for _, config := range configs {
		if config.Name != currentValidator {
			conditions = append(conditions, fmt.Sprintf("! -f /shared/%s_nodeid", config.Name))
		}
	}
	return strings.Join(conditions, " || ")
}

func generatePeersString(configs []ValidatorConfig, currentValidator string) string {
	var peerCommands []string
	for _, config := range configs {
		if config.Name != currentValidator {
			peerCommands = append(peerCommands,
				fmt.Sprintf("NODE_%s_ID=$$(cat /shared/%s_nodeid)",
					strings.ToUpper(config.Name),
					config.Name))
		}
	}

	var peerParts []string
	for _, config := range configs {
		if config.Name != currentValidator {
			peerParts = append(peerParts,
				fmt.Sprintf("$$NODE_%s_ID@%s:%d",
					strings.ToUpper(config.Name),
					config.Name,
					26656))
		}
	}

	peerCommands = append(peerCommands,
		fmt.Sprintf("PEERS=\"%s\"",
			strings.Join(peerParts, ",")))

	return strings.Join(peerCommands, "\n        ")
}
