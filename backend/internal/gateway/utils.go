package gateway

import (
	"encoding/hex"
	"fmt"

	"github.com/brocaar/lorawan"
)

func formatEUI(eui lorawan.EUI64) string {
	euiHex := hex.EncodeToString(eui[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s-%s-%s-%s",
		euiHex[0:2], euiHex[2:4], euiHex[4:6], euiHex[6:8],
		euiHex[8:10], euiHex[10:12], euiHex[12:14], euiHex[14:16])
}
