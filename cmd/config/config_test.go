package config

import "testing"

func ptr[T any](v T) *T { return &v }

func TestEvaluateRootValidation(t *testing.T) {
	tests := []struct {
		name    string
		pkg     string
		root    *string
		wantErr bool
	}{
		{name: "unset root is fine", pkg: "pretty", root: nil},
		{name: "domain root is a prefix of anything", pkg: "pretty", root: ptr("")},
		{name: "root equal to the path", pkg: "svc", root: ptr("svc")},
		{name: "root is a path prefix", pkg: "svc/kamino", root: ptr("svc")},
		{name: "root is not a prefix", pkg: "pretty", root: ptr("svc"), wantErr: true},
		{name: "partial segment is not a prefix", pkg: "svcx/kamino", root: ptr("svc"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{Packages: PackageConfig{
				tt.pkg: {Target: "https://example.com/repo", Root: tt.root},
			}}
			err := c.Evaluate()
			if tt.wantErr != (err != nil) {
				t.Fatalf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
