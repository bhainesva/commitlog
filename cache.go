package commitlog

type RequestType uint

const (
	READ RequestType = iota
	WRITE
	DELETE
)

type Request struct {
	Type    RequestType
	Payload jobCacheEntry
	Key     string
	Out     chan Request
}

func Cache(ch chan Request) {
	store := map[string]jobCacheEntry{}

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
