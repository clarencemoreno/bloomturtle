package bloomfilter

type BloomFilter struct {
	size      uint
	bitset    []bool
	hashFuncs []func([]byte) uint
}

// New creates a new Bloom filter
func New(size uint, hashFuncs []func([]byte) uint) *BloomFilter {
	return &BloomFilter{
		size:      size,
		bitset:    make([]bool, size),
		hashFuncs: hashFuncs,
	}
}

// Add adds a new element to the Bloom filter
func (bf *BloomFilter) Add(data []byte) {
	for _, hashFunc := range bf.hashFuncs {
		index := hashFunc(data) % bf.size
		bf.bitset[index] = true
	}
}

// Contains checks if an element is in the Bloom filter
func (bf *BloomFilter) Contains(data []byte) bool {
	for _, hashFunc := range bf.hashFuncs {
		index := hashFunc(data) % bf.size
		if !bf.bitset[index] {
			return false
		}
	}
	return true
}
