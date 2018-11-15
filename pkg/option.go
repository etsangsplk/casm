package casm

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/lthibault/casm/pkg/transport/quic"
)

// Option represents a setting
type Option func(*basicHost) Option

// OptListenAddr sets the listen address
func OptListenAddr(addr string) Option {
	return func(h *basicHost) (prev Option) {
		h.addr.addr = addr
		return
	}
}

// OptTransport sets the transport
func OptTransport(t Transport) Option {
	return func(h *basicHost) (prev Option) {
		h.t = t
		return
	}
}

func optSetID() Option {
	return func(h *basicHost) (prev Option) {
		prev = optSetID()
		h.addr.IDer = NewID()
		return
	}
}

func defaultOpts(qc *quic.Config, tc *tls.Config) []Option {
	t := quic.NewTransport(quic.OptQuic(qc), quic.OptTLS(tc))
	return []Option{
		optSetID(),
		OptListenAddr("localhost:1987"),
		OptTransport(t),
	}
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}

	// generate a random serial number (a real cert authority would have some logic behind this)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		panic("failed to generate serial number: " + err.Error())
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		// SignatureAlgorithm:    x509.ECDSAWithSHA512,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 87600), // in 10 years
		BasicConstraintsValid: true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}, InsecureSkipVerify: true}
}
