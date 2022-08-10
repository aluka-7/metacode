package metacode

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPairsMD(t *testing.T) {
	for _, test := range []struct {
		// input
		kv []interface{}
		// output
		md Metadata
	}{
		{[]interface{}{}, Metadata{}},
		{[]interface{}{"k1", "v1", "k1", "v2"}, Metadata{"k1": "v2"}},
	} {
		md := Pairs(test.kv...)
		if !reflect.DeepEqual(md, test.md) {
			t.Fatalf("Pairs(%v) = %v, want %v", test.kv, md, test.md)
		}
	}
}
func TestCopy(t *testing.T) {
	const key, val = "key", "val"
	orig := Pairs(key, val)
	copy := orig.Copy()
	if !reflect.DeepEqual(orig, copy) {
		t.Errorf("copied value not equal to the original, got %v, want %v", copy, orig)
	}
	orig[key] = "foo"
	if v := copy[key]; v != val {
		t.Errorf("change in original should not affect copy, got %q, want %q", v, val)
	}
}
func TestJoin(t *testing.T) {
	for _, test := range []struct {
		mds  []Metadata
		want Metadata
	}{
		{[]Metadata{}, Metadata{}},
		{[]Metadata{Pairs("foo", "bar")}, Pairs("foo", "bar")},
		{[]Metadata{Pairs("foo", "bar"), Pairs("foo", "baz")}, Pairs("foo", "bar", "foo", "baz")},
		{[]Metadata{Pairs("foo", "bar"), Pairs("foo", "baz"), Pairs("zip", "zap")}, Pairs("foo", "bar", "foo", "baz", "zip", "zap")},
	} {
		md := Join(test.mds...)
		if !reflect.DeepEqual(md, test.want) {
			t.Errorf("context's metadata is %v, want %v", md, test.want)
		}
	}
}

func TestWithContext(t *testing.T) {
	md := Metadata(map[string]interface{}{RemoteIP: "127.0.0.1", Mirror: true})
	c := NewContext(context.Background(), md)
	ctx := WithContext(c)
	md1, ok := FromContext(ctx)
	if !ok {
		t.Errorf("expect ok be true")
		t.FailNow()
	}
	if !reflect.DeepEqual(md1, md) {
		t.Errorf("expect md1 equal to md")
		t.FailNow()
	}
}

func TestBool(t *testing.T) {
	md := Metadata{RemoteIP: "127.0.0.1"}
	mdContext := NewContext(context.Background(), md)
	assert.Equal(t, false, Bool(mdContext, Mirror))

	mdContext = NewContext(context.Background(), Metadata{Mirror: true})
	assert.Equal(t, true, Bool(mdContext, Mirror))

	mdContext = NewContext(context.Background(), Metadata{Mirror: "true"})
	assert.Equal(t, true, Bool(mdContext, Mirror))

	mdContext = NewContext(context.Background(), Metadata{Mirror: "1"})
	assert.Equal(t, true, Bool(mdContext, Mirror))

	mdContext = NewContext(context.Background(), Metadata{Mirror: "0"})
	assert.Equal(t, false, Bool(mdContext, Mirror))

}
func TestInt64(t *testing.T) {
	mdContext := NewContext(context.Background(), Metadata{Mid: int64(1)})
	assert.Equal(t, int64(1), Int64(mdContext, Mid))
	mdContext = NewContext(context.Background(), Metadata{Mid: int64(2)})
	assert.NotEqual(t, int64(1), Int64(mdContext, Mid))
	mdContext = NewContext(context.Background(), Metadata{Mid: 10})
	assert.NotEqual(t, int64(10), Int64(mdContext, Mid))
}

func TestRange(t *testing.T) {
	for _, test := range []struct {
		filterFunc func(key string) bool
		md         Metadata
		want       Metadata
	}{
		{
			nil,
			Pairs("foo", "bar"),
			Pairs("foo", "bar"),
		},
		{
			IsOutgoingKey,
			Pairs("foo", "bar", RemoteIP, "127.0.0.1", Mirror, "false"),
			Pairs(RemoteIP, "127.0.0.1", Mirror, "false"),
		},
		{
			IsOutgoingKey,
			Pairs("foo", "bar", Caller, "app-feed", RemoteIP, "127.0.0.1", Mirror, "true"),
			Pairs(RemoteIP, "127.0.0.1", Mirror, "true"),
		},
		{
			IsIncomingKey,
			Pairs("foo", "bar", Caller, "app-feed", RemoteIP, "127.0.0.1", Mirror, "true"),
			Pairs(Caller, "app-feed", RemoteIP, "127.0.0.1", Mirror, "true"),
		},
	} {
		var mds []Metadata
		c := NewContext(context.Background(), test.md)
		ctx := WithContext(c)
		Range(ctx,
			func(key string, value interface{}) {
				mds = append(mds, Pairs(key, value))
			},
			test.filterFunc)
		rmd := Join(mds...)
		if !reflect.DeepEqual(rmd, test.want) {
			t.Fatalf("Range(%v) = %v, want %v", test.md, rmd, test.want)
		}
		if test.filterFunc == nil {
			var mds []Metadata
			Range(ctx,
				func(key string, value interface{}) {
					mds = append(mds, Pairs(key, value))
				})
			rmd := Join(mds...)
			if !reflect.DeepEqual(rmd, test.want) {
				t.Fatalf("Range(%v) = %v, want %v", test.md, rmd, test.want)
			}
		}
	}
}
