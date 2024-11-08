# Blog Chain Testnet

A multi-validator testnet setup for the Blog blockchain using Docker Compose.

## Network Architecture

The testnet consists of 3 validator nodes running in separate Docker containers:

| Node | Container Name | Ports (External:Internal) |
|------|---------------|--------------------------|
| Validator 1 | blog-validator1 | 26656:26656 (P2P) <br> 26657:26657 (RPC) <br> 1317:1317 (REST) <br> 9090:9090 (GRPC) |
| Validator 2 | blog-validator2 | 26666:26656 (P2P) <br> 26667:26657 (RPC) <br> 1327:1317 (REST) <br> 9092:9090 (GRPC) |
| Validator 3 | blog-validator3 | 26676:26656 (P2P) <br> 26677:26657 (RPC) <br> 1337:1317 (REST) <br> 9093:9090 (GRPC) |

## Prerequisites

- Docker
- Docker Compose

## Setup Instructions

1. Build the blockchain binary:
```bash
ignite chain build
```

2. Create a `release` directory and copy the binary:
```bash
mkdir release
cp build/blogd release/
```

3. Start the testnet:
```bash
docker-compose up -d
```

4. Check if all validators are running:
```bash
docker ps
```

## Validator Information

### Validator 1 (Primary)
- Moniker: validator1
- Account: alice
- Home: /root/.blog
- Access REST API: http://localhost:1317
- Access RPC: http://localhost:26657
- Access GRPC: http://localhost:9090

### Validator 2
- Moniker: validator2
- Account: bob
- Home: /root/.blog
- Access REST API: http://localhost:1327
- Access RPC: http://localhost:26667
- Access GRPC: http://localhost:9092

### Validator 3
- Moniker: validator3
- Account: carol
- Home: /root/.blog
- Access REST API: http://localhost:1337
- Access RPC: http://localhost:26677
- Access GRPC: http://localhost:9093

## Testing the Network

### Check Node Status
```bash
# Validator 1
curl http://localhost:26657/status

# Validator 2
curl http://localhost:26667/status

# Validator 3
curl http://localhost:26677/status
```

### Create a Blog Post (Using Validator 1)
```bash
docker exec blog-validator1 blogd tx blog create-post "Hello" "World" \
  --from alice \
  --chain-id blog-testnet \
  --keyring-backend test \
  -y
```

### Query Posts (Can be done from any validator)
```bash
# From Validator 1
docker exec blog-validator1 blogd q blog list-post

# From Validator 2
docker exec blog-validator2 blogd q blog list-post

# From Validator 3
docker exec blog-validator3 blogd q blog list-post
```

### Update a Post
```bash
docker exec blog-validator1 blogd tx blog update-post "Updated Title" "Updated Body" 0 \
  --from alice \
  --chain-id blog-testnet \
  --keyring-backend test \
  -y
```

### Delete a Post
```bash
docker exec blog-validator1 blogd tx blog delete-post 0 \
  --from alice \
  --chain-id blog-testnet \
  --keyring-backend test \
  -y
```

## Container Management

### View Logs
```bash
# Validator 1 logs
docker logs blog-validator1

# Validator 2 logs
docker logs blog-validator2

# Validator 3 logs
docker logs blog-validator3
```

### Access Container Shell
```bash
# Validator 1 shell
docker exec -it blog-validator1 sh

# Validator 2 shell
docker exec -it blog-validator2 sh

# Validator 3 shell
docker exec -it blog-validator3 sh
```

### Stop the Network
```bash
docker-compose down
```

### Clean Up
To remove all containers and data:
```bash
docker-compose down -v
rm -rf validator1-data validator2-data validator3-data
```

## Network Details

- Chain ID: blog-testnet
- Validator Setup: 3 validators
- Token Denominations: stake, token
- Block Time: ~5 seconds
- Consensus: Tendermint
- Network: All containers run on a custom Docker network named "blog-network"

## Troubleshooting

1. If nodes are not connecting:
   - Check logs using `docker logs blog-validator1`
   - Verify persistent peers in config.toml
   - Ensure ports are correctly mapped

2. If transactions fail:
   - Verify account has sufficient funds
   - Check chain-id matches
   - Ensure correct keyring backend is specified

3. If containers fail to start:
   - Verify binary is correctly built and placed in release folder
   - Check Docker logs for specific errors
   - Ensure no port conflicts on host machine

## Security Notes

This setup is intended for testing and development purposes only. For production:
- Configure proper security measures
- Use secure passwords and keys
- Enable proper firewalls
- Configure proper persistent storage
- Set up monitoring and alerting