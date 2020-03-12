package certificates

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

const (
	MAXCERTAGE = 5 * 365 * 24 * time.Hour
	MAXLEAFAGE = 24 * time.Hour
	USAGE      = x509.KeyUsageDigitalSignature |
		x509.KeyUsageContentCommitment |
		x509.KeyUsageKeyEncipherment |
		x509.KeyUsageDataEncipherment |
		x509.KeyUsageKeyAgreement |
		x509.KeyUsageCertSign |
		x509.KeyUsageCRLSign
	CERTTYPE         = "CERTIFICATE"
	KEYTYPE          = "ECDSA PRIVATE KEY"
	ROOTCERTFILENAME = "rootCert.cert"
	ROOTKEYFILENAME  = "rootKey.pem"
	PERMISSIONS      = 0644
)

type Cert struct {
	tlsCert tls.Certificate
}

var rootCertificate Cert

func GetRootCertificate() Cert {
	err := LoadRootCertificate()
	if err != nil {
		err = CreateRootCertificate()
	}
	rootCertificate.tlsCert.Leaf, err = x509.ParseCertificate(rootCertificate.tlsCert.Certificate[0])
	return rootCertificate
}

func LoadRootCertificate() error {
	rootCertFile, err := ioutil.ReadFile(ROOTCERTFILENAME)
	if err != nil {
		log.Println("Couldn't read cert file")
	}
	rootKeyFile, err := ioutil.ReadFile(ROOTKEYFILENAME)
	if err != nil {
		log.Println("Couldn't read key file")
	}
	if err != nil {
		return err
	}
	certificate, err := tls.X509KeyPair(rootCertFile, rootKeyFile)
	if err != nil {
		return err
	}
	rootCertificate = Cert{
		tlsCert: certificate,
	}
	return nil
}

func CreateRootCertificate() error {
	// first - generate key
	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		log.Println("Couldn't generate key")
		return err
	}
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		log.Println("Couldn't generate serial number")
		return err
	}
	hostname, err := os.Hostname()
	if err != nil {
		log.Println("Couldn't get hostname")
		return err
	}
	certStruct := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: hostname},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(MAXCERTAGE),
		KeyUsage:              USAGE,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
		SignatureAlgorithm:    x509.ECDSAWithSHA512,
	}
	certDerFile, err := x509.CreateCertificate(rand.Reader, certStruct, certStruct, key.Public(), key)
	if err != nil {
		log.Println("Couldn't create der type certificate")
		return err
	}
	keyDerFile, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		log.Println("Couldn't create der type key")
		return err
	}
	certPemFile := pem.EncodeToMemory(&pem.Block{
		Type:  CERTTYPE,
		Bytes: certDerFile,
	})
	keyPemFile := pem.EncodeToMemory(&pem.Block{
		Type:  KEYTYPE,
		Bytes: keyDerFile,
	})

	certificate, err := tls.X509KeyPair(certPemFile, keyPemFile)
	if err != nil {
		return err
	}
	rootCertificate = Cert{
		tlsCert: certificate,
	}

	err = ioutil.WriteFile(ROOTCERTFILENAME, certPemFile, PERMISSIONS)
	if err == nil {
		err = ioutil.WriteFile(ROOTKEYFILENAME, keyPemFile, PERMISSIONS)
	}
	return nil
}

func CreateLeafCertificate(hosts ...string) (*tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		log.Println("Couldn't generate key")
		return nil, err
	}
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		log.Println("Couldn't generate serial number")
		return nil, err
	}
	certStruct := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: hosts[0]},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(MAXLEAFAGE),
		KeyUsage:              USAGE,
		BasicConstraintsValid: true,
		SignatureAlgorithm:    x509.ECDSAWithSHA512,
	}
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			certStruct.IPAddresses = append(certStruct.IPAddresses, ip)
		} else {
			certStruct.DNSNames = append(certStruct.DNSNames, h)
		}
	}

	certDerFile, err := x509.CreateCertificate(rand.Reader, certStruct, rootCertificate.tlsCert.Leaf, key.Public(), rootCertificate.tlsCert.PrivateKey)
	if err != nil {
		log.Println("Couldn't create der type certificate")
		return nil, err
	}
	certLeafFile, err := x509.ParseCertificate(certDerFile)
	if err != nil {
		log.Println("Couldn't create leaf certificate")
		return nil, err
	}

	tlsCert := &tls.Certificate{
		Certificate: [][]byte{certDerFile},
		PrivateKey:  key,
		Leaf:        certLeafFile,
	}

	return tlsCert, nil

}
