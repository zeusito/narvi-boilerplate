package toolbox

import "uuid"

type TypeIdPrefix string

const (
	TypeIdPrefixIdentity     TypeIdPrefix = "id"
	TypeIdPrefixOrganization TypeIdPrefix = "org"
)

// GenerateTypeId generates a new type ID with the specified prefix
// The generated ID is a UUID v7 encoded in Crockford Base32, prefixed with the provided TypeIdPrefix
func GenerateTypeId(prefix TypeIdPrefix) string {
	id := uuid.NewV7()
	encoded := encodeCrockfordBase32(id[:])
	return string(prefix) + "_" + encoded
}
