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

I picked the Hashcash algorithm for ZenQuote, and here's why:

1. **Simplicity:** Hashcash is simple. It's just easier to plug into our system. No need for fancy setups or configurations.

2. **Efficiency:** Hashcash isn't a resource hog. It offers a fair computational challenge without slowing everything down.

3. **Dependability:** Hashcash has been around. It's not some fresh-out-of-the-oven experiment. It's a tool that we know works, and works well.

When you stack Hashcash up against other algorithms, it's the balance that makes it shine. Some algorithms might be more advanced, but they often need more resources or complicated setups. Hashcash finds the sweet spot between being effective and being resource-friendly.

Sure, Hashcash isn't flawless. The level of challenge it offers needs to be just right. Make it too easy, and it doesn't do its job of preventing spam. Make it too hard, and it might block actual users from our service. But, with the right tuning, this isn't a major worry.

All things considered, Hashcash provides a straightforward, efficient, and dependable method for adding a Proof of Work feature to ZenQuote. It fits our bill, and that's why it's the algorithm we're going with.
