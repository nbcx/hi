package hi

// todo: wait check
type Execer interface {
	Next()
	GetIndex() int8
	SetIndex(int8)
	Param(key string) string
	SetParams(Params)
	GetParams() Params
	AddParam(key, value string)
	SetFullPath(string) // todo: 待整合到Execer
	FullPath() string
	Copy() Execer
	WriterMem() *responseWriter
	SetWriterMem(rw responseWriter)
	Abort()
	Header(key, value string)
	AbortWithStatus(code int)
}

func NewExecer[T IContext](ctx T, handlers HandlersChain[T]) Execer {
	return &Exec[T]{ctx: ctx, handlers: handlers, index: -1}
}

type Exec[T IContext] struct {
	index     int8
	handlers  HandlersChain[T]
	ctx       T
	params    Params
	fullPath  string
	writerMem responseWriter
}

func (c *Exec[T]) Copy() Execer {
	exec := &Exec[T]{} // todo: test
	exec.fullPath = c.fullPath
	return c
}

func (c *Exec[T]) Abort() {
	c.index = abortIndex
}

func (c *Exec[T]) AbortWithStatus(code int) {
	c.writerMem.WriteHeader(code)
	c.writerMem.WriteHeaderNow()
	c.Abort()
}

func (c *Exec[T]) Header(key, value string) {
	if value == "" {
		c.writerMem.Header().Del(key)
		return
	}
	c.writerMem.Header().Set(key, value)
}

func (c *Exec[T]) WriterMem() *responseWriter {
	return &c.writerMem
}

func (c *Exec[T]) SetWriterMem(rw responseWriter) {
	c.writerMem = rw
}

func (c *Exec[T]) SetFullPath(fullPath string) {
	c.fullPath = fullPath
}

func (c *Exec[T]) GetIndex() int8 {
	return c.index
}

func (c *Exec[T]) FullPath() string {
	return c.fullPath
}

func (c *Exec[T]) SetIndex(index int8) {
	c.index = index
}

func (c *Exec[T]) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		if c.handlers[c.index] == nil {
			continue
		}
		c.handlers[c.index](c.ctx)
		c.index++
	}
}

func (c *Exec[T]) Param(key string) string {
	return c.params.ByName(key)
}

func (c *Exec[T]) SetParams(params Params) {
	c.params = params
}

func (c *Exec[T]) GetParams() Params {
	return c.params
}

// AddParam adds param to context and
// replaces path param key with given value for e2e testing purposes
// Example Route: "/user/:id"
// AddParam("id", 1)
// Result: "/user/1"
func (c *Exec[T]) AddParam(key, value string) {
	c.params = append(c.params, Param{Key: key, Value: value})
}
