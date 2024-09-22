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
}

func NewExecer[T IContext](ctx T, handlers HandlersChain[T]) Execer {
	return &Exec[T]{ctx: ctx, handlers: handlers, index: -1}
}

type Exec[T IContext] struct {
	index    int8
	handlers HandlersChain[T]
	ctx      T
	Params   Params
}

func (c *Exec[T]) GetIndex() int8 {
	return c.index
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
	return c.Params.ByName(key)
}

func (c *Exec[T]) SetParams(params Params) {
	c.Params = params
}

func (c *Exec[T]) GetParams() Params {
	return c.Params
}

// AddParam adds param to context and
// replaces path param key with given value for e2e testing purposes
// Example Route: "/user/:id"
// AddParam("id", 1)
// Result: "/user/1"
func (c *Exec[T]) AddParam(key, value string) {
	c.Params = append(c.Params, Param{Key: key, Value: value})
}
