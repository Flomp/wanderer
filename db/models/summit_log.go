package models

import (
	"reflect"
	"time"
	"unsafe"

	pub "github.com/go-ap/activitypub"
	"github.com/valyala/fastjson"
)

const (
	SummitLogType pub.ActivityVocabularyType = "SummitLog"
)

type SummitLog struct {
	pub.Object

	SummitLogId   string    `jsonld:"summitlog_id,omitempty"`
	TrailId       string    `jsonld:"trail_id,omitempty"`
	Distance      float64   `jsonld:"distance,omitempty"`
	ElevationGain float64   `jsonld:"elevation_gain,omitempty"`
	ElevationLoss float64   `jsonld:"elevation_loss,omitempty"`
	Duration      float64   `jsonld:"duration,omitempty"`
	Date          time.Time `jsonld:"date,omitempty"`
	Thumbnail     string    `jsonld:"thumbnail,omitempty"`
	Gpx           string    `jsonld:"gpx,omitempty"`
}

// SummitLogNew initializes a SummitLog type object
func SummitLogNew() *SummitLog {
	o := SummitLog{}
	o.Type = SummitLogType
	return &o
}

func (r SummitLog) MarshalJSON() ([]byte, error) {
	b, err := r.Object.MarshalJSON()
	if len(b) == 0 || err != nil {
		return nil, err
	}

	b = b[:len(b)-1]
	pub.JSONWriteStringProp(&b, "summitlog_id", r.SummitLogId)
	pub.JSONWriteStringProp(&b, "trail_id", r.TrailId)
	pub.JSONWriteFloatProp(&b, "distance", r.Distance)
	pub.JSONWriteFloatProp(&b, "elevation_gain", r.ElevationGain)
	pub.JSONWriteFloatProp(&b, "elevation_loss", r.ElevationLoss)
	pub.JSONWriteFloatProp(&b, "duration", r.Duration)
	pub.JSONWriteTimeProp(&b, "date", r.Date)
	pub.JSONWriteStringProp(&b, "thumbnail", r.Thumbnail)
	pub.JSONWriteStringProp(&b, "gpx", r.Gpx)

	pub.JSONWrite(&b, '}')
	return b, nil
}

func JSONLoadSummitLog(val *fastjson.Value, r *SummitLog) error {
	r.SummitLogId = pub.JSONGetString(val, "summitlog_id")
	r.TrailId = pub.JSONGetString(val, "trail_id")
	r.Distance = pub.JSONGetFloat(val, "distance")
	r.ElevationGain = pub.JSONGetFloat(val, "elevation_gain")
	r.ElevationLoss = pub.JSONGetFloat(val, "elevation_loss")
	r.Duration = pub.JSONGetFloat(val, "duration")
	r.Date = pub.JSONGetTime(val, "date")
	r.Thumbnail = pub.JSONGetString(val, "thumbnail")
	r.Gpx = pub.JSONGetString(val, "gpx")

	return pub.OnObject(&r.Object, func(o *pub.Object) error {
		return pub.JSONLoadObject(val, o)
	})
}

func (r *SummitLog) UnmarshalJSON(data []byte) error {
	p := fastjson.Parser{}
	val, err := p.ParseBytes(data)
	if err != nil {
		return err
	}
	return JSONLoadSummitLog(val, r)
}

// ToSummitLog tries to convert the it Item to a SummitLog Object.
func ToSummitLog(it pub.Item) (*SummitLog, error) {
	switch i := it.(type) {
	case *SummitLog:
		return i, nil
	case SummitLog:
		return &i, nil
	case *pub.Object:
		return (*SummitLog)(unsafe.Pointer(i)), nil
	case pub.Object:
		return (*SummitLog)(unsafe.Pointer(&i)), nil
	default:
		typ := reflect.TypeOf(new(SummitLog))
		if i, ok := reflect.ValueOf(it).Convert(typ).Interface().(*SummitLog); ok {
			return i, nil
		}
	}
	return nil, pub.ErrorInvalidType[pub.Object](it)
}

type withSummitLogFn func(*SummitLog) error

// OnSummitLog calls function fn on it Item if it can be asserted to type *SummitLog
func OnSummitLog(it pub.Item, fn withSummitLogFn) error {
	if it == nil {
		return nil
	}
	ob, err := ToSummitLog(it)
	if err != nil {
		return err
	}
	return fn(ob)
}

func (t SummitLog) GetType() pub.ActivityVocabularyType {
	return t.Object.Type
}

// GetID returns the ID corresponding to the current Tombstone
func (t SummitLog) GetID() pub.ID {
	return t.Object.ID
}
