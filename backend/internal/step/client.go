package step

import (
	"archive/zip"
	"bytes"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type CertBundle struct {
	CertPEM     []byte
	KeyPEM      []byte
	ChainPEM    []byte
	FullChainPEM []byte
	PFXData     []byte
	Serial      string
	NotAfter    time.Time
}

type StepClient struct {
	CAURL             string
	ProvisionerName   string
	ProvisionerPassword string
}

func NewStepClient(caURL, provisionerName, provisionerPassword string) *StepClient {
	return &StepClient{
		CAURL:             caURL,
		ProvisionerName:   provisionerName,
		ProvisionerPassword: provisionerPassword,
	}
}

func (s *StepClient) IssueCertificate(cn string, sans []string, notAfterDays int) (*CertBundle, error) {
	// Create temporary directory for certificate files
	tempDir, err := os.MkdirTemp("", "step-cert-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	certPath := filepath.Join(tempDir, "cert.crt")
	keyPath := filepath.Join(tempDir, "cert.key")

	// Build step command
	args := []string{
		"ca", "certificate",
		cn,
		certPath,
		keyPath,
		"--ca-url", s.CAURL,
		"--provisioner", s.ProvisionerName,
		"--password", s.ProvisionerPassword,
		"--not-after", fmt.Sprintf("%dd", notAfterDays),
	}

	// Add SANs if provided
	for _, san := range sans {
		args = append(args, "--san", san)
	}

	// Execute step command
	cmd := exec.Command("step", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("step command failed: %s, error: %w", string(output), err)
	}

	// Read certificate and key files
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cert file: %w", err)
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	// Get certificate chain
	chainPEM, err := s.getChain(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain: %w", err)
	}

	// Create full chain
	fullChainPEM := append(certPEM, chainPEM...)

	// Extract serial number and expiry from certificate
	serial, notAfter, err := s.parseCertificate(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return &CertBundle{
		CertPEM:      certPEM,
		KeyPEM:       keyPEM,
		ChainPEM:     chainPEM,
		FullChainPEM: fullChainPEM,
		Serial:       serial,
		NotAfter:     notAfter,
	}, nil
}

func (s *StepClient) SignCSR(csrPEM string, notAfterDays int) (*CertBundle, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "step-csr-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	csrPath := filepath.Join(tempDir, "csr.pem")
	certPath := filepath.Join(tempDir, "cert.crt")

	// Write CSR to file
	if err := os.WriteFile(csrPath, []byte(csrPEM), 0644); err != nil {
		return nil, fmt.Errorf("failed to write CSR file: %w", err)
	}

	// Build step command
	args := []string{
		"ca", "sign",
		csrPath,
		certPath,
		"--ca-url", s.CAURL,
		"--provisioner", s.ProvisionerName,
		"--password", s.ProvisionerPassword,
		"--not-after", fmt.Sprintf("%dd", notAfterDays),
	}

	// Execute step command
	cmd := exec.Command("step", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("step command failed: %s, error: %w", string(output), err)
	}

	// Read certificate file
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cert file: %w", err)
	}

	// Get certificate chain
	chainPEM, err := s.getChain(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain: %w", err)
	}

	// Create full chain
	fullChainPEM := append(certPEM, chainPEM...)

	// Extract serial number and expiry from certificate
	serial, notAfter, err := s.parseCertificate(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return &CertBundle{
		CertPEM:      certPEM,
		ChainPEM:     chainPEM,
		FullChainPEM: fullChainPEM,
		Serial:       serial,
		NotAfter:     notAfter,
	}, nil
}

func (s *StepClient) RevokeCertificate(serial string) error {
	args := []string{
		"ca", "revoke",
		serial,
		"--ca-url", s.CAURL,
		"--provisioner", s.ProvisionerName,
		"--password", s.ProvisionerPassword,
	}

	cmd := exec.Command("step", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("step revoke failed: %s, error: %w", string(output), err)
	}

	return nil
}

func (s *StepClient) CreatePFX(certPEM, keyPEM, password string) ([]byte, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "step-pfx-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	certPath := filepath.Join(tempDir, "cert.crt")
	keyPath := filepath.Join(tempDir, "cert.key")
	pfxPath := filepath.Join(tempDir, "cert.p12")
	passwordPath := filepath.Join(tempDir, "password.txt")

	// Write certificate and key files
	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		return nil, fmt.Errorf("failed to write cert file: %w", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0644); err != nil {
		return nil, fmt.Errorf("failed to write key file: %w", err)
	}
	if err := os.WriteFile(passwordPath, []byte(password), 0644); err != nil {
		return nil, fmt.Errorf("failed to write password file: %w", err)
	}

	// Build step command
	args := []string{
		"certificate", "p12",
		pfxPath,
		certPath,
		keyPath,
		"--password-file", passwordPath,
	}

	// Execute step command
	cmd := exec.Command("step", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("step p12 command failed: %s, error: %w", string(output), err)
	}

	// Read PFX file
	pfxData, err := os.ReadFile(pfxPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PFX file: %w", err)
	}

	return pfxData, nil
}

func (s *StepClient) getChain(certPath string) ([]byte, error) {
	args := []string{
		"certificate", "chain",
		certPath,
		"--ca-url", s.CAURL,
	}

	cmd := exec.Command("step", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("step chain command failed: %s, error: %w", string(output), err)
	}

	return output, nil
}

func (s *StepClient) parseCertificate(certPEM []byte) (string, time.Time, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return "", time.Time{}, fmt.Errorf("failed to decode PEM block")
	}

	// Use openssl to get certificate details
	cmd := exec.Command("openssl", "x509", "-noout", "-serial", "-enddate")
	cmd.Stdin = bytes.NewReader(certPEM)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("openssl command failed: %s, error: %w", string(output), err)
	}

	lines := strings.Split(string(output), "\n")
	var serial string
	var notAfter time.Time

	for _, line := range lines {
		if strings.HasPrefix(line, "serial=") {
			serial = strings.TrimPrefix(line, "serial=")
		} else if strings.HasPrefix(line, "notAfter=") {
			dateStr := strings.TrimPrefix(line, "notAfter=")
			notAfter, err = time.Parse("Jan 2 15:04:05 2006 MST", dateStr)
			if err != nil {
				return "", time.Time{}, fmt.Errorf("failed to parse date: %w", err)
			}
		}
	}

	return serial, notAfter, nil
}

func (s *StepClient) CreateDownloadBundle(bundle *CertBundle, format string, pfxPassword string) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Add certificate files
	if err := s.addFileToZip(zipWriter, "cert.pem", bundle.CertPEM); err != nil {
		return nil, err
	}
	if err := s.addFileToZip(zipWriter, "chain.pem", bundle.ChainPEM); err != nil {
		return nil, err
	}
	if err := s.addFileToZip(zipWriter, "fullchain.pem", bundle.FullChainPEM); err != nil {
		return nil, err
	}

	// Add private key if available
	if len(bundle.KeyPEM) > 0 {
		if err := s.addFileToZip(zipWriter, "privkey.pem", bundle.KeyPEM); err != nil {
			return nil, err
		}
	}

	// Add PFX if requested
	if format == "pfx" && len(bundle.KeyPEM) > 0 {
		pfxData, err := s.CreatePFX(bundle.CertPEM, bundle.KeyPEM, pfxPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to create PFX: %w", err)
		}
		if err := s.addFileToZip(zipWriter, "cert.p12", pfxData); err != nil {
			return nil, err
		}
	}

	// Add README with installation instructions
	readme := s.generateReadme(format, pfxPassword)
	if err := s.addFileToZip(zipWriter, "README.txt", []byte(readme)); err != nil {
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *StepClient) addFileToZip(zipWriter *zip.Writer, filename string, data []byte) error {
	fileWriter, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}
	_, err = fileWriter.Write(data)
	return err
}

func (s *StepClient) generateReadme(format, pfxPassword string) string {
	readme := "# Certificate Installation Instructions\n\n"
	readme += "## Files in this bundle:\n"
	readme += "- cert.pem: Your certificate\n"
	readme += "- chain.pem: Certificate chain (intermediate CAs)\n"
	readme += "- fullchain.pem: Certificate + chain (use this for most applications)\n"
	readme += "- privkey.pem: Private key (keep this secure!)\n"

	if format == "pfx" {
		readme += "- cert.p12: PFX/PKCS#12 bundle (password: " + pfxPassword + ")\n"
	}

	readme += "\n## Installation Instructions\n\n"
	readme += "### Linux/Ubuntu (Nginx, Apache, etc.)\n"
	readme += "```bash\n"
	readme += "# Copy files to appropriate locations\n"
	readme += "sudo cp fullchain.pem /etc/ssl/certs/your-domain.crt\n"
	readme += "sudo cp privkey.pem /etc/ssl/private/your-domain.key\n\n"
	readme += "# For Nginx, update your server block:\n"
	readme += "# ssl_certificate /etc/ssl/certs/your-domain.crt;\n"
	readme += "# ssl_certificate_key /etc/ssl/private/your-domain.key;\n\n"
	readme += "# Reload nginx\n"
	readme += "sudo nginx -s reload\n"
	readme += "```\n\n"

	readme += "### Windows (IIS)\n"
	readme += "1. Import cert.p12 into Certificate Store\n"
	readme += "2. Use IIS Manager to bind the certificate to your site\n\n"

	readme += "### Trust the CA Root\n"
	readme += "To trust this CA on client systems:\n\n"
	readme += "**Linux/Ubuntu:**\n"
	readme += "```bash\n"
	readme += "sudo cp chain.pem /usr/local/share/ca-certificates/my-ca.crt\n"
	readme += "sudo update-ca-certificates\n"
	readme += "```\n\n"

	readme += "**Windows PowerShell:**\n"
	readme += "```powershell\n"
	readme += "Import-Certificate -FilePath chain.pem -CertStoreLocation Cert:\\LocalMachine\\Root\n"
	readme += "```\n\n"

	readme += "## Verification\n"
	readme += "```bash\n"
	readme += "# Verify certificate chain\n"
	readme += "openssl verify -CAfile chain.pem cert.pem\n\n"
	readme += "# Check certificate details\n"
	readme += "openssl x509 -in cert.pem -text -noout\n\n"
	readme += "# Test SSL connection\n"
	readme += "openssl s_client -connect your-domain:443 -showcerts\n"
	readme += "```\n"

	return readme
}
