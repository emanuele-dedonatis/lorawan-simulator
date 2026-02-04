package gateway

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/brocaar/lorawan"
)

func formatEUI(eui lorawan.EUI64) string {
	euiHex := hex.EncodeToString(eui[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s-%s-%s-%s",
		euiHex[0:2], euiHex[2:4], euiHex[4:6], euiHex[6:8],
		euiHex[8:10], euiHex[10:12], euiHex[12:14], euiHex[14:16])
}

// formatEUIasID6 converts an EUI64 to ID6 format (similar to IPv6 notation)
// ID6 represents a 64-bit EUI as four 16-bit blocks separated by colons
// Examples:
//   ::0 = 00-00-00-00-00-00-00-00
//   1:: = 00-01-00-00-00-00-00-00
//   f:a123:f8:100 = 00-0f-a1-23-00-f8-01-00
func formatEUIasID6(eui lorawan.EUI64) string {
	euiHex := hex.EncodeToString(eui[:])
	
	// Split into four 16-bit blocks (4 hex chars each)
	blocks := []string{
		euiHex[0:4],   // First 16 bits
		euiHex[4:8],   // Second 16 bits
		euiHex[8:12],  // Third 16 bits
		euiHex[12:16], // Fourth 16 bits
	}
	
	// Remove leading zeros from each block
	for i := range blocks {
		blocks[i] = strings.TrimLeft(blocks[i], "0")
		if blocks[i] == "" {
			blocks[i] = "0"
		}
	}
	
	// Find the longest sequence of consecutive "0" blocks for compression
	maxStart, maxLen := -1, 0
	currentStart, currentLen := -1, 0
	
	for i, block := range blocks {
		if block == "0" {
			if currentStart == -1 {
				currentStart = i
				currentLen = 1
			} else {
				currentLen++
			}
			
			if currentLen > maxLen {
				maxStart = currentStart
				maxLen = currentLen
			}
		} else {
			currentStart = -1
			currentLen = 0
		}
	}
	
	// Build the ID6 string with :: compression if applicable
	if maxLen > 1 {
		result := ""
		
		// Add blocks before the compressed sequence
		if maxStart > 0 {
			result = strings.Join(blocks[0:maxStart], ":")
		}
		
		// Add the compression marker
		if maxStart == 0 && maxStart+maxLen == 4 {
			// All zeros: ::0
			result = "::0"
		} else if maxStart == 0 {
			// Starts with zeros
			result = "::" + strings.Join(blocks[maxStart+maxLen:], ":")
		} else if maxStart+maxLen == 4 {
			// Ends with zeros
			result = result + "::"
		} else {
			// Zeros in the middle
			result = result + "::" + strings.Join(blocks[maxStart+maxLen:], ":")
		}
		
		return result
	}
	
	// No compression needed
	return strings.Join(blocks, ":")
}
