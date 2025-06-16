package main

import (
	"os"
	"testing"
)

type ExampleConfig struct {
	BoolField   bool
	StringField string
}

func Test_Store(t *testing.T) {
	store, err := New[ExampleConfig]("./tmp/test.db", "ExampleConfig")
	defer os.RemoveAll("./tmp")

	if err != nil {
		t.Fatalf("failed to create config store: %v", err)
	}
	defer store.Close()

	original := ExampleConfig{
		BoolField:   true,
		StringField: "hello world",
	}

	err = store.Save(original)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loaded, err := store.Get()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded != original {
		t.Errorf("config mismatch.\nExpected: %+v\nGot: %+v", original, loaded)
	}
}
