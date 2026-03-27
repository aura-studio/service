package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/aura-studio/cast"
	"github.com/aura-studio/dynamic"
	"github.com/aura-studio/encodingx"
	"github.com/aura-studio/magic"
	"github.com/aura-studio/mesh/device"
	"github.com/aura-studio/mesh/message"
	"github.com/aura-studio/mesh/route"
	"github.com/aura-studio/safe"
	"github.com/aura-studio/style"
)

var _ dynamic.Tunnel = (*Service)(nil)

type envelope struct {
	Meta map[string]any `json:"meta"`
	Data string         `json:"data"`
}

type Service struct {
	target any
	bus    *device.Router
	client *device.Client
	router *device.Router
	init   func()
	close  func()
}

func New(target any) *Service {
	t := reflect.TypeOf(target)
	if !(t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct) {
		log.Panic("service must be a pointer to struct")
	}

	bus := device.NewBus()
	client := device.NewClient(magic.Client)
	bus.Integrate(client)
	router := device.NewRouter(magic.Server).Integrate(target)
	bus.Integrate(router)

	s := &Service{
		target: target,
		bus:    bus,
		client: client,
		router: router,
	}

	if init, ok := t.MethodByName("Init"); ok {
		if init.Type.NumIn() == 1 && init.Type.NumOut() == 0 {
			s.init = func() {
				init.Func.Call([]reflect.Value{reflect.ValueOf(target)})
			}
		}
	}
	if s.init == nil {
		s.init = func() {}
	}

	if close, ok := t.MethodByName("Close"); ok {
		if close.Type.NumIn() == 1 && close.Type.NumOut() == 0 {
			s.close = func() {
				close.Func.Call([]reflect.Value{reflect.ValueOf(target)})
			}
		}
	}
	if s.close == nil {
		s.close = func() {}
	}

	return s
}

func (s *Service) Meta() string {
	return ""
}

func (s *Service) Init() {
	s.init()
}

func (s *Service) Invoke(routePath string, req string) (rsp string) {
	ctx := NewContext(context.Background())

	// Decode request envelope
	var reqEnv envelope
	if err := json.Unmarshal([]byte(req), &reqEnv); err != nil {
		b, _ := json.Marshal(envelope{Meta: map[string]any{"Error": err.Error()}})
		return string(b)
	}

	// Decode base64 request data
	reqData, err := base64.StdEncoding.DecodeString(reqEnv.Data)
	if err != nil {
		b, _ := json.Marshal(envelope{Meta: map[string]any{"Error": err.Error()}})
		return string(b)
	}

	// Set request meta into context as "Request.{Key}"
	for k, v := range reqEnv.Meta {
		ctx.Set("Request."+k, cast.ToString(v))
	}

	// Route via mesh
	strs := strings.Split(routePath, "/")
	if len(strs) < 2 {
		b, _ := json.Marshal(envelope{Meta: map[string]any{"Error": fmt.Errorf("invalid route path: %s", routePath).Error()}})
		return string(b)
	}

	if err := safe.DoWithContext(ctx, func(ctx context.Context) error {
		return s.client.Invoke(ctx, &message.Message{
			Route: route.NewChainRoute(device.Addr(s.client),
				append([]string{"", magic.Server, style.Standardize(strs[1], magic.SeparatorHyphen)}, strs[2:]...)),
			Encoding: encodingx.NewJSON(),
			Data:     reqData,
		}, device.NewFuncProcessor(func(ctx context.Context, msg *message.Message) error {
			// Build response envelope from "Response.*" context keys
			rspMeta := map[string]any{}
			if c, ok := ctx.(*Context); ok {
				for k, v := range c.data {
					if after, found := strings.CutPrefix(k, "Response."); found {
						rspMeta[after] = v
					}
				}
			}
			rspEnv := envelope{
				Meta: rspMeta,
				Data: base64.StdEncoding.EncodeToString(msg.Data),
			}
			rspBytes, _ := json.Marshal(rspEnv)
			rsp = string(rspBytes)
			return nil
		}))
	}); err != nil {
		b, _ := json.Marshal(envelope{Meta: map[string]any{"Error": err.Error()}})
		return string(b)
	}

	return
}

func (s *Service) Close() {
	s.close()
}
