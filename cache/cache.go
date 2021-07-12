package cache

// Copied from: https://medium.com/@melvinodsa/building-a-high-performant-concurrent-cache-in-golang-b6442c20b2ca

type RequestType uint

const (
	READ RequestType = iota
	WRITE
	DELETE
)

func WriteEntry(ch chan Request, id string, entry interface{}) {
	ch <- Request{
		Type:    WRITE,
		Payload: entry,
		Key:     id,
	}
}

type Request struct {
	Type    RequestType
	Payload interface{}
	Key     string
	Out     chan Request
}

func Initialize(ch chan Request) {
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
