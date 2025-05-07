package federation

import (
	"reflect"
	"time"
	"unsafe"

	pub "github.com/go-ap/activitypub"
	"github.com/valyala/fastjson"
)

const (
	TrailType pub.ActivityVocabularyType = "Trail"
)

type Trail struct {
	pub.Object

	Distance      float64   `jsonld:"distance,omitempty"`
	ElevationGain float64   `jsonld:"elevation_gain,omitempty"`
	ElevationLoss float64   `jsonld:"elevation_loss,omitempty"`
	Duration      float64   `jsonld:"duration,omitempty"`
	Difficulty    string    `jsonld:"difficulty,omitempty"`
	Category      string    `jsonld:"category,omitempty"`
	Date          time.Time `jsonld:"date,omitempty"`
	Thumbnail     string    `jsonld:"thumbnail,omitempty"`
	Gpx           string    `jsonld:"gpx,omitempty"`
}

// TrailNew initializes a Trail type object
func TrailNew() *Trail {
	o := Trail{}
	o.Type = TrailType
	return &o
}

func (r Trail) MarshalJSON() ([]byte, error) {
	b, err := r.Object.MarshalJSON()
	if len(b) == 0 || err != nil {
		return nil, err
	}

	b = b[:len(b)-1]
	pub.JSONWriteFloatProp(&b, "distance", r.Distance)
	pub.JSONWriteFloatProp(&b, "elevation_gain", r.ElevationGain)
	pub.JSONWriteFloatProp(&b, "elevation_loss", r.ElevationLoss)
	pub.JSONWriteFloatProp(&b, "duration", r.Duration)
	pub.JSONWriteStringProp(&b, "difficulty", r.Difficulty)
	pub.JSONWriteStringProp(&b, "category", r.Category)
	pub.JSONWriteTimeProp(&b, "date", r.Date)
	pub.JSONWriteStringProp(&b, "thumbnail", r.Thumbnail)
	pub.JSONWriteStringProp(&b, "gpx", r.Gpx)

	pub.JSONWrite(&b, '}')
	return b, nil
}

func JSONLoadTrail(val *fastjson.Value, r *Trail) error {
	r.Distance = pub.JSONGetFloat(val, "distance")
	r.ElevationGain = pub.JSONGetFloat(val, "elevation_gain")
	r.ElevationLoss = pub.JSONGetFloat(val, "elevation_loss")
	r.Duration = pub.JSONGetFloat(val, "duration")
	r.Difficulty = pub.JSONGetString(val, "difficulty")
	r.Category = pub.JSONGetString(val, "category")
	r.Date = pub.JSONGetTime(val, "date")
	r.Thumbnail = pub.JSONGetString(val, "thumbnail")
	r.Gpx = pub.JSONGetString(val, "gpx")

	return pub.OnObject(&r.Object, func(o *pub.Object) error {
		return pub.JSONLoadObject(val, o)
	})
}

func (r *Trail) UnmarshalJSON(data []byte) error {
	p := fastjson.Parser{}
	val, err := p.ParseBytes(data)
	if err != nil {
		return err
	}
	return JSONLoadTrail(val, r)
}

// ToTrail tries to convert the it Item to a Trail Object.
func ToTrail(it pub.Item) (*Trail, error) {
	switch i := it.(type) {
	case *Trail:
		return i, nil
	case Trail:
		return &i, nil
	case *pub.Object:
		return (*Trail)(unsafe.Pointer(i)), nil
	case pub.Object:
		return (*Trail)(unsafe.Pointer(&i)), nil
	default:
		typ := reflect.TypeOf(new(Trail))
		if i, ok := reflect.ValueOf(it).Convert(typ).Interface().(*Trail); ok {
			return i, nil
		}
	}
	return nil, pub.ErrorInvalidType[pub.Object](it)
}

type withTrailFn func(*Trail) error

// OnTrail calls function fn on it Item if it can be asserted to type *Trail
func OnTrail(it pub.Item, fn withTrailFn) error {
	if it == nil {
		return nil
	}
	ob, err := ToTrail(it)
	if err != nil {
		return err
	}
	return fn(ob)
}

func (t Trail) GetType() pub.ActivityVocabularyType {
	return t.Object.Type
}

// GetID returns the ID corresponding to the current Tombstone
func (t Trail) GetID() pub.ID {
	return t.Object.ID
}
