package route_test

import (
	"fmt"
	"testing"

	"github.com/aura-studio/magic"
	"github.com/aura-studio/service/route"
	"github.com/aura-studio/style"
)

func TestRoute(t *testing.T) {
	style := style.ChainStyle{
		ChainSeperator: magic.SeparatorPeriod,
		WordSeparator:  magic.SeparatorUnderscore,
	}
	src := style.Chain("bus.test_route.src")
	dst := style.Chain("bus.test_route.dst")
	route := route.NewChainRoute(src, dst)
	t.Logf("%v", route)
	if fmt.Sprint(route) != "[Bus:TestRoute:Src] -> [<Bus>:TestRoute:Dst]" {
		t.Fatal("route string is not as expected")
	}
}

func Test_GoogleChain(t *testing.T) {
	src := style.GoogleChain("bus/QueryAll/src")
	fmt.Println(src)
}
