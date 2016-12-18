package client

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
)

var gTLSConfig *tls.Config

func tlsConfig(certPemPath, keyPemPath, caCertPath string) (*tls.Config, error) {
	if gTLSConfig == nil {
		mainCert, err := tls.LoadX509KeyPair(certPemPath, keyPemPath)
		if err != nil {
			return nil, err
		}

		pemBytes, err := ioutil.ReadFile(caCertPath)
		if err != nil {
			return nil, err
		}

		roots := x509.NewCertPool()
		roots.AppendCertsFromPEM(pemBytes)

		gTLSConfig = &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			Certificates:             []tls.Certificate{mainCert},
			PreferServerCipherSuites: true,
			RootCAs:                  roots,
			ClientCAs:                roots,
			ClientAuth:               tls.RequireAndVerifyClientCert,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		}
	}

	return gTLSConfig, nil
}
