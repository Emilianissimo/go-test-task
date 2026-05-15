package helpers

import (
	"fmt"
	"hash/fnv"
	"time"
)

func GenerateFastTxID(walletID int64) string {
	h := fnv.New64a()
	data := fmt.Sprintf("%d-%d", walletID, time.Now().UnixNano())
	_, err := h.Write([]byte(data))
	if err != nil {
		return ""
	}
	return fmt.Sprintf("tx_%x", h.Sum64())
}
