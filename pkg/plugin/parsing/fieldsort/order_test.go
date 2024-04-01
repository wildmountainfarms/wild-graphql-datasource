package fieldsort

import (
	"strings"
	"testing"
)

func TestOrderRealistic(t *testing.T) {
	order := New()
	order.State([]string{"batteryVoltage", "dateMillis", "meta.name", "meta.displayName"})
	order.State([]string{"batteryVoltage", "dateMillis", "meta"})
	//println(strings.Join(order.GetOrder(), ", "))
}
func TestOrderSimple(t *testing.T) {
	order := New()
	order.State([]string{"a", "b", "c", "f"})
	order.State([]string{"b", "d", "e"})
	order.State([]string{"c", "e", "f"})
	//order.State([]string{"c", "d"})
	result := order.GetOrder()
	//println(strings.Join(result, ", "))
	if result[0] != "a" || result[1] != "b" || result[4] != "e" || result[5] != "f" {
		// Note that the order of c and d here is undetermined, which is OK
		t.Errorf("Incorrect result: %s", strings.Join(result, ", "))
	}
}
