package converter

import (
	"bytes"
	"fmt"
	"os/exec"
)

// StripSignature removes CMS/PKCS#7 signatures from mobileconfig data
func StripSignature(data []byte) ([]byte, error) {
	// Check if data starts with CMS signature markers
	if !bytes.HasPrefix(data, []byte("-----BEGIN PKCS7-----")) &&
		!bytes.HasPrefix(data, []byte{0x30, 0x80}) &&  // DER indefinite length
		!bytes.HasPrefix(data, []byte{0x30, 0x82}) {   // DER definite length
		return nil, ErrAlreadyUnsigned
	}

	// Use openssl to strip the signature
	cmd := exec.Command("openssl", "smime", "-verify", "-noverify", "-inform", "DER")
	cmd.Stdin = bytes.NewReader(data)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Try PEM format
		cmd = exec.Command("openssl", "smime", "-verify", "-noverify", "-inform", "PEM")
		cmd.Stdin = bytes.NewReader(data)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrSignatureStrippingFailed, stderr.String())
		}
	}

	return stdout.Bytes(), nil
}
