package main

import "testing"

func TestIsExpiring(t *testing.T) {
	testCases := []struct {
		name     string
		keystore string
		pemFile  string
		want     bool
	}{
		{"old", "certs/old/keystore.jks", "certs/old/consumer-sa.crt", true},
		{"new", "certs/new/keystore.jks", "certs/new/consumer-sa.crt", false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cert, err := parsePEMFile(tc.pemFile)
			if err != nil {
				t.Error(err)
			}
			if got := isExpiring(cert); got != tc.want {
				t.Errorf("got %v; wanted %v\n", got, tc.want)
			}
			certs, err := parseKeystore(tc.keystore, "certs/storepass")
			if err != nil {
				t.Error(err)
			}
			for _, cert := range certs {
				got := isExpiring(cert)
				if got != tc.want {
					t.Errorf("got %v; wanted %v\n", got, tc.want)
				}
			}
		})
	}
}
