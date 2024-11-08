# Blog Chain Development Setup

Quick guide to run a local 3-validator testnet using Docker Compose.

## Prerequisites
- Docker
- Docker Compose
- Ignite CLI

## Build Process

1. Build the chain binary for multiple platforms:
```bash
ignite chain build --release -t linux:amd64 -t darwin:amd64 -t darwin:arm64
```
This command:
- `--release`: Creates optimized binaries in the `release/` directory
- `-t linux:amd64`: Linux binary for x86_64 systems (used by Docker)
- `-t darwin:amd64`: macOS binary for Intel chips
- `-t darwin:arm64`: macOS binary for Apple Silicon

The binaries will be created in the `release/` directory:
- `blogd-linux-amd64`: For Linux/Docker
- `blogd-darwin-amd64`: For Intel Macs
- `blogd-darwin-arm64`: For M1/M2 Macs

2. Rename the Linux binary for Docker:
```bash
cp release/blogd-linux-amd64 release/blogd
```

## Setup & Run

1. Create required directories:
```bash
mkdir -p validator1-data validator2-data validator3-data shared
```

2. Start the network:
```bash
docker-compose up
```

## Access Points

- Validator1: 
  - RPC: `localhost:26657`
  - REST: `localhost:1317`
  - gRPC: `localhost:9090`
- Validator2:
  - RPC: `localhost:26667`
  - REST: `localhost:1327`
  - gRPC: `localhost:9092`
- Validator3:
  - RPC: `localhost:26677`
  - REST: `localhost:1337`
  - gRPC: `localhost:9093`

## Test Accounts
Using `--keyring-backend test`:
- alice (validator1): Initial balance 1000000000000stake,1000000000000token
- bob (validator2): Initial balance 1000000000000stake,1000000000000token
- carol (validator3): Initial balance 1000000000000stake,1000000000000token

## Common Operations

### Check Validator Status
```bash
# Query validators (from validator1)
docker exec blog-validator1 blogd query staking validators

# Check specific validator status
docker exec blog-validator1 blogd status
```

### Submit Transactions
```bash
# Create a post from alice's account
docker exec blog-validator1 blogd tx blog create-post "Title" "Body" --from alice --chain-id blog-testnet --keyring-backend test -y
```

### Query Chain State
```bash
# List all posts
docker exec blog-validator1 blogd query blog list-post

# Show specific post
docker exec blog-validator1 blogd query blog show-post 1
```

## Stop & Reset
```bash
# Stop the network
docker-compose down

# Clean all data (for fresh start)
rm -rf validator{1,2,3}-data shared/*
```

## Development Tips
- Use `docker-compose logs -f` to follow logs from all validators
- Each validator's data is persisted in its respective `validator{1,2,3}-data` directory
- The `shared` directory is used for genesis coordination between validators
- Gas prices are set to 0.00001stake by default