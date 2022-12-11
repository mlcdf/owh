package remote

import (
	"testing"
)

func TestSkipFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "normal file",
			path: "index.html",
			want: false,
		},
		{
			name: "git folder",
			path: ".git",
			want: true,
		},
		{
			name: "file inside git folder",
			path: ".git/FETCH_HEAD",
			want: true,
		},
		{
			name: ".htaccess file",
			path: ".htaccess",
			want: false,
		},
		{
			name: "file instead node_modules",
			path: "node_modules/yolo/index.js",
			want: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			if got := skipFile(test.path); got != test.want {
				t.Errorf("want skipFile=%t, got %t for %s", got, test.want, test.path)
			}
		})
	}
}
