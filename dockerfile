FROM alpine:latest

# Install required tools
RUN apk add --no-cache \
    curl \
    jq \
    bash

# Copy binary and make it executable 
COPY release/blogd /usr/local/bin/blogd
RUN chmod +x /usr/local/bin/blogd

# Expose necessary ports
# P2P, RPC, REST API, GRPC
EXPOSE 26656 26657 1317 9090

# Set working directory
WORKDIR /root

# Create directory for chain data
RUN mkdir -p /root/.blog

# Keep container running (docker-compose will override command)
CMD ["blogd", "start"]