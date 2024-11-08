# Setting Up Multiple Validators at Genesis Time
This guide demonstrates how to initialize a Cosmos SDK chain with multiple validators from the start.

## Initial Setup

```bash
# Create network and containers for all validators
docker network create blog-testnet

# Start containers for all validators
docker run -d --name validator1 \
    --network blog-testnet \
    -p 26656:26656 -p 26657:26657 -p 1317:1317 \
    blog-chain /bin/bash -c "tail -f /dev/null"

docker run -d --name validator2 \
    --network blog-testnet \
    -p 26666:26656 -p 26667:26657 -p 1318:1317 \
    blog-chain /bin/bash -c "tail -f /dev/null"

docker run -d --name validator3 \
    --network blog-testnet \
    -p 26676:26656 -p 26677:26657 -p 1319:1317 \
    blog-chain /bin/bash -c "tail -f /dev/null"
```

## Step 1: Initialize All Nodes
```bash
# Initialize all validators
docker exec -it validator1 bash -c 'blogd init validator1 --chain-id blog-testnet'
docker exec -it validator2 bash -c 'blogd init validator2 --chain-id blog-testnet'
docker exec -it validator3 bash -c 'blogd init validator3 --chain-id blog-testnet'
```

## Step 2: Create Keys for All Validators
```bash
# Create keys for all validators
docker exec -it validator1 bash -c 'blogd keys add val1 --keyring-backend test'
docker exec -it validator2 bash -c 'blogd keys add val2 --keyring-backend test'
docker exec -it validator3 bash -c 'blogd keys add val3 --keyring-backend test'

# Save addresses for next step
VAL1_ADDR=$(docker exec -it validator1 bash -c 'blogd keys show val1 -a --keyring-backend test')
VAL2_ADDR=$(docker exec -it validator2 bash -c 'blogd keys show val2 -a --keyring-backend test')
VAL3_ADDR=$(docker exec -it validator3 bash -c 'blogd keys show val3 -a --keyring-backend test')
```

## Step 3: Add Genesis Accounts (On validator1)
```bash
# Add all validator accounts to genesis
docker exec -it validator1 bash -c \
"blogd genesis add-genesis-account $VAL1_ADDR 1000000000000stake,1000000000000token && \
 blogd genesis add-genesis-account $VAL2_ADDR 1000000000000stake,1000000000000token && \
 blogd genesis add-genesis-account $VAL3_ADDR 1000000000000stake,1000000000000token"
```

## Step 4: Create Gentx for Each Validator
```bash
# Create gentx for validator1 (on validator1)
docker exec -it validator1 bash -c \
'blogd genesis gentx val1 900000000000stake \
  --chain-id blog-testnet \
  --moniker="validator1" \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --keyring-backend test'

# Copy genesis.json to validator2 and validator3
docker cp validator1:/root/.blog/config/genesis.json /tmp/genesis.json
docker cp /tmp/genesis.json validator2:/root/.blog/config/genesis.json
docker cp /tmp/genesis.json validator3:/root/.blog/config/genesis.json

# Create gentx for validator2 (on validator2)
docker exec -it validator2 bash -c \
'blogd genesis gentx val2 800000000000stake \
  --chain-id blog-testnet \
  --moniker="validator2" \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --keyring-backend test'

# Create gentx for validator3 (on validator3)
docker exec -it validator3 bash -c \
'blogd genesis gentx val3 700000000000stake \
  --chain-id blog-testnet \
  --moniker="validator3" \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --keyring-backend test'
```

## Step 5: Collect All Gentxs
```bash
# Copy gentxs from validator2 and validator3 to validator1
docker cp validator2:/root/.blog/config/gentx/. /tmp/gentx-validator2
docker cp validator3:/root/.blog/config/gentx/. /tmp/gentx-validator3
docker cp /tmp/gentx-validator2 validator1:/root/.blog/config/gentx/
docker cp /tmp/gentx-validator3 validator1:/root/.blog/config/gentx/

# Collect all gentxs
docker exec -it validator1 bash -c 'blogd genesis collect-gentxs'

# Copy final genesis to other validators
docker cp validator1:/root/.blog/config/genesis.json /tmp/genesis.json
docker cp /tmp/genesis.json validator2:/root/.blog/config/genesis.json
docker cp /tmp/genesis.json validator3:/root/.blog/config/genesis.json
```

## Step 6: Set Up Peer Connections
```bash
# Get node IDs
VAL1_ID=$(docker exec -it validator1 bash -c 'blogd tendermint show-node-id')
VAL2_ID=$(docker exec -it validator2 bash -c 'blogd tendermint show-node-id')
VAL3_ID=$(docker exec -it validator3 bash -c 'blogd tendermint show-node-id')

# Configure persistent peers for validator1
docker exec -it validator1 bash -c \
'sed -i.bak -e "s/^persistent_peers *=.*/persistent_peers = \"'$VAL2_ID@validator2:26656,$VAL3_ID@validator3:26656'\"/" $HOME/.blog/config/config.toml'

# Configure persistent peers for validator2
docker exec -it validator2 bash -c \
'sed -i.bak -e "s/^persistent_peers *=.*/persistent_peers = \"'$VAL1_ID@validator1:26656,$VAL3_ID@validator3:26656'\"/" $HOME/.blog/config/config.toml'

# Configure persistent peers for validator3
docker exec -it validator3 bash -c \
'sed -i.bak -e "s/^persistent_peers *=.*/persistent_peers = \"'$VAL1_ID@validator1:26656,$VAL2_ID@validator2:26656'\"/" $HOME/.blog/config/config.toml'
```

## Step 7: Start the Chain
```bash
# Start all validators
docker exec -it validator1 bash -c 'blogd start'
docker exec -it validator2 bash -c 'blogd start'
docker exec -it validator3 bash -c 'blogd start'
```

## Step 8: Verify Setup
```bash
# Check validator set
docker exec -it validator1 bash -c 'blogd query staking validators --output json | jq'

# Check individual validator status
docker exec -it validator1 bash -c 'blogd status'
docker exec -it validator2 bash -c 'blogd status'
docker exec -it validator3 bash -c 'blogd status'
```

## Key Differences from Single Validator Setup
1. All validators are included in genesis state
2. Gentxs from all validators are collected before chain start
3. Initial stake distribution is determined at genesis
4. All validators start participating in consensus from block 1
5. No need for create-validator transactions after chain start

## Important Considerations
1. **Stake Distribution**: Consider the initial token distribution and stake amounts to ensure proper decentralization
2. **Validator Coordination**: All validators must coordinate to combine their gentxs before genesis
3. **Genesis Parameters**: Make sure parameters like `max_validators` are set appropriately
4. **Backup**: Keep backup of all keys and genesis configurations
5. **Testing**: Test the setup in a test environment before going to production

Remember to maintain security of validator keys and ensure all validators are properly synchronized before starting the chain.