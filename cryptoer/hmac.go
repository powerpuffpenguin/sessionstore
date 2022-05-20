package cryptoer

import (
	"crypto"
	"crypto/hmac"
	"encoding/base64"
	"errors"
)

// Implements the HMAC-SHA family of signing methods signing methods
// Expects key type of []byte for both signing and validation
type SigningMethodHMAC struct {
	Name string
	Hash crypto.Hash
}

// Specific instances for HS256 and company
var (
	SigningMethodHMD5   *SigningMethodHMAC
	SigningMethodHS1    *SigningMethodHMAC
	SigningMethodHS256  *SigningMethodHMAC
	SigningMethodHS384  *SigningMethodHMAC
	SigningMethodHS512  *SigningMethodHMAC
	ErrSignatureInvalid = errors.New("signature is invalid")
)

func init() {
	// HMD5
	SigningMethodHMD5 = &SigningMethodHMAC{"HMD5", crypto.MD5}
	RegisterSigningMethod(SigningMethodHMD5.Alg(), func() SigningMethod {
		return SigningMethodHMD5
	})
	// HS1
	SigningMethodHS1 = &SigningMethodHMAC{"HS1", crypto.SHA1}
	RegisterSigningMethod(SigningMethodHS1.Alg(), func() SigningMethod {
		return SigningMethodHS1
	})
	// HS256
	SigningMethodHS256 = &SigningMethodHMAC{"HS256", crypto.SHA256}
	RegisterSigningMethod(SigningMethodHS256.Alg(), func() SigningMethod {
		return SigningMethodHS256
	})

	// HS384
	SigningMethodHS384 = &SigningMethodHMAC{"HS384", crypto.SHA384}
	RegisterSigningMethod(SigningMethodHS384.Alg(), func() SigningMethod {
		return SigningMethodHS384
	})

	// HS512
	SigningMethodHS512 = &SigningMethodHMAC{"HS512", crypto.SHA512}
	RegisterSigningMethod(SigningMethodHS512.Alg(), func() SigningMethod {
		return SigningMethodHS512
	})
}

func (m *SigningMethodHMAC) Alg() string {
	return m.Name
}

// Verify the signature of HSXXX tokens.  Returns nil if the signature is valid.
func (m *SigningMethodHMAC) Verify(key, value []byte, signature string) error {
	// Decode signature, for comparison
	sig, err := DecodeSegment(signature)
	if err != nil {
		return err
	}

	// Can we use the specified hashing method?
	if !m.Hash.Available() {
		return ErrHashUnavailable
	}

	// This signing method is symmetric, so we validate the signature
	// by reproducing the signature from the signing string and key, then
	// comparing that against the provided signature.
	hasher := hmac.New(m.Hash.New, key)
	hasher.Write(value)
	if !hmac.Equal(sig, hasher.Sum(nil)) {
		return ErrSignatureInvalid
	}

	// No validation errors.  Signature is good.
	return nil
}

// Implements the Sign method from SigningMethod for this signing method.
func (m *SigningMethodHMAC) Sign(key, value []byte) (string, error) {
	if !m.Hash.Available() {
		return "", ErrHashUnavailable
	}

	hasher := hmac.New(m.Hash.New, key)
	hasher.Write(value)

	return EncodeSegment(hasher.Sum(nil)), nil
}

func EncodeSegment(seg []byte) string {
	return base64.RawURLEncoding.EncodeToString(seg)
}

func DecodeSegment(seg string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(seg)
}
