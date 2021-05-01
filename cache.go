package commitlog

// Copied from: https://medium.com/@melvinodsa/building-a-high-performant-concurrent-cache-in-golang-b6442c20b2ca

type requestType uint

const (
	READ requestType = iota
	WRITE
	DELETE
)

func writeCacheEntry(ch chan cacheRequest, id string, entry interface{}) {
	ch <- cacheRequest{
		Type:    WRITE,
		Payload: entry,
		Key:     id,
	}
}

type cacheRequest struct {
	Type    requestType
	Payload interface{}
	Key     string
	Out     chan cacheRequest
}

func initializeCache(ch chan cacheRequest) {
	store := map[string]interface{}{}

	for req := range ch {
		switch req.Type {
		case READ:
			req.Payload = store[req.Key]
			req.Out <- req
		case WRITE:
			store[req.Key] = req.Payload
		case DELETE:
			delete(store, req.Key)
		}
	}
}
