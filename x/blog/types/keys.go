package types

const (
	// ModuleName defines the module name
	ModuleName = "blog"
	// StoreKey defines the primary module store key
	StoreKey = ModuleName
	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_blog"
	// PostKey is used to uniquely identify posts within the system
	PostKey = "Post/value/"
	// PostCountKey is used to keep track of the ID of the latest post
	PostCountKey = "Post/count/"
)

var (
	ParamsKey = []byte("p_blog")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
