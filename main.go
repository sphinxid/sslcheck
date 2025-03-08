package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// TLS protocol versions mapped to their names
var tlsVersions = map[uint16]string{
	tls.VersionSSL30: "SSL 3.0 (deprecated)",
	tls.VersionTLS10: "TLS 1.0 (deprecated)",
	tls.VersionTLS11: "TLS 1.1 (deprecated)",
	tls.VersionTLS12: "TLS 1.2",
	tls.VersionTLS13: "TLS 1.3",
}

func main() {
	// Parse command line arguments
	host := flag.String("host", "", "Host to check (e.g., example.com)")
	port := flag.String("port", "443", "Port to connect to (default: 443)")
	timeout := flag.Int("timeout", 10, "Connection timeout in seconds")
	verbose := flag.Bool("verbose", false, "Show detailed certificate information")
	flag.Parse()

	if *host == "" {
		fmt.Println("Error: Host is required")
		fmt.Println("Usage: sslcheck -host example.com [-port 443] [-timeout 10] [-verbose]")
		os.Exit(1)
	}

	// Run the SSL check
	err := checkSSL(*host, *port, time.Duration(*timeout)*time.Second, *verbose)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func checkSSL(host, port string, timeout time.Duration, verbose bool) error {
	fmt.Printf("Checking SSL/TLS for %s:%s\n\n", host, port)

	// Check supported protocols
	fmt.Println("=== TLS Protocol Support ===")
	checkTLSVersions(host, port, timeout)
	fmt.Println()

	// Connect using the highest available protocol
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: timeout},
		"tcp",
		fmt.Sprintf("%s:%s", host, port),
		&tls.Config{
			InsecureSkipVerify: true, // We do our own verification
			ServerName:         host,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Get certificate chain
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return fmt.Errorf("no certificates found")
	}

	fmt.Println("=== Certificate Chain ===")
	// Check leaf certificate (server certificate)
	fmt.Println("1. Server Certificate:")
	leafCert := certs[0]
	checkCertificate(leafCert, host, verbose)

	// Check intermediate certificates
	intermediatesPool := x509.NewCertPool()
	for i, cert := range certs[1:] {
		if i == len(certs)-2 { // Last certificate is usually the root
			fmt.Printf("\n3. Root CA Certificate:\n")
		} else {
			fmt.Printf("\n2. Intermediate Certificate %d:\n", i+1)
		}
		checkCertificate(cert, "", verbose)
		intermediatesPool.AddCert(cert)
	}

	// Verify certificate chain
	fmt.Println("\n=== Certificate Chain Verification ===")
	roots := x509.NewCertPool()
	// Try to use system root CA pool
	systemRoots, err := x509.SystemCertPool()
	if err == nil {
		roots = systemRoots
	}

	opts := x509.VerifyOptions{
		DNSName:       host,
		Intermediates: intermediatesPool,
		Roots:         roots,
	}

	_, err = leafCert.Verify(opts)
	if err != nil {
		fmt.Println("❌ Certificate chain verification FAILED")
		fmt.Printf("   Reason: %v\n", err)
	} else {
		fmt.Println("✅ Certificate chain verification PASSED")
	}

	return nil
}

func checkCertificate(cert *x509.Certificate, hostname string, verbose bool) {
	// Check validity period
	now := time.Now()
	validFrom := cert.NotBefore
	validTo := cert.NotAfter
	daysLeft := int(validTo.Sub(now).Hours() / 24)

	fmt.Printf("   Subject: %s\n", cert.Subject.CommonName)
	fmt.Printf("   Issuer: %s\n", cert.Issuer.CommonName)
	fmt.Printf("   Valid from: %s to %s (%d days left)\n", 
		validFrom.Format("2006-01-02"), 
		validTo.Format("2006-01-02"), 
		daysLeft)

	// Check if certificate is valid
	if now.Before(validFrom) {
		fmt.Println("   ❌ Certificate is not yet valid")
	} else if now.After(validTo) {
		fmt.Println("   ❌ Certificate has expired")
	} else {
		fmt.Println("   ✅ Certificate date is valid")
	}

	// Check hostname match for leaf certificate
	if hostname != "" {
		if err := cert.VerifyHostname(hostname); err != nil {
			fmt.Printf("   ❌ Hostname verification FAILED: %v\n", err)
		} else {
			fmt.Println("   ✅ Hostname verification PASSED")
		}
	}

	// Check if certificate is self-signed
	if cert.Issuer.CommonName == cert.Subject.CommonName {
		fmt.Println("   ℹ️ Self-signed certificate detected")
	}

	// Print detailed certificate information if verbose mode is enabled
	if verbose {
		fmt.Println("   --- Detailed Certificate Information ---")
		fmt.Printf("   Serial Number: %X\n", cert.SerialNumber)
		fmt.Printf("   Signature Algorithm: %s\n", cert.SignatureAlgorithm)
		fmt.Printf("   Public Key Algorithm: %s\n", cert.PublicKeyAlgorithm)
		
		// Print Subject Alternative Names
		if len(cert.DNSNames) > 0 {
			fmt.Printf("   DNS Names: %s\n", strings.Join(cert.DNSNames, ", "))
		}
		
		// Print key usage if present
		if cert.KeyUsage != 0 {
			fmt.Printf("   Key Usage: %v\n", formatKeyUsage(cert.KeyUsage))
		}
		
		// Print extended key usage if present
		if len(cert.ExtKeyUsage) > 0 {
			fmt.Printf("   Extended Key Usage: %v\n", formatExtKeyUsage(cert.ExtKeyUsage))
		}
	}
}

func checkTLSVersions(host, port string, timeout time.Duration) {
	versions := []uint16{
		tls.VersionSSL30,
		tls.VersionTLS10,
		tls.VersionTLS11,
		tls.VersionTLS12,
		tls.VersionTLS13,
	}

	for _, version := range versions {
		config := &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         version,
			MaxVersion:         version,
			ServerName:         host,
		}

		conn, err := tls.DialWithDialer(
			&net.Dialer{Timeout: timeout},
			"tcp",
			fmt.Sprintf("%s:%s", host, port),
			config,
		)

		if err != nil {
			fmt.Printf("   ❌ %s: Not supported\n", tlsVersions[version])
		} else {
			conn.Close()
			negotiatedVersion := conn.ConnectionState().Version
			if negotiatedVersion == version {
				if version <= tls.VersionTLS11 {
					fmt.Printf("   ⚠️ %s: Supported (DEPRECATED, SECURITY RISK)\n", tlsVersions[version])
				} else {
					fmt.Printf("   ✅ %s: Supported\n", tlsVersions[version])
				}
			} else {
				fmt.Printf("   ❓ %s: Server negotiated different version\n", tlsVersions[version])
			}
		}
	}
}

func formatKeyUsage(usage x509.KeyUsage) string {
	var usages []string
	
	if usage&x509.KeyUsageDigitalSignature != 0 {
		usages = append(usages, "DigitalSignature")
	}
	if usage&x509.KeyUsageContentCommitment != 0 {
		usages = append(usages, "ContentCommitment")
	}
	if usage&x509.KeyUsageKeyEncipherment != 0 {
		usages = append(usages, "KeyEncipherment")
	}
	if usage&x509.KeyUsageDataEncipherment != 0 {
		usages = append(usages, "DataEncipherment")
	}
	if usage&x509.KeyUsageKeyAgreement != 0 {
		usages = append(usages, "KeyAgreement")
	}
	if usage&x509.KeyUsageCertSign != 0 {
		usages = append(usages, "CertSign")
	}
	if usage&x509.KeyUsageCRLSign != 0 {
		usages = append(usages, "CRLSign")
	}
	if usage&x509.KeyUsageEncipherOnly != 0 {
		usages = append(usages, "EncipherOnly")
	}
	if usage&x509.KeyUsageDecipherOnly != 0 {
		usages = append(usages, "DecipherOnly")
	}
	
	return strings.Join(usages, ", ")
}

func formatExtKeyUsage(usage []x509.ExtKeyUsage) string {
	var usages []string
	
	for _, u := range usage {
		switch u {
		case x509.ExtKeyUsageAny:
			usages = append(usages, "Any")
		case x509.ExtKeyUsageServerAuth:
			usages = append(usages, "ServerAuth")
		case x509.ExtKeyUsageClientAuth:
			usages = append(usages, "ClientAuth")
		case x509.ExtKeyUsageCodeSigning:
			usages = append(usages, "CodeSigning")
		case x509.ExtKeyUsageEmailProtection:
			usages = append(usages, "EmailProtection")
		case x509.ExtKeyUsageIPSECEndSystem:
			usages = append(usages, "IPSECEndSystem")
		case x509.ExtKeyUsageIPSECTunnel:
			usages = append(usages, "IPSECTunnel")
		case x509.ExtKeyUsageIPSECUser:
			usages = append(usages, "IPSECUser")
		case x509.ExtKeyUsageTimeStamping:
			usages = append(usages, "TimeStamping")
		case x509.ExtKeyUsageOCSPSigning:
			usages = append(usages, "OCSPSigning")
		default:
			usages = append(usages, fmt.Sprintf("Unknown(%d)", u))
		}
	}
	
	return strings.Join(usages, ", ")
}
