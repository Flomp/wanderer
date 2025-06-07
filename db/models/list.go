package models

import (
	"reflect"
	"unsafe"

	pub "github.com/go-ap/activitypub"
	"github.com/valyala/fastjson"
)

const (
	ListType pub.ActivityVocabularyType = "List"
)

type List struct {
	pub.Object

	Avatar string `jsonld:"avatar,omitempty"`
}

// ListNew initializes a List type object
func ListNew() *List {
	o := List{}
	o.Type = ListType
	return &o
}

func (r List) MarshalJSON() ([]byte, error) {
	b, err := r.Object.MarshalJSON()
	if len(b) == 0 || err != nil {
		return nil, err
	}

	b = b[:len(b)-1]
	pub.JSONWriteStringProp(&b, "avatar", r.Avatar)

	pub.JSONWrite(&b, '}')
	return b, nil
}

func JSONLoadList(val *fastjson.Value, r *List) error {
	r.Avatar = pub.JSONGetString(val, "avatar")

	return pub.OnObject(&r.Object, func(o *pub.Object) error {
		return pub.JSONLoadObject(val, o)
	})
}

func (r *List) UnmarshalJSON(data []byte) error {
	p := fastjson.Parser{}
	val, err := p.ParseBytes(data)
	if err != nil {
		return err
	}
	return JSONLoadList(val, r)
}

// ToList tries to convert the it Item to a List Object.
func ToList(it pub.Item) (*List, error) {
	switch i := it.(type) {
	case *List:
		return i, nil
	case List:
		return &i, nil
	case *pub.Object:
		return (*List)(unsafe.Pointer(i)), nil
	case pub.Object:
		return (*List)(unsafe.Pointer(&i)), nil
	default:
		typ := reflect.TypeOf(new(List))
		if i, ok := reflect.ValueOf(it).Convert(typ).Interface().(*List); ok {
			return i, nil
		}
	}
	return nil, pub.ErrorInvalidType[pub.Object](it)
}

type withListFn func(*List) error

// OnList calls function fn on it Item if it can be asserted to type *List
func OnList(it pub.Item, fn withListFn) error {
	if it == nil {
		return nil
	}
	ob, err := ToList(it)
	if err != nil {
		return err
	}
	return fn(ob)
}

func (t List) GetType() pub.ActivityVocabularyType {
	return t.Object.Type
}

// GetID returns the ID corresponding to the current Tombstone
func (t List) GetID() pub.ID {
	return t.Object.ID
}
