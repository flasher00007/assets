package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// USDT TRC20 contract address on TRON mainnet
	USDTContractAddress = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
	// TRON API endpoint
	TronAPIURL = "https://api.trongrid.io"
)

type TransferRequest struct {
	PrivateKey string `json:"privateKey"`
	ToAddress  string `json:"toAddress"`
	Amount     string `json:"amount"`
}

type TransferResponse struct {
	Success bool   `json:"success"`
	TxID    string `json:"txId,omitempty"`
	Error   string `json:"error,omitempty"`
}

type TronTransaction struct {
	TxID      string                 `json:"txID"`
	RawData   map[string]interface{} `json:"raw_data"`
	RawDataHex string                `json:"raw_data_hex"`
}

type TronBroadcastResponse struct {
	Result  bool   `json:"result"`
	TxID    string `json:"txid"`
	Message string `json:"message,omitempty"`
}

func main() {
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/api/transfer", handleTransfer)

	port := "8090"
	log.Printf("Server starting on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func handleTransfer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		json.NewEncoder(w).Encode(TransferResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(TransferResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	txID, err := sendTRC20USDT(req.PrivateKey, req.ToAddress, req.Amount)
	if err != nil {
		log.Printf("Transfer error: %v", err)
		json.NewEncoder(w).Encode(TransferResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(TransferResponse{
		Success: true,
		TxID:    txID,
	})
}

func sendTRC20USDT(privateKeyHex, toAddress, amountStr string) (string, error) {
	// Remove 0x prefix if present
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")

	// Parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %v", err)
	}

	// Get sender address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("failed to get public key")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fromAddressTron := hexToTronAddress(fromAddress.Hex())

	// Parse amount (USDT has 6 decimals)
	amount := new(big.Float)
	amount.SetString(amountStr)
	multiplier := new(big.Float).SetFloat64(1000000) // 10^6 for USDT decimals
	amountInSmallestUnit := new(big.Float).Mul(amount, multiplier)
	amountInt, _ := amountInSmallestUnit.Int(nil)

	// Prepare TRC20 transfer function call
	// Function signature: transfer(address,uint256)
	// Method ID: a9059cbb
	methodID := "a9059cbb"

	// Encode recipient address (remove T prefix and convert to hex, pad to 32 bytes)
	toAddrHex := tronAddressToHex(toAddress)
	toAddrPadded := fmt.Sprintf("%064s", toAddrHex)

	// Encode amount (pad to 32 bytes)
	amountHex := fmt.Sprintf("%064s", amountInt.Text(16))

	// Combine method ID + parameters
	parameter := methodID + toAddrPadded + amountHex

	// Create transaction
	tx, err := createTRC20Transaction(fromAddressTron, USDTContractAddress, parameter)
	if err != nil {
		return "", fmt.Errorf("failed to create transaction: %v", err)
	}

	// Sign transaction
	signedTx, err := signTransaction(tx, privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %v", err)
	}

	// Broadcast transaction
	txID, err := broadcastTransaction(signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to broadcast transaction: %v", err)
	}

	return txID, nil
}

func createTRC20Transaction(ownerAddress, contractAddress, parameter string) (*TronTransaction, error) {
	payload := map[string]interface{}{
		"owner_address":     ownerAddress,
		"contract_address":  contractAddress,
		"function_selector": "transfer(address,uint256)",
		"parameter":         parameter,
		"fee_limit":         1000000, // 1 TRX
		"call_value":        0,
		"visible":           true,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(TronAPIURL+"/wallet/triggersmartcontract", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result["Error"] != nil {
		return nil, fmt.Errorf("API error: %v", result["Error"])
	}

	transaction := result["transaction"].(map[string]interface{})
	
	tx := &TronTransaction{
		RawData: transaction["raw_data"].(map[string]interface{}),
	}

	// Convert raw_data to hex
	rawDataBytes, _ := json.Marshal(tx.RawData)
	tx.RawDataHex = hex.EncodeToString(rawDataBytes)

	return tx, nil
}

func signTransaction(tx *TronTransaction, privateKeyHex string) (map[string]interface{}, error) {
	// Serialize raw_data
	rawDataBytes, err := json.Marshal(tx.RawData)
	if err != nil {
		return nil, err
	}

	// Hash the raw data
	hash := sha256.Sum256(rawDataBytes)

	// Sign with private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}

	signature, err := crypto.Sign(hash[:], privateKey)
	if err != nil {
		return nil, err
	}

	// Remove recovery ID (last byte) for TRON
	if len(signature) == 65 {
		signature = signature[:64]
	}

	signedTx := map[string]interface{}{
		"raw_data":  tx.RawData,
		"signature": []string{hex.EncodeToString(signature)},
		"visible":   true,
	}

	return signedTx, nil
}

func broadcastTransaction(signedTx map[string]interface{}) (string, error) {
	jsonData, err := json.Marshal(signedTx)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(TronAPIURL+"/wallet/broadcasttransaction", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result TronBroadcastResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if !result.Result {
		return "", fmt.Errorf("broadcast failed: %s", result.Message)
	}

	return result.TxID, nil
}

func hexToTronAddress(hexAddr string) string {
	// This is a simplified conversion - in production use proper base58check encoding
	hexAddr = strings.TrimPrefix(hexAddr, "0x")
	return "T" + hexAddr[:40]
}

func tronAddressToHex(tronAddr string) string {
	// Remove T prefix and return hex
	// This is simplified - in production use proper base58check decoding
	return strings.TrimPrefix(tronAddr, "T")
}
