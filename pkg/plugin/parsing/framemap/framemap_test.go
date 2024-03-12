package framemap

import (
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"reflect"
	"testing"
	"time"
)

func TestEqualityAndHash(t *testing.T) {
	labelsA := data.Labels{
		"asdf": "a",
	}
	labelsB := data.Labels{
		"asdf": "b",
	}
	if labelsEqual(labelsA, labelsB) {
		t.Error("labelsA should not equal labelsB")
	}
	if labelsHash(labelsA) == labelsHash(labelsB) {
		t.Error("Should not have the same hashes!")
	}
}

func TestFrameMap(t *testing.T) {
	fm := New()
	labelsA := data.Labels{
		"asdf": "a",
	}
	labelsB := data.Labels{
		"asdf": "b",
	}
	dataA := map[string]interface{}{
		"batteryVoltage": []float64{22.4, 22.5},
		"dateMillis":     []time.Time{time.UnixMilli(1705974650887), time.UnixMilli(1705974659884)},
	}
	dataB := map[string]interface{}{
		"batteryVoltage": []float64{22.43, 22.51},
		"dateMillis":     []time.Time{time.UnixMilli(1705974650888), time.UnixMilli(1705974659885)},
	}

	_, exists := fm.Get(labelsA)
	if exists {
		t.Error("We haven't put anything in the map! Nothing should exist!")
	}

	fm.Put(labelsA, dataA)
	expectingDataA, exists := fm.Get(labelsA)
	if !exists {
		t.Error("We expect that fm.get(labelsA) exists now!")
	}
	if !reflect.DeepEqual(dataA, expectingDataA) {
		t.Error("We expect this to be dataA!")
	}

	fm.Put(labelsB, dataB)
	expectingDataB, exists := fm.Get(labelsB)
	if !exists {
		t.Error("We expect that fm.get(labelsB) exists now!")
	}
	if !reflect.DeepEqual(dataB, expectingDataB) {
		t.Error("We expect this to be dataB!")
	}

	if len(fm.ToFrames()) != 2 {
		t.Error("ToFrames() should result in an array of size 2!")
	}
}
