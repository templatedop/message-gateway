package response

type Entity[T any] struct {
	data    T
	status  int
	headers map[string]string
}

func NewEntity[T any](status int, data T, headers map[string]string) Entity[T] {
	return Entity[T]{
		status:  status,
		data:    data,
		headers: headers,
	}
}

func (e Entity[T]) Status() int {
	if e.status == 0 {
		return 200
	}
	return e.status
}

func (e Entity[T]) Headers() map[string]string {
	return e.headers
}

func (e Entity[T]) ToResponse() any {
	return Success(e.data)
}

// GetContentType implements Stature interface
func (e Entity[T]) GetContentType() string {
	if ct, ok := e.headers["Content-Type"]; ok {
		return ct
	}
	return "application/json"
}

// GetContentDisposition implements Stature interface
func (e Entity[T]) GetContentDisposition() string {
	if cd, ok := e.headers["Content-Disposition"]; ok {
		return cd
	}
	return ""
}

// ResponseType implements Stature interface
func (e Entity[T]) ResponseType() string {
	return "json"
}

// Object implements Stature interface
func (e Entity[T]) Object() []byte {
	return nil // Not used for JSON responses
}
