# Searcher Bot for ERC-4337 

Disclaimer: This is only built for Goerli for now, ERC-4337 for Mainnet will come Soon™. But better to prepare yourself for when it arrives

Disclaimer: I take no responsibility for whatever you do with this. Definitely not audited and there aren't even any unit tests. Use a seperate private key with only Goerli ETH on it, don't be dumb.


## Explainer

[ERC-4337](https://eips.ethereum.org/EIPS/eip-4337) is a proposal to enable account abstraction on Ethereum without requiring a hard-fork.

It builds on top of the already existing mempool, creating the concept of `User Operation` objects. 

These `User Operation` objects must be routed through an appropriate [EntryPoint smart contract](https://github.com/eth-infinitism/account-abstraction/blob/main/contracts/EntryPoint.sol) by bundlers, which will ensure that the wallet dealing with the `User Operation` compensates the bundler for including the `User Operation`.

Most `User Operation` objects will usually only succeed when routed through the EntryPoint they were meant for.

The way these `User Operation` objects are broadcast is via p2p, on a separate protocol. The [Nethermind](https://github.com/NethermindEth/nethermind) client has full support for this protocol.

By bundling `User Operation` objects into transactions and submitting them on chain, you can be compensated if the `User Operation` fee is high enough. Beware, the Nethermind Goerli Validator is also trying to bundle and submit transactions, however it will only do so when it is its turn to propose a new block. By running a script like this, you can usually include the `User Operation` yourself and claim the fee.

## How to run this script

First, you must download and build the latest version of [Nethermind](https://github.com/NethermindEth/nethermind).

Then, run it with the config `goerli_aa` to enable the `User Operation` pool, the relevant p2p protocol and websocket subscriptions.

Wait for it to sync to Goerli head (including the state sync), this may take a bit


Fill in the .env file with the PRIVATE_KEY env variable, without the 0x at the start. This wallet will be used to bundle and send. Make sure it has Goerli ETH on it.

Then, build and run this go script. This script will subscribe to any new `User Operation` objects entering the Nethermind node's pool, and submit them on chain if they pay enough.

WARNING: The logic to only submit if the `User Operation` fee is high enough is turned off, since this is Goerli and it is useful for testing.

You can submit `User Operation` objects yourself for testing with the help of [this](https://medium.com/nethermind-eth/erc-4337-account-abstraction-is-already-here-e9588b789e15) article.