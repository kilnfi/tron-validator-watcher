package tron

import (
	"context"
	"fmt"

	"crypto/sha256"
	"encoding/hex"

	"github.com/btcsuite/btcutil/base58"
)

func GetFirstBlockNum(ctx context.Context, client *Client, blockNum int64, blockTimestamp int64) (*Block, error) {
	var targetBlock *Block

	blockTimestampSeconds := blockTimestamp / 1000
	roundStartTimestamp := (blockTimestampSeconds - (blockTimestampSeconds % RoundDuration)) * 1000
	estimatedFirstBlockNum := blockNum - ((blockTimestampSeconds % RoundDuration) / BlockTime)

	// We have calculated an estimated block ID for the first block of the round based on its expected timestamp.
	// However, this estimation may be inaccurate due to possible cold periods caused by the maintenance window
	// or if a producer missed blocks during the round.
	// To find the correct block ID efficiently, we use a binary search, which allows us to quickly locate
	// the first block with a timestamp greater than or equal to the expected round start timestamp.
	// This approach significantly reduces the number of API requests compared to a sequential search.
	low := estimatedFirstBlockNum
	high := blockNum
	for low <= high {
		mid := (low + high) / 2
		block, err := client.Network.GetBlockByNumber(ctx, mid)
		if err != nil {
			return nil, fmt.Errorf("GetFirstBlock: unable to fetch block: %v", err)
		}

		if block.BlockHeader.RawData.Timestamp == roundStartTimestamp {
			targetBlock = block
			break
		}

		if block.BlockHeader.RawData.Timestamp < roundStartTimestamp {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	if targetBlock == nil {
		return nil, fmt.Errorf("GetFirstBlock: unable to find the first block for the current round")
	}

	return targetBlock, nil
}

// ConvertAddressToHex converts a Base58 address to a hex address by following the
// specs from Tron network
func ConvertAddressToHex(address string) string {
	decoded := base58.Decode(address)

	// Remove the last 4 bytes (checksum)
	// This is how thr is supposed to be done according to the Tron network
	hexAddress := hex.EncodeToString(decoded[:len(decoded)-4])
	return hexAddress
}

func ConvertAddressToBase58(address string) (string, error) {
	decodedHex, err := hex.DecodeString(address)
	if err != nil {
		return "", fmt.Errorf("ConvertAddressToBase58: unable to decode hex address: %v", err)
	}

	// Double hash for the address
	hash := doubleHashSHA256(decodedHex)

	// Add the first 4 bytes of the address
	checksum := hash[:4]
	decodedHex = append(decodedHex, checksum...)

	encodedBase58 := base58.Encode(decodedHex)

	return encodedBase58, nil
}

func doubleHashSHA256(data []byte) []byte {
	hash0 := sha256.Sum256(data)     // Premier hash
	hash1 := sha256.Sum256(hash0[:]) // Deuxième hash
	return hash1[:]
}
