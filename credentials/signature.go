package credentials

// SignatureType is type of Authorization requested for a given HTTP request.
type SignatureType int

// Different types of supported signatures - default is SignatureV4 or SignatureDefault.
const (
	// SignatureDefault is always set to v1.
	SignatureDefault   SignatureType = iota
	SignatureAnonymous               // Anonymous signature signifies, no signature.
)

// IsAnonymous - is signature empty?
func (s SignatureType) IsAnonymous() bool {
	return s == SignatureAnonymous
}

// Stringer humanized version of signature type,
// strings returned here are case insensitive.
func (s SignatureType) String() string {
	return "Anonymous"
}

func parseSignatureType(str string) SignatureType {
	return SignatureAnonymous
}
