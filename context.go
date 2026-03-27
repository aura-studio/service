package service

import "context"

type contextKey string

const contextKeySelf contextKey = "Service.Context"

type Context struct {
	context.Context
	data map[string]any
}

func NewContext(parent context.Context) *Context {
	c := &Context{
		Context: parent,
		data:    make(map[string]any),
	}
	// Store self so it can be retrieved from any wrapped context
	c.data[string(contextKeySelf)] = c
	return c
}

// ContextFrom retrieves the *Context from any context in the chain.
func ContextFrom(ctx context.Context) *Context {
	if v := ctx.Value(contextKeySelf); v != nil {
		return v.(*Context)
	}
	return nil
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
	if k, ok := key.(contextKey); ok {
		if v, exists := c.data[string(k)]; exists {
			return v
		}
	}
	return c.Context.Value(key)
}
