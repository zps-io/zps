package action

import (
	"fmt"
	"strings"
)

type Tag struct {
	Name  string `json:"name" hcl:"name,label"`
	Value string `json:"value" hcl:"value"`
}

func NewTag() *Tag {
	return &Tag{}
}

func (t *Tag) Key() string {
	return t.Name
}

func (t *Tag) Type() string {
	return "Tag"
}

func (t *Tag) Columns() string {
	return strings.Join([]string{
		strings.ToUpper(t.Type()),
		t.Name,
		t.Value,
	}, "|")
}

func (t *Tag) Id() string {
	return fmt.Sprint(t.Type(), ".", t.Key())
}

func (t *Tag) Condition() *bool {
	return nil
}

func (t *Tag) MayFail() bool {
	return false
}

func (t *Tag) IsValid() bool {
	return true
}
