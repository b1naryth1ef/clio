# Clio

Clio is an advanced P2P network which represents a completely distributed and decentralized nature. It uses peer seeding, message passing, and advanced routing to keep a structurally simple and yet robust P2P network. It was written in Go, using OpenGPG as the encryption and identification layer. The intention for building it was to create a dynamic sharing network similar to current social networks, but completely distributed, encrypted, and simple.

### Structure
Each node in a Clio network represents an OpenGPG public key fingerprint, allowing messages to be dynamically encrypted and routed too any node on the network regardless of active connections. Clio also implements a caching and archiving system much like a distributed database to allow for (somewhat) delayed message and data retreival. This allows a workflow similar to a website or email.

## cliod
Cliod is the systems code for the network, including backend API's for joining and working with the network.

## clicli
The frontend interactions code (???)

## TODO List

- Start testing the actual network, this means we need to either generate a bunch of pgp keys or fake them
- Fully implement crates and syncing, write some good logic for this
- Wrap stuff cleanly in a API and build out a frontend for sheeet
- Implement proper routing tables
- Test propagation time on a large network, etc
- Robustness testing, reliability testing
- Figure out the limits of the networking layer
- Refactor/cleanup?

## ETC

- TCP vs UDP? Read up on hole punching/etc
- Proper routing table algorithims