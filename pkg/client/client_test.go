package client

import (
	"net/url"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
)

func TestBottleDetailURL(t *testing.T) {
	mustParse := func(u string) url.URL {
		t.Helper()
		uu, err := url.Parse(u)
		assert.NoError(t, err)
		return *uu
	}
	type args struct {
		serverURL url.URL
		dgst      digest.Digest
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"basic", args{mustParse("https://foo.com"), digest.Digest("sha:123")}, "https://foo.com/www/bottle.html?digest=sha%3A123"},
		{"slash", args{mustParse("https://foo.com/"), digest.Digest("sha:123")}, "https://foo.com/www/bottle.html?digest=sha%3A123"},
		{"path", args{mustParse("https://foo.com/foo"), digest.Digest("sha:123")}, "https://foo.com/foo/www/bottle.html?digest=sha%3A123"},
		{"path slash", args{mustParse("https://foo.com/foo/"), digest.Digest("sha:123")}, "https://foo.com/foo/www/bottle.html?digest=sha%3A123"},
		{"query-string", args{mustParse("https://foo.com/foo?bar=go"), digest.Digest("sha:123")}, "https://foo.com/foo/www/bottle.html?bar=go&digest=sha%3A123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BottleDetailURL(tt.args.serverURL, tt.args.dgst)
			assert.Equal(t, tt.want, got)
		})
	}
}
