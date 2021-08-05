package cache

// Based on: https://medium.com/@melvinodsa/building-a-high-performant-concurrent-cache-in-golang-b6442c20b2ca

type Cache chan request

type requestType uint

const (
	READ requestType = iota
	WRITE
	DELETE
)

func New() Cache {
	return From(map[string]interface{}{})
}

func From(cache map[string]interface{}) Cache {
	ch := make(chan request)
	go initialize(ch, cache)
	return ch
}

func (c Cache) Write(key string, entry interface{}) {
	c <- request{
		Type:    WRITE,
		Payload: entry,
		Key:     key,
	}
}

func (c Cache) Delete(key string) {
	c <- request{
		Type: DELETE,
		Key:  key,
	}
}

func (c Cache) Read(key string) interface{} {
	out := make(chan request)
	c <- request{
		Type:    READ,
		Key:     key,
		Out: out,
	}
	got := <-out
	return got.Payload
}

type request struct {
	Type    requestType
	Payload interface{}
	Key     string
	Out     chan request
}

func initialize(ch chan request, initial map[string]interface{}) {
	store := map[string]interface{}{}
	// Copy map so caller can't maintain reference
	// to the internal store and break thread safety
	for k, v := range initial {
		store[k] = v
	}

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
