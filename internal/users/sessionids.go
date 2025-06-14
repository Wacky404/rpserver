package users

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

const SessionPrefix string = "rpu_"

// generates an random id for session unique identifier
func GenID(byteSize int) string {
	id := make([]byte, byteSize)
	_, err := rand.Read(id) // this is never supposed to error apparently
	if err != nil {
		return ""
	}
	fmtID := hex.EncodeToString(id)
	return fmtID
}

// hash using sha256
func HashSID(sid string) string {
	h := sha256.New()
	h.Write([]byte(sid))
	hb := h.Sum(nil)
	hs := hex.EncodeToString(hb)
	return hs
}

// verify integrity of sid against database sid
func VerifySID(sid string, dbsid string) bool {
	n := HashSID(sid)
	return n == dbsid
}
