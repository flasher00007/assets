# TRC20 USDT Sender

A web application to send TRC20 USDT on the TRON network with a fixed 1 TRX transaction fee.

## Features

- ✅ Send any amount of TRC20 USDT
- ✅ Fixed 1 TRX transaction fee (automatically included)
- ✅ No balance checking required
- ✅ Simple web interface
- ✅ Direct transaction creation from private key
- ✅ Real-time transaction status
- ✅ Transaction ID with TronScan link

## Requirements

- Go 1.19 or higher
- Internet connection (to connect to TRON network)

## Installation

1. Navigate to the project directory:
```bash
cd trc20-sender
```

2. Download dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o trc20-sender main.go
```

## Usage

1. Start the server:
```bash
./trc20-sender
```

2. Open your browser and navigate to:
```
http://localhost:8090
```

3. Fill in the form:
   - **Private Key**: Your TRON wallet private key (without 0x prefix)
   - **Recipient Address**: The TRON address to send USDT to (starts with T)
   - **Amount**: The amount of USDT to send (e.g., 10.5)

4. Click "SEND USDT" button

5. Wait for the transaction to be processed

6. Once successful, you'll receive a transaction ID with a link to view it on TronScan

## Security Warning

⚠️ **IMPORTANT**: Never share your private key with anyone. This application processes transactions locally, but always ensure you're running it in a secure environment.

## Technical Details

- **USDT Contract Address**: TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t (TRON Mainnet)
- **Transaction Fee**: 1 TRX (1,000,000 SUN)
- **USDT Decimals**: 6
- **Network**: TRON Mainnet
- **API Endpoint**: https://api.trongrid.io

## How It Works

1. The application takes your private key and derives the sender address
2. Creates a TRC20 transfer transaction using the USDT contract
3. Signs the transaction with your private key
4. Broadcasts the signed transaction to the TRON network
5. Returns the transaction ID for tracking

## API Endpoint

The application exposes a REST API endpoint:

### POST /api/transfer

Request body:
```json
{
  "privateKey": "your_private_key_here",
  "toAddress": "TRecipientAddressHere",
  "amount": "10.5"
}
```

Response (Success):
```json
{
  "success": true,
  "txId": "transaction_id_here"
}
```

Response (Error):
```json
{
  "success": false,
  "error": "error_message_here"
}
```

## Troubleshooting

### Port Already in Use
If port 8090 is already in use, edit `main.go` and change the port number:
```go
port := "8090"  // Change to another port like "8091"
```

### Transaction Failed
- Ensure you have enough TRX in your wallet for the fee (at least 1 TRX)
- Verify the recipient address is valid and starts with 'T'
- Check that your private key is correct
- Ensure you have USDT balance in your wallet

### Connection Issues
- Check your internet connection
- Verify that api.trongrid.io is accessible
- Try again after a few moments if the network is congested

## License

MIT License

## Disclaimer

This software is provided "as is", without warranty of any kind. Use at your own risk. Always verify transaction details before sending.
