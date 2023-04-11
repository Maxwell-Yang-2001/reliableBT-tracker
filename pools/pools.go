package pools

import (
	"github.com/crimist/trakx/bencoding"
	"github.com/crimist/trakx/tracker/storage"
)

var (
	Peerlists4        *Pool[[]byte]
	Peerlists6        *Pool[[]byte]
	BaselineProviders *Pool[[]byte]
	Dictionaries      *Pool[*bencoding.Dictionary]
	Peers             *Pool[*storage.Peer]
)

func Initialize(numwantLimit int) {
	peerlist4Max := 6 * numwantLimit  // ipv4 + port
	peerlist6Max := 18 * numwantLimit // ipv6 + port
	baselineProvidersMax := 18

	Peerlists4 = NewPool(func() any {
		data := make([]byte, peerlist4Max)
		return data
	}, func(data []byte) {
		data = data[:peerlist4Max]
		_ = data
	})

	Peerlists6 = NewPool(func() any {
		data := make([]byte, peerlist6Max)
		return data
	}, func(data []byte) {
		data = data[:peerlist6Max]
		_ = data
	})

	BaselineProviders = NewPool(func() any {
		data := make([]byte, baselineProvidersMax)
		return data
	}, func(data []byte) {
		data = data[:baselineProvidersMax]
		_ = data
	})

	Dictionaries = NewPool(func() any {
		return bencoding.NewDictionary()
	}, func(dictionary *bencoding.Dictionary) {
		dictionary.Reset()
	})

	Peers = NewPool[*storage.Peer](func() any {
		return new(storage.Peer)
	}, nil)
}
