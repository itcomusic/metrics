package metrics

import (
	"math"
	"testing"
)

func TestValidateMetricSuccess(t *testing.T) {
	f := func(s string) {
		t.Helper()
		if err := validateMetric(s); err != nil {
			t.Fatalf("cannot validate %q: %s", s, err)
		}
	}
	f("a")
	f("_9:8")
	f("a{}")
	f(`a{foo="bar"}`)
	f(`foo{bar="baz", x="y\"z"}`)
	f(`foo{bar="b}az"}`)
	f(`:foo:bar{bar="a",baz="b"}`)
	f(`some.foo{bar="baz"}`)
}

func TestValidateMetricError(t *testing.T) {
	f := func(s string) {
		t.Helper()
		if err := validateMetric(s); err == nil {
			t.Fatalf("expecting non-nil error when validating %q", s)
		}
	}
	f("")
	f("{}")

	// superflouos space
	f("a ")
	f(" a")
	f(" a ")
	f("a {}")
	f("a{} ")
	f("a{ }")
	f(`a{foo ="bar"}`)
	f(`a{ foo="bar"}`)
	f(`a{foo= "bar"}`)
	f(`a{foo="bar" }`)
	f(`a{foo="bar" ,baz="a"}`)

	// invalid tags
	f("a{foo}")
	f("a{=}")
	f(`a{=""}`)
	f(`a{`)
	f(`a}`)
	f(`a{foo=}`)
	f(`a{foo="`)
	f(`a{foo="}`)
	f(`a{foo="bar",}`)
	f(`a{foo="bar", x`)
	f(`a{foo="bar", x=`)
	f(`a{foo="bar", x="`)
	f(`a{foo="bar", x="}`)
}

func TestValidateUpperBoundBucketsError(t *testing.T) {
	f := func(b []float64) {
		t.Helper()
		if err := validateUpperBoundBuckets(b); err == nil {
			t.Fatalf("expecting non-nil error when validating %v", b)
		}
	}
	f([]float64{-1})
	f([]float64{math.NaN()})
	f([]float64{1, 0})
	f([]float64{1, 1})
}
