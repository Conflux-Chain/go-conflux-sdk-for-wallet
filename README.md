
[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](https://github.com/Conflux-Chain/go-conflux-sdk/blob/master/LICENSE)
[![Documentation](https://img.shields.io/badge/Documentation-GoDoc-green.svg)](https://godoc.org/github.com/Conflux-Chain/go-conflux-sdk)

# Conflux Golang API For Wallet
This is a SDK for developers who would like to port a wallet with Conflux network.

Conflux has provided [conflux-go-sdk](https://github.com/Conflux-Chain/go-conflux-sdk) to support the operation of communicating with nodes, accounts manager and contract operation. The conflux-go-sdk-for-wallet is developed for the convenience of wallet development, there are some complex APIs are provided through communicate with centralized server. Currently, it is mainly for querying summary of user transactions and token transfer event.

The Conflux Golang API allows any Golang client to interact with a local or remote Conflux node based on JSON-RPC 2.0 protocol. With Conflux Golang API, user can easily manage accounts, send transactions, deploy smart contracts and query blockchain information.

## Install
```
go get github.com/Conflux-Chain/go-conflux-sdk-for-wallet
```
You can also add the Conflux Golang API For Wallet into vendor folder.
```
govendor fetch github.com/Conflux-Chain/go-conflux-sdk-for-wallet
```

## Usage

[api document](https://github.com/Conflux-Chain/go-conflux-sdk-for-wallet/blob/master/api.md)

