package runtime

import "testing"

func TestNormalizeRuntimeMode(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{in: "", want: RuntimeVM},
		{in: "treewalk", want: RuntimeVM},
		{in: "TREEWALK", want: RuntimeVM},
		{in: "vm", want: RuntimeVM},
		{in: " VM ", want: RuntimeVM},
		{in: "jit", wantErr: true},
	}

	for _, tt := range tests {
		got, err := NormalizeRuntimeMode(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("NormalizeRuntimeMode(%q): expected error, got nil", tt.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("NormalizeRuntimeMode(%q): unexpected error: %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("NormalizeRuntimeMode(%q): got %q want %q", tt.in, got, tt.want)
		}
	}
}
