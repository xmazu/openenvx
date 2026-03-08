package proxy

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	caValidityYears  = 10
	certValidityDays = 365
)

var (
	hostCertsMu sync.Mutex
	hostCerts   = make(map[string]*certKeyPair)
)

type certKeyPair struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
}

// ensureCA creates CA key and cert in stateDir if they don't exist.
// Returns (caCert, caKey, nil) or error.
func ensureCA(stateDir string) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	caKeyPath := filepath.Join(stateDir, "ca-key.pem")
	caCertPath := filepath.Join(stateDir, "ca.pem")

	// Load existing if present
	if keyPEM, err := os.ReadFile(caKeyPath); err == nil {
		if certPEM, err := os.ReadFile(caCertPath); err == nil {
			block, _ := pem.Decode(keyPEM)
			if block != nil {
				key, err := x509.ParseECPrivateKey(block.Bytes)
				if err == nil {
					block, _ := pem.Decode(certPEM)
					if block != nil {
						cert, err := x509.ParseCertificate(block.Bytes)
						if err == nil && time.Until(cert.NotAfter) > 24*time.Hour {
							return cert, key, nil
						}
					}
				}
			}
		}
	}

	// Generate new CA
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate CA key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber:          randSerial(),
		Subject:               pkix.Name{Organization: []string{"OpenEnvX CA"}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(caValidityYears, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, fmt.Errorf("create CA cert: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("parse CA cert: %w", err)
	}

	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, nil, err
	}
	if err := writePEM(caKeyPath, "EC PRIVATE KEY", keyBytes); err != nil {
		return nil, nil, err
	}
	if err := writePEM(caCertPath, "CERTIFICATE", certDER); err != nil {
		return nil, nil, err
	}
	_ = os.Remove(filepath.Join(filepath.Dir(caCertPath), trustMarkerFile))

	return cert, key, nil
}

// ensureServerCert creates default server cert for localhost + *.localhost.
func ensureServerCert(stateDir string, caCert *x509.Certificate, caKey *ecdsa.PrivateKey) (certPEM, keyPEM []byte, err error) {
	serverCertPath := filepath.Join(stateDir, "server.pem")
	serverKeyPath := filepath.Join(stateDir, "server-key.pem")

	if c, err := os.ReadFile(serverCertPath); err == nil {
		if k, err := os.ReadFile(serverKeyPath); err == nil {
			if cert, _ := parseCertPEM(c); cert != nil && time.Until(cert.NotAfter) > 24*time.Hour {
				return c, k, nil
			}
		}
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate server key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber:          randSerial(),
		Subject:               pkix.Name{Organization: []string{"OpenEnvX Server"}},
		DNSNames:              []string{"localhost", "*.localhost"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, certValidityDays),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, caCert, &key.PublicKey, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("create server cert: %w", err)
	}

	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, nil, err
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	if err := os.WriteFile(serverCertPath, certPEM, 0644); err != nil {
		return nil, nil, err
	}
	if err := os.WriteFile(serverKeyPath, keyPEM, 0600); err != nil {
		return nil, nil, err
	}

	return certPEM, keyPEM, nil
}

// GetCertForHost returns a TLS certificate for the given hostname (e.g. myapp.localhost).
// Uses cache in stateDir/host-certs/ for per-host certs (RFC 2606: *.localhost edge case).
func GetCertForHost(stateDir string, hostname string, caCert *x509.Certificate, caKey *ecdsa.PrivateKey) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	hostname = strings.ToLower(strings.TrimSpace(hostname))
	if hostname == "" {
		return nil, nil, fmt.Errorf("empty hostname")
	}

	hostCertsMu.Lock()
	if p, ok := hostCerts[hostname]; ok {
		hostCertsMu.Unlock()
		return p.cert, p.key, nil
	}
	hostCertsMu.Unlock()

	certsDir := filepath.Join(stateDir, "host-certs")
	if err := os.MkdirAll(certsDir, 0755); err != nil {
		return nil, nil, err
	}

	safe := strings.ReplaceAll(hostname, ".", "_")
	certPath := filepath.Join(certsDir, safe+".pem")
	keyPath := filepath.Join(certsDir, safe+"-key.pem")

	if certPEM, err := os.ReadFile(certPath); err == nil {
		if keyPEM, err := os.ReadFile(keyPath); err == nil {
			cert, err := parseCertPEM(certPEM)
			if err == nil && cert != nil && time.Until(cert.NotAfter) > 24*time.Hour {
				key, err := parseKeyPEM(keyPEM)
				if err == nil && key != nil {
					hostCertsMu.Lock()
					hostCerts[hostname] = &certKeyPair{cert: cert, key: key}
					hostCertsMu.Unlock()
					return cert, key, nil
				}
			}
		}
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate host key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber:          randSerial(),
		Subject:               pkix.Name{Organization: []string{"OpenEnvX " + hostname}},
		DNSNames:              []string{hostname},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, certValidityDays),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, caCert, &key.PublicKey, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("create host cert: %w", err)
	}

	cert, _ := x509.ParseCertificate(certDER)
	keyBytes, _ := x509.MarshalECPrivateKey(key)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	_ = os.WriteFile(certPath, certPEM, 0644)
	_ = os.WriteFile(keyPath, keyPEM, 0600)

	hostCertsMu.Lock()
	hostCerts[hostname] = &certKeyPair{cert: cert, key: key}
	hostCertsMu.Unlock()

	return cert, key, nil
}

// trustMarkerPath is written after CA is successfully trusted. Skip TrustCA if it exists.
const trustMarkerFile = "ca-trusted"

// TrustCA adds the CA to the system trust store (only once per CA; uses marker file).
// macOS: security add-trusted-cert (uses login keychain)
// Linux: copy to /usr/local/share/ca-certificates/ and update-ca-certificates
func TrustCA(stateDir string) error {
	markerPath := filepath.Join(stateDir, trustMarkerFile)
	if _, err := os.Stat(markerPath); err == nil {
		return nil // already trusted
	}

	caPath := filepath.Join(stateDir, "ca.pem")
	if _, err := os.Stat(caPath); err != nil {
		return fmt.Errorf("CA not found: %w", err)
	}

	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		keychain := filepath.Join(home, "Library", "Keychains", "login.keychain-db")
		if _, err := os.Stat(keychain); err != nil {
			keychain = "/Library/Keychains/System.keychain"
		}
		cmd := exec.Command("security", "add-trusted-cert", "-r", "trustRoot", "-k", keychain, caPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("add trusted cert (may need sudo): %w", err)
		}
	case "linux":
		dest := "/usr/local/share/ca-certificates/oexctl-ca.crt"
		data, err := os.ReadFile(caPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dest, data, 0644); err != nil {
			return fmt.Errorf("copy CA to %s (run with sudo): %w", dest, err)
		}
		cmd := exec.Command("update-ca-certificates")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("update-ca-certificates: %w", err)
		}
	default:
		return fmt.Errorf("TrustCA not supported on %s", runtime.GOOS)
	}
	_ = os.WriteFile(markerPath, []byte{}, 0644)
	return nil
}

// EnsureCerts creates CA and server certs if needed. Returns stateDir for use with TrustCA/GetCertForHost.
func EnsureCerts(proxyPort int) (stateDir string, err error) {
	stateDir, err = StateDirPath(proxyPort)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return "", err
	}
	caCert, caKey, err := ensureCA(stateDir)
	if err != nil {
		return "", err
	}
	_, _, err = ensureServerCert(stateDir, caCert, caKey)
	if err != nil {
		return "", err
	}
	return stateDir, nil
}

func randSerial() *big.Int {
	b := make([]byte, 16)
	rand.Read(b)
	return new(big.Int).SetBytes(b)
}

func writePEM(path string, typ string, data []byte) error {
	block := &pem.Block{Type: typ, Bytes: data}
	return os.WriteFile(path, pem.EncodeToMemory(block), 0600)
}

func parseCertPEM(pemData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("no PEM block")
	}
	return x509.ParseCertificate(block.Bytes)
}

func parseKeyPEM(pemData []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("no PEM block")
	}
	return x509.ParseECPrivateKey(block.Bytes)
}
