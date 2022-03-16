package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/joho/godotenv"
	"log"
	"math/big"
	"os"
)

func InitWallet() Wallet {
	godotenv.Load()
	privateKeyString := os.Getenv("PRIVATE_KEY")

	pk, err := crypto.HexToECDSA(privateKeyString)
	if err != nil {
		fmt.Println("Failed to decode private key, make sure you do not put 0x in front of the hex")
		log.Fatal(err)
	}
	pubZk := pk.Public()
	pubZkECDSA := pubZk.(*ecdsa.PublicKey)
	address := crypto.PubkeyToAddress(*pubZkECDSA)

	return Wallet{
		PK:      pk,
		Address: address,
	}
}

type Wallet struct {
	PK      *ecdsa.PrivateKey
	Address common.Address
}

func InitClients() (*rpc.Client, *ethclient.Client){
	rpcClient, err := rpc.Dial("ws://localhost:8546")
	if err != nil {
		log.Fatal(err)
	}

	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal(err)
	}

	return rpcClient, client
}

func GetBaseFee(client *ethclient.Client) *big.Int {
	latestBlock, _ := client.BlockByNumber(context.Background(), nil)
	return misc.CalcBaseFee(configToUse, latestBlock.Header())
}

func GoodEntryPoint(ep common.Address) bool {
	goodEntryPoint := false
	for _, safeEp := range safeEntryPoints {
		if safeEp == ep {
			goodEntryPoint = true
			break
		}
	}
	return goodEntryPoint
}