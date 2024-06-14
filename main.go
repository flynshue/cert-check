package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pavlo-v-chernykh/keystore-go/v4"
	"github.com/spf13/viper"
)

func main() {
	for {
		kafkaCert, err := parsePEMFile(viper.GetString("certFile"))
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		isExpiring(kafkaCert)
		keystoreCerts, err := parseKeystore(viper.GetString("keystore"), viper.GetString("storepass"))
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		for _, certs := range keystoreCerts {
			isExpiring(certs)
		}
		time.Sleep(time.Second * 20)
	}
}

func parsePEMFile(name string) (*x509.Certificate, error) {
	b, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	pemBlock, _ := pem.Decode(b)
	return x509.ParseCertificate(pemBlock.Bytes)
}

func parseKeystore(keystoreFilename, passFile string) ([]*x509.Certificate, error) {
	keyFile, err := os.Open(keystoreFilename)
	if err != nil {
		return nil, err
	}
	defer keyFile.Close()
	b, err := os.ReadFile(passFile)
	if err != nil {
		return nil, err
	}
	jks := keystore.New()
	if err := jks.Load(keyFile, b); err != nil {
		return nil, fmt.Errorf("error loading keystore; %s", err)
	}
	x509Certs := []*x509.Certificate{}
	for _, alias := range jks.Aliases() {
		if jks.IsPrivateKeyEntry(alias) {
			certs, err := jks.GetPrivateKeyEntryCertificateChain(alias)
			if err != nil {
				return nil, fmt.Errorf("error getting private key cert %s", err)
			}
			for _, cert := range certs {
				c, err := x509.ParseCertificate(cert.Content)
				if err != nil {
					log.Println(err)
					continue
				}
				x509Certs = append(x509Certs, c)
			}
		}
		if jks.IsTrustedCertificateEntry(alias) {
			cert, err := jks.GetTrustedCertificateEntry(alias)
			if err != nil {
				return nil, err
			}
			c, err := x509.ParseCertificate(cert.Certificate.Content)
			if err != nil {
				return nil, err
			}
			x509Certs = append(x509Certs, c)
		}

	}
	return x509Certs, nil
}

func isExpiring(cert *x509.Certificate) bool {
	warnDate := cert.NotAfter.AddDate(0, 0, -14)
	now := time.Now()
	if warnDate.Unix() <= now.Unix() {
		log.Printf("WARNING: Subject: %s certificate is about to expire %+v\n", cert.Subject.String(), cert.NotAfter)
		return true

	} else {
		log.Printf("Cert OK; Subject %s Expires: %+v\n", cert.Subject.String(), cert.NotAfter)
		return false
	}
}

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("certFile", "certs/new/consumer-sa.crt")
	viper.SetDefault("keystore", "certs/new/keystore.jks")
	viper.SetDefault("storepass", "certs/storepass")
}
