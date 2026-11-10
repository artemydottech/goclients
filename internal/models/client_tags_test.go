package models

import (
	"reflect"
	"strings"
	"testing"
)

func TestNormalizeClientTags(t *testing.T) {
	input := []string{" VIP ", "Новый", "vip"}
	got, err := NormalizeClientTags(input)
	if err != nil || !reflect.DeepEqual(got, []string{"vip", "новый"}) {
		t.Fatalf("tags: %+v %v", got, err)
	}
	if input[0] != " VIP " {
		t.Fatal("input changed")
	}
	for _, bad := range [][]string{{" "}, {strings.Repeat("я", 33)}, {"a\nb"}, make([]string, 21)} {
		if _, err := NormalizeClientTags(bad); err == nil {
			t.Fatalf("accepted %+v", bad)
		}
	}
	empty, err := NormalizeClientTags([]string{})
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatalf("clear: %+v %v", empty, err)
	}
}
