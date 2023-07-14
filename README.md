# ZenQuote

## Overview

ZenQuote is a simple TCP server and client application that provides Zen quotes after solving a Proof of Work (PoW) challenge. The server issues a challenge to the client, and the client needs to find a solution before receiving a quote. The concept is inspired by the PoW mechanism used in Blockchain technology, ensuring client computation contribution before receiving a service.

## Running the Server and Client

### Prerequisites

- Docker
- Docker Compose
- Make

### Steps

1. Clone the repository: `git clone https://github.com/your_username/zenquote.git`
2. Navigate to the root directory: `cd zenquote`
3. Build the Docker images: `make docker-build`
4. Start the Docker services: `make docker-up`

To stop the services, use: `make docker-down`

For logs, use: `make docker-log`

## Running the Tests

Use the command: `make test`

## Why Hashcash Algorithm?

Hashcash has been selected for ZenQuote due to three main reasons:


1. Simplicity: Hashcash is a straightforward algorithm that is easily incorporated into our application.

2. Efficiency: Hashcash doesn't require substantial resources while providing an adequate computational challenge.

3. Reliability: Hashcash is a tried-and-true tool, assuring its stable performance in our application.

This makes it the perfect choice for implementing the Proof of Work mechanism in ZenQuote.
