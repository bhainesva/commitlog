package commitlog

type RequestType uint

const (
	READ RequestType = iota
	WRITE
	DELETE
)

type CacheRequest struct {
	Type    RequestType
	Payload interface{}
	Key     string
	Out     chan CacheRequest
}

func NewCache(ch chan CacheRequest) {
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
