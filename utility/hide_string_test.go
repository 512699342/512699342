package utility

import (
	"testing"
)

func TestHideName(t *testing.T) {
	nameMap := map[string]string{"张三": "张*", "王麻子": "王**", "四个名字": "四***"}
	for origName, hideName := range nameMap {
		if HideName(origName) != hideName {
			t.Errorf("HideName fail,name:%s,hideName:%s", origName, hideName)
		}
	}
}

func TestHidePhone(t *testing.T) {
	phone := "12345678900"
	hidePhone := HidePhone(phone)
	if hidePhone != "123****8900" {
		t.Errorf("HidePhone fail,phone:%s,hidePhone:%s", phone, hidePhone)
	}
}
