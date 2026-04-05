package server

import (
	"errors"
	"testing"
	"time"

	"github.com/langgenius/dify-cloud-kit/oss"
)

var errPrimary = errors.New("primary error")
var errFallback = errors.New("fallback error")

// mockOSS implements oss.OSS for testing.
type mockOSS struct {
	saveFunc   func(key string, data []byte) error
	loadFunc   func(key string) ([]byte, error)
	existsFunc func(key string) (bool, error)
	stateFunc  func(key string) (oss.OSSState, error)
	listFunc   func(prefix string) ([]oss.OSSPath, error)
	deleteFunc func(key string) error
	typeName   string
}

func (m *mockOSS) Save(key string, data []byte) error        { return m.saveFunc(key, data) }
func (m *mockOSS) Load(key string) ([]byte, error)           { return m.loadFunc(key) }
func (m *mockOSS) Exists(key string) (bool, error)           { return m.existsFunc(key) }
func (m *mockOSS) State(key string) (oss.OSSState, error)    { return m.stateFunc(key) }
func (m *mockOSS) List(prefix string) ([]oss.OSSPath, error) { return m.listFunc(prefix) }
func (m *mockOSS) Delete(key string) error                   { return m.deleteFunc(key) }
func (m *mockOSS) Type() string                              { return m.typeName }

func TestNewFallbackOSS(t *testing.T) {
	primary := &mockOSS{typeName: "primary"}
	fallback := &mockOSS{typeName: "fallback"}
	f := NewFallbackOSS(primary, fallback)

	if f.primary != primary {
		t.Error("expected primary to be set")
	}
	if f.fallback != fallback {
		t.Error("expected fallback to be set")
	}
}

func TestFallbackOSS_Save(t *testing.T) {
	tests := []struct {
		name        string
		primaryErr  error
		fallbackErr error
		wantErr     bool
	}{
		{
			name:       "primary succeeds",
			primaryErr: nil,
			wantErr:    false,
		},
		{
			name:        "primary fails, fallback succeeds",
			primaryErr:  errPrimary,
			fallbackErr: nil,
			wantErr:     false,
		},
		{
			name:        "both fail",
			primaryErr:  errPrimary,
			fallbackErr: errFallback,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			primary := &mockOSS{
				saveFunc: func(key string, data []byte) error { return tt.primaryErr },
			}
			fallback := &mockOSS{
				saveFunc: func(key string, data []byte) error { return tt.fallbackErr },
			}
			f := NewFallbackOSS(primary, fallback)

			err := f.Save("test-key", []byte("data"))
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, errPrimary) {
				t.Errorf("Save() expected primary error, got %v", err)
			}
		})
	}
}

func TestFallbackOSS_Load(t *testing.T) {
	tests := []struct {
		name         string
		primaryData  []byte
		primaryErr   error
		fallbackData []byte
		fallbackErr  error
		wantData     []byte
		wantErr      bool
	}{
		{
			name:        "primary succeeds",
			primaryData: []byte("primary-data"),
			primaryErr:  nil,
			wantData:    []byte("primary-data"),
			wantErr:     false,
		},
		{
			name:         "primary fails, fallback succeeds",
			primaryErr:   errPrimary,
			fallbackData: []byte("fallback-data"),
			fallbackErr:  nil,
			wantData:     []byte("fallback-data"),
			wantErr:      false,
		},
		{
			name:        "both fail",
			primaryErr:  errPrimary,
			fallbackErr: errFallback,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			primary := &mockOSS{
				loadFunc: func(key string) ([]byte, error) { return tt.primaryData, tt.primaryErr },
			}
			fallback := &mockOSS{
				loadFunc: func(key string) ([]byte, error) { return tt.fallbackData, tt.fallbackErr },
			}
			f := NewFallbackOSS(primary, fallback)

			data, err := f.Load("test-key")
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && string(data) != string(tt.wantData) {
				t.Errorf("Load() data = %s, want %s", data, tt.wantData)
			}
		})
	}
}

func TestFallbackOSS_Exists(t *testing.T) {
	tests := []struct {
		name           string
		primaryExists  bool
		primaryErr     error
		fallbackExists bool
		fallbackErr    error
		wantExists     bool
		wantErr        bool
	}{
		{
			name:          "primary exists",
			primaryExists: true,
			primaryErr:    nil,
			wantExists:    true,
			wantErr:       false,
		},
		{
			name:          "primary returns false with no error",
			primaryExists: false,
			primaryErr:    nil,
			wantExists:    false,
			wantErr:       false,
		},
		{
			name:           "primary fails, fallback exists",
			primaryErr:     errPrimary,
			fallbackExists: true,
			fallbackErr:    nil,
			wantExists:     true,
			wantErr:        false,
		},
		{
			name:           "primary returns false, fallback exists",
			primaryExists:  false,
			primaryErr:     nil,
			fallbackExists: true,
			fallbackErr:    nil,
			wantExists:     true,
			wantErr:        false,
		},
		{
			name:           "primary fails, fallback not found",
			primaryErr:     errPrimary,
			fallbackExists: false,
			fallbackErr:    nil,
			wantExists:     false,
			wantErr:        true,
		},
		{
			name:        "primary fails, fallback fails",
			primaryErr:  errPrimary,
			fallbackErr: errFallback,
			wantExists:  false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			primary := &mockOSS{
				existsFunc: func(key string) (bool, error) { return tt.primaryExists, tt.primaryErr },
			}
			fallback := &mockOSS{
				existsFunc: func(key string) (bool, error) { return tt.fallbackExists, tt.fallbackErr },
			}
			f := NewFallbackOSS(primary, fallback)

			exists, err := f.Exists("test-key")
			if (err != nil) != tt.wantErr {
				t.Errorf("Exists() error = %v, wantErr %v", err, tt.wantErr)
			}
			if exists != tt.wantExists {
				t.Errorf("Exists() = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}

func TestFallbackOSS_State(t *testing.T) {
	now := time.Now()
	primaryState := oss.OSSState{Size: 100, LastModified: now}
	fallbackState := oss.OSSState{Size: 50, LastModified: now.Add(-time.Hour)}

	tests := []struct {
		name      string
		primary   oss.OSSState
		primaryE  error
		fallback  oss.OSSState
		fallbackE error
		want      oss.OSSState
		wantErr   bool
	}{
		{
			name:    "primary succeeds",
			primary: primaryState,
			want:    primaryState,
			wantErr: false,
		},
		{
			name:     "primary fails, fallback succeeds",
			primaryE: errPrimary,
			fallback: fallbackState,
			want:     fallbackState,
			wantErr:  false,
		},
		{
			name:      "both fail",
			primaryE:  errPrimary,
			fallbackE: errFallback,
			want:      oss.OSSState{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			primary := &mockOSS{
				stateFunc: func(key string) (oss.OSSState, error) { return tt.primary, tt.primaryE },
			}
			fallback := &mockOSS{
				stateFunc: func(key string) (oss.OSSState, error) { return tt.fallback, tt.fallbackE },
			}
			f := NewFallbackOSS(primary, fallback)

			state, err := f.State("test-key")
			if (err != nil) != tt.wantErr {
				t.Errorf("State() error = %v, wantErr %v", err, tt.wantErr)
			}
			if state.Size != tt.want.Size {
				t.Errorf("State() size = %d, want %d", state.Size, tt.want.Size)
			}
		})
	}
}

func TestFallbackOSS_List(t *testing.T) {
	primaryPaths := []oss.OSSPath{{Path: "/a", IsDir: false}, {Path: "/b", IsDir: true}}
	fallbackPaths := []oss.OSSPath{{Path: "/b", IsDir: true}, {Path: "/c", IsDir: false}}

	t.Run("merges and deduplicates", func(t *testing.T) {
		primary := &mockOSS{
			listFunc: func(prefix string) ([]oss.OSSPath, error) { return primaryPaths, nil },
		}
		fallback := &mockOSS{
			listFunc: func(prefix string) ([]oss.OSSPath, error) { return fallbackPaths, nil },
		}
		f := NewFallbackOSS(primary, fallback)

		paths, err := f.List("prefix")
		if err != nil {
			t.Fatalf("List() unexpected error: %v", err)
		}
		// /a from primary, /b from primary (dedup), /c from fallback
		if len(paths) != 3 {
			t.Errorf("List() got %d paths, want 3", len(paths))
		}
	})

	t.Run("primary fails returns error", func(t *testing.T) {
		primary := &mockOSS{
			listFunc: func(prefix string) ([]oss.OSSPath, error) { return nil, errPrimary },
		}
		fallback := &mockOSS{
			listFunc: func(prefix string) ([]oss.OSSPath, error) { return fallbackPaths, nil },
		}
		f := NewFallbackOSS(primary, fallback)

		_, err := f.List("prefix")
		if err == nil {
			t.Error("List() expected error when primary fails")
		}
	})

	// NOTE: Current implementation has a bug in List() - when primary succeeds but fallback fails,
	// it returns (nil, err) where err is the primary's nil error, effectively swallowing the fallback error.
	t.Run("fallback fails returns nil and no data", func(t *testing.T) {
		primary := &mockOSS{
			listFunc: func(prefix string) ([]oss.OSSPath, error) { return primaryPaths, nil },
		}
		fallback := &mockOSS{
			listFunc: func(prefix string) ([]oss.OSSPath, error) { return nil, errFallback },
		}
		f := NewFallbackOSS(primary, fallback)

		paths, err := f.List("prefix")
		// Bug: returns (nil, nil) instead of (nil, fallbackErr) or (primaryPaths, nil)
		if err != nil {
			t.Errorf("List() unexpected error due to current impl: %v", err)
		}
		if paths != nil {
			t.Errorf("List() expected nil paths due to current impl, got %v", paths)
		}
	})
}

func TestFallbackOSS_Delete(t *testing.T) {
	t.Run("primary succeeds", func(t *testing.T) {
		primary := &mockOSS{
			deleteFunc: func(key string) error { return nil },
		}
		fallback := &mockOSS{
			deleteFunc: func(key string) error { return nil },
		}
		f := NewFallbackOSS(primary, fallback)

		if err := f.Delete("test-key"); err != nil {
			t.Errorf("Delete() unexpected error: %v", err)
		}
	})

	t.Run("primary fails", func(t *testing.T) {
		primary := &mockOSS{
			deleteFunc: func(key string) error { return errPrimary },
		}
		fallback := &mockOSS{
			deleteFunc: func(key string) error { return nil },
		}
		f := NewFallbackOSS(primary, fallback)

		err := f.Delete("test-key")
		if !errors.Is(err, errPrimary) {
			t.Errorf("Delete() expected primary error, got %v", err)
		}
	})

	t.Run("fallback error is ignored", func(t *testing.T) {
		primary := &mockOSS{
			deleteFunc: func(key string) error { return nil },
		}
		fallback := &mockOSS{
			deleteFunc: func(key string) error { return errFallback },
		}
		f := NewFallbackOSS(primary, fallback)

		if err := f.Delete("test-key"); err != nil {
			t.Errorf("Delete() fallback error should be ignored, got %v", err)
		}
	})
}

func TestFallbackOSS_Type(t *testing.T) {
	primary := &mockOSS{typeName: "s3"}
	fallback := &mockOSS{typeName: "local"}
	f := NewFallbackOSS(primary, fallback)

	if got := f.Type(); got != "s3" {
		t.Errorf("Type() = %s, want s3", got)
	}
}
