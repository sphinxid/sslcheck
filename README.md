# SSLCheck

A command-line tool written in Go to check SSL/TLS configurations for HTTPS hosts.

## Features

- Validates server certificates (leaf certificates)
- Validates intermediate certificates
- Validates root CA certificates
- Checks which SSL/TLS protocols are supported (identifying deprecated protocols)
- Verifies certificate chain validity
- Displays certificate expiration information

## Installation

### Pre-built binaries

You can download pre-built binaries for Linux, macOS, and Windows from the [GitHub Releases](https://github.com/yourusername/sslcheck/releases) page.

1. Download the appropriate binary for your platform
2. Extract the zip file
3. Make the binary executable (Linux/macOS only):
   ```bash
   chmod +x sslcheck_*
   ```
4. Move the binary to a directory in your PATH (optional):
   ```bash
   # Linux/macOS
   sudo mv sslcheck_* /usr/local/bin/sslcheck
   
   # Windows
   # Move the .exe file to a location in your PATH
   ```

### Prerequisites

- Go 1.16 or higher (only needed if building from source)

### Building from source

```bash
# Clone the repository
git clone https://github.com/yourusername/sslcheck.git
cd sslcheck

# Build the binary
go build -o sslcheck

# Optional: Move to a directory in your PATH
sudo mv sslcheck /usr/local/bin/
```

## Usage

```bash
# Basic usage
./sslcheck -host example.com

# Specify a different port
./sslcheck -host example.com -port 8443

# Set a custom timeout (in seconds)
./sslcheck -host example.com -timeout 5

# Show detailed certificate information
./sslcheck -host example.com -verbose
```

### Command-line options

- `-host`: The hostname to check (required)
- `-port`: The port to connect to (default: 443)
- `-timeout`: Connection timeout in seconds (default: 10)
- `-verbose`: Show detailed certificate information

## Example Output

```
Checking SSL/TLS for example.com:443

=== TLS Protocol Support ===
   ❌ SSL 3.0 (deprecated): Not supported
   ❌ TLS 1.0 (deprecated): Not supported
   ❌ TLS 1.1 (deprecated): Not supported
   ✅ TLS 1.2: Supported
   ✅ TLS 1.3: Supported

=== Certificate Chain ===
1. Server Certificate:
   Subject: example.com
   Issuer: DigiCert SHA2 Secure Server CA
   Valid from: 2023-01-15 to 2024-01-15 (250 days left)
   ✅ Certificate date is valid
   ✅ Hostname verification PASSED

2. Intermediate Certificate 1:
   Subject: DigiCert SHA2 Secure Server CA
   Issuer: DigiCert Global Root CA
   Valid from: 2013-03-08 to 2028-03-08 (1800 days left)
   ✅ Certificate date is valid

3. Root CA Certificate:
   Subject: DigiCert Global Root CA
   Issuer: DigiCert Global Root CA
   Valid from: 2006-11-10 to 2031-11-10 (2500 days left)
   ✅ Certificate date is valid
   ℹ️ Self-signed certificate detected

=== Certificate Chain Verification ===
✅ Certificate chain verification PASSED
```

## License

MIT
