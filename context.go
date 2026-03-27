package service

import "context"

type Context struct {
	context.Context
	data map[string]any
}

func NewContext(parent context.Context) *Context {
	return &Context{
		Context: parent,
		data:    make(map[string]any),
	}
}

func (c *Context) Set(key string, value any) {
	c.data[key] = value
}

func (c *Context) Value(key any) any {
	if k, ok := key.(string); ok {
		if v, exists := c.data[k]; exists {
			return v
		}
	}
	return c.Context.Value(key)
}
