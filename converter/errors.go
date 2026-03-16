package converter

import "errors"

var (
	ErrNotPlist               = errors.New("the file is not a valid property list")
	ErrNoPayloadContent       = errors.New("no PayloadContent array found in the mobileconfig")
	ErrNoConvertiblePayloads  = errors.New("none of the payloads in this mobileconfig have matching entries in the Intune Settings Catalog")
	ErrAlreadyUnsigned        = errors.New("file is not signed")
	ErrSignatureStrippingFailed = errors.New("signature removal failed")
)
