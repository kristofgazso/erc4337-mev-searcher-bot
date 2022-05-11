package main

import (
	"context"
	"erc4337-mev-searcher-bot/abis"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"log"
	"math/big"
)

var (
	safeEntryPoints = []common.Address{
		common.HexToAddress("0x602ab3881ff3fa8da60a8f44cf633e91ba1fdb69"),
		common.HexToAddress("0xD3ebD80aFf10B54795708Ece1d6D3253e3258A05"),
	}
    httpClient   *ethclient.Client
    rpcClient    *rpc.Client
	wallet       Wallet
	chainId      *big.Int
	configToUse = params.GoerliChainConfig
)

/*
type UserOperation struct {
	Sender               common.Address
	Nonce                *big.Int
	InitCode             []byte
	CallData             []byte
	CallGas              *big.Int
	VerificationGas      *big.Int
	PreVerificationGas   *big.Int
	MaxFeePerGas         *big.Int `abi:"maxFeePerGas"`
	MaxPriorityFeePerGas *big.Int `abi:"maxPriorityFeePerGas"`
	Paymaster            common.Address
	PaymasterData        []byte
	Signature            []byte
}
 */

type UserOperationWithEntryPoint struct {
	UserOperation abis.UserOperation
	EntryPoint common.Address
}

type UserOperationJSON struct {
	UserOperation struct {
		Sender               common.Address `json:"sender"`
		Nonce                string `json:"nonce"`
		CallData             string `json:"callData"`
		InitCode             string `json:"initCode"`
		CallGas              string `json:"callGas"`
		VerificationGas      string `json:"verificationGas"`
		PreVerificationGas   string `json:"preVerificationGas"`
		MaxFeePerGas         string `json:"maxFeePerGas"`
		MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas"`
		Paymaster            common.Address `json:"paymaster"`
		Signature            string `json:"signature"`
		PaymasterData        string `json:"paymasterData"`
	} `json:"userOperation"`
	EntryPoint common.Address `json:"entryPoint"`
}

func main() {
	wallet = InitWallet()
	rpcClient, httpClient = InitClients()
	chainId, _ = httpClient.ChainID(context.Background())

	// make websocket subscription
	opJSONchannel := make(chan UserOperationJSON, 100)
	extraParam := make(map[string]interface{})
	extraParam["entryPoints"] = safeEntryPoints
	extraParam["includeUserOperations"] = true
	sub, err := rpcClient.EthSubscribe(
		context.Background(),
		opJSONchannel,
		"newPendingUserOperations",
		extraParam,
	)
	if err != nil {
		log.Fatal(err)
	}

	for jsonOp := range opJSONchannel {
		op := ConvertFromJSONToNative(jsonOp)
		HandleOp(op)
	}

	sub.Unsubscribe()
}


func HandleOp(opWithEp UserOperationWithEntryPoint) {
	op := opWithEp.UserOperation
	ep := opWithEp.EntryPoint

	log.Printf("new pending op with sender %s and nonce %s\n", op.Sender, op.Nonce)

	if !GoodEntryPoint(ep) {
		return
	}

	log.Printf("pending op has safe entryPoint %s\n", ep)


	entryPointAbi, err := abis.NewEntryPoint(ep, httpClient)
	if err != nil {
		log.Fatal(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(wallet.PK, chainId)
	if err != nil {
		log.Fatal(err)
	}

	auth.GasTipCap = big.NewInt(3e9)
	auth.GasFeeCap = big.NewInt(100e9) // YOU SHOULD PROBABLY OPTIMIZE THIS

	baseFee := GetBaseFee(httpClient)
	effectiveGasPrice := math.BigMin(new(big.Int).Add(op.MaxPriorityFeePerGas, baseFee), op.MaxFeePerGas)

	if auth.GasFeeCap.Cmp(effectiveGasPrice) > 0 {
		log.Printf("op doesn't pay enough, op effective gas price: %s, tx gas price: %s\n", effectiveGasPrice, auth.GasPrice)
		// we would lose money, so normally we would not submit it normally but this is good for testing
		// return
		log.Printf("but we are submitting anyway to test\n")
	} else {
		log.Printf("op fee adequate, op effective gas price: %s, tx gas price: %s\n", effectiveGasPrice, auth.GasPrice)
	}


	tx, err := entryPointAbi.HandleOps(auth, []abis.UserOperation{op}, wallet.Address)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Submitted op tx with hash", tx.Hash())

	rcp, err := bind.WaitMined(context.Background(), httpClient, tx)
	if err != nil {
		log.Fatal(err)
	}

	if rcp.Status == 1 {
		log.Println("Success!")
	} else {
		log.Println("Failure!")
	}
}

func ConvertFromJSONToNative(json UserOperationJSON) UserOperationWithEntryPoint {
	nonce, _ := hexutil.DecodeBig(json.UserOperation.Nonce)
	initCode, _ := hexutil.Decode(json.UserOperation.InitCode)
	callData, _ := hexutil.Decode(json.UserOperation.CallData)
	callGas, _ := hexutil.DecodeBig(json.UserOperation.CallGas)
	verificationGas, _ := hexutil.DecodeBig(json.UserOperation.VerificationGas)
	preVerificationGas, _ := hexutil.DecodeBig(json.UserOperation.PreVerificationGas)
	maxFeePerGas, _ := hexutil.DecodeBig(json.UserOperation.MaxFeePerGas)
	maxPriorityFeePerGas, _ := hexutil.DecodeBig(json.UserOperation.MaxPriorityFeePerGas)
	paymasterData, _ := hexutil.Decode(json.UserOperation.PaymasterData)
	signature, _ := hexutil.Decode(json.UserOperation.Signature)

	return UserOperationWithEntryPoint{
		UserOperation: abis.UserOperation{
			Sender:               json.UserOperation.Sender,
			Nonce:                nonce,
			InitCode:             initCode,
			CallData:             callData,
			CallGas:              callGas,
			VerificationGas:      verificationGas,
			PreVerificationGas:   preVerificationGas,
			MaxFeePerGas:         maxFeePerGas,
			MaxPriorityFeePerGas: maxPriorityFeePerGas,
			Paymaster:            json.UserOperation.Paymaster,
			PaymasterData:        paymasterData,
			Signature:            signature,
		},
		EntryPoint:    json.EntryPoint,
	}
}

