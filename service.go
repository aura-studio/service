package service

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/aura-studio/encodingx"
	"github.com/aura-studio/magic"
	"github.com/aura-studio/safe"
	"github.com/aura-studio/service/device"
	"github.com/aura-studio/service/message"
	"github.com/aura-studio/service/route"
	"github.com/aura-studio/style"
)

type Service struct {
	Options
	target any
	bus    *device.Router
	client *device.Client
	router *device.Router

	init  func()
	close func()
}

func New(target any, opts ...Option) *Service {
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
		Options: defaultOptions,
		target:  target,
		client:  client,
		router:  router,
	}

	for _, opt := range opts {
		opt(s)
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
	strs := strings.Split(routePath, "/")
	if len(strs) < 2 {
		return fmt.Errorf("error://invalid route path: %s", routePath).Error()
	}

	if err := safe.DoWithTimeout(60*time.Second, func(ctx context.Context) error {
		return s.client.Invoke(ctx, &message.Message{
			Route: route.NewChainRoute(device.Addr(s.client),
				append([]string{"", magic.Server, style.Standardize(strs[1], magic.SeparatorHyphen)}, strs[2:]...)),
			Encoding: encodingx.NewJSON(),
			Data:     []byte(req),
		}, device.NewFuncProcessor(func(ctx context.Context, msg *message.Message) error {
			rsp = string(msg.Data)
			return nil
		}))
	}); err != nil {
		return fmt.Errorf("error://%w", err).Error()
	}
	return
}

func (s *Service) Close() {
	s.close()
}

func (s *Service) Bus() *device.Router {
	return s.bus
}

func (s *Service) Client() *device.Client {
	return s.client
}

func (s *Service) Router() *device.Router {
	return s.router
}
