package consts

const (
	// CacheKeyPrefixProduct is the Redis key namespace for product entries.
	// Format: "product:<id>"
	// Add one prefix per entity here to keep all cache namespaces visible in one place.
	CacheKeyPrefixProduct = "product"

	// ErrMsgInvalidID is returned by HTTP handlers when a path parameter cannot be
	// parsed as a valid integer ID.
	ErrMsgInvalidID = "invalid id"
)
