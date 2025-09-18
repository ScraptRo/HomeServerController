package common

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net"
	"os"
	"runtime"
)

const idFile = "./client_id.txt"

func getHardwareInfo() string {
	// grab MACs
	macs := ""
	ifs, _ := net.Interfaces()
	for _, i := range ifs {
		if len(i.HardwareAddr) > 0 {
			macs += i.HardwareAddr.String()
		}
	}

	// add some system info
	sys := runtime.GOOS + runtime.GOARCH

	return macs + sys
}

func generateID() string {
	data := []byte(getHardwareInfo())
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func GetOrCreateID() string {
	// if file exists, read it
	if b, err := os.ReadFile(idFile); err == nil && len(b) > 0 {
		return string(b)
	}

	// else generate
	id := generateID()
	os.WriteFile(idFile, []byte(id), 0600)
	return id
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func MinimumAddUserGrade() uint8 {
	return 0
}
