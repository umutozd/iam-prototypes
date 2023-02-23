package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type AuthLib interface {
	// Name returns the name of the underlying authentication library.
	Name() AuthLibName

	// Authorize returns whether user has access to take the given action
	// on the given resource.
	Authorize(userId string, usergroups []string, resource any, action string) bool
}

type AuthLibName string

const (
	AuthLibNameCasbin AuthLibName = "casbin"
	AuthLibNameGorbac AuthLibName = "gorbac"
	AuthLibNameOso    AuthLibName = "oso"
)

type JOSEHeader map[string]string

const (
	HeaderMediaType    = "typ"
	HeaderKeyAlgorithm = "alg"
	HeaderKeyID        = "kid"
)

type JWT struct {
	RawHeader  string
	Header     JOSEHeader
	RawPayload string
	Payload    []byte
	Signature  []byte
}

func (j *JWT) KeyID() (string, bool) {
	kID, ok := j.Header[HeaderKeyID]
	return kID, ok
}
func (j *JWT) Claims() (Claims, error) {
	return decodeClaims(j.Payload)
}
func (j *JWT) DecodeClaims(out interface{}) error {
	return json.Unmarshal(j.Payload, out)
}

// Encoded data part of the token which may be signed.
func (j *JWT) Data() string {
	return strings.Join([]string{j.RawHeader, j.RawPayload}, ".")
}

// Full encoded JWT token string in format: header.claims.signature
func (j *JWT) Encode() string {
	d := j.Data()
	s := encodeSegment(j.Signature)
	return strings.Join([]string{d, s}, ".")
}

func ParseJWT(raw string) (*JWT, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed JWT, only %d segments", len(parts))
	}

	rawSig := parts[2]
	jwt := &JWT{
		RawHeader:  parts[0],
		RawPayload: parts[1],
	}

	header, err := decodeHeader(jwt.RawHeader)
	if err != nil {
		return nil, fmt.Errorf("malformed JWT, unable to decode header, %s", err)
	}
	if err = header.validate(); err != nil {
		return nil, fmt.Errorf("malformed JWT, %s", err)
	}
	jwt.Header = header

	payload, err := decodeSegment(jwt.RawPayload)
	if err != nil {
		return nil, fmt.Errorf("malformed JWT, unable to decode payload: %s", err)
	}
	jwt.Payload = payload

	sig, err := decodeSegment(rawSig)
	if err != nil {
		return nil, fmt.Errorf("malformed JWT, unable to decode signature: %s", err)
	}
	jwt.Signature = sig

	return jwt, nil
}

func decodeHeader(seg string) (JOSEHeader, error) {
	b, err := decodeSegment(seg)
	if err != nil {
		return nil, err
	}

	var h JOSEHeader
	err = json.Unmarshal(b, &h)
	if err != nil {
		return nil, err
	}

	return h, nil
}

// Decode JWT specific base64url encoding with padding stripped
func decodeSegment(seg string) ([]byte, error) {
	if l := len(seg) % 4; l != 0 {
		seg += strings.Repeat("=", 4-l)
	}
	return base64.URLEncoding.DecodeString(seg)
}

func (j JOSEHeader) validate() error {
	if _, exists := j[HeaderKeyAlgorithm]; !exists {
		return fmt.Errorf("header missing %q parameter", HeaderKeyAlgorithm)
	}

	return nil
}

type Claims map[string]interface{}

func decodeClaims(payload []byte) (Claims, error) {
	var c Claims
	if err := json.Unmarshal(payload, &c); err != nil {
		return nil, fmt.Errorf("malformed JWT claims, unable to decode: %v", err)
	}
	return c, nil
}

// Encode JWT specific base64url encoding with padding stripped
func encodeSegment(seg []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(seg), "=")
}
