package session

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

var (
	testEnvFile = "/tmp/test.env"
	testEnvData = map[string]string{
		"people": "shit",
		"moo":    "boo",
		"foo":    "bar",
	}
	testEnvSorted = []string{"foo", "moo", "people"}
)

func setup(t testing.TB, envFile bool, envFileData bool) {
	teardown(t)

	if envFile {
		if fp, err := os.OpenFile(testEnvFile, os.O_RDONLY|os.O_CREATE, 0666); err == nil {
			fp.Close()
		} else {
			panic(err)
		}
	}

	if envFileData {
		if raw, err := json.Marshal(testEnvData); err != nil {
			panic(err)
		} else if err = ioutil.WriteFile(testEnvFile, raw, 0755); err != nil {
			panic(err)
		}
	}
}

func teardown(t testing.TB) {
	if err := os.RemoveAll(testEnvFile); err != nil {
		panic(err)
	}
}

func TestSessionEnvironmentWithoutFile(t *testing.T) {
	if env, err := NewEnvironment(""); env == nil {
		t.Fatal("expected valid environment")
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if len(env.Data) != 0 {
		t.Fatalf("expected empty environment, found %d elements", len(env.Data))
	}
}

func TestSessionEnvironmentWithInvalidFile(t *testing.T) {
	if env, err := NewEnvironment("/idontexist"); env == nil {
		t.Fatal("expected valid environment")
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if len(env.Data) != 0 {
		t.Fatalf("expected empty environment, found %d elements", len(env.Data))
	}
}

func TestSessionEnvironmentWithEmptyFile(t *testing.T) {
	setup(t, true, false)
	defer teardown(t)

	if env, err := NewEnvironment(testEnvFile); env == nil {
		t.Fatal("expected environment")
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if len(env.Data) != 0 {
		t.Fatalf("expected empty environment, found %d elements", len(env.Data))
	}
}

func TestSessionEnvironmentWithDataFile(t *testing.T) {
	setup(t, true, true)
	defer teardown(t)

	if env, err := NewEnvironment(testEnvFile); env == nil {
		t.Fatal("expected environment")
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if len(env.Data) != len(testEnvData) {
		t.Fatalf("expected %d, found %d", len(testEnvData), len(env.Data))
	} else if !reflect.DeepEqual(env.Data, testEnvData) {
		t.Fatalf("unexpected contents: %v", env.Data)
	}
}

func TestSessionEnvironmentSaveWithError(t *testing.T) {
	setup(t, false, false)
	defer teardown(t)

	if env, err := NewEnvironment(testEnvFile); env == nil {
		t.Fatal("expected environment")
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if err = env.Save("/lulz/nope"); err == nil {
		t.Fatal("expected error")
	}
}

func TestSessionEnvironmentSave(t *testing.T) {
	setup(t, false, false)
	defer teardown(t)

	env, err := NewEnvironment(testEnvFile)
	if env == nil {
		t.Fatal("expected environment")
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	env.Data["new"] = "value"
	if err = env.Save(testEnvFile); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env, err := NewEnvironment(testEnvFile); env == nil {
		t.Fatal("expected environment")
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if !reflect.DeepEqual(env.Data, map[string]string{"new": "value"}) {
		t.Fatalf("unexpected contents: %v", env.Data)
	}
}

func TestSessionEnvironmentHas(t *testing.T) {
	setup(t, true, true)
	defer teardown(t)

	env, err := NewEnvironment(testEnvFile)
	if env == nil {
		t.Fatal("expected environment")
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if len(env.Data) != len(testEnvData) {
		t.Fatalf("expected %d, found %d", len(testEnvData), len(env.Data))
	}

	for k := range testEnvData {
		if !env.Has(k) {
			t.Fatalf("could not find key '%s'", k)
		}
	}

	for _, k := range []string{"these", "keys", "should", "not", "be", "found"} {
		if env.Has(k) {
			t.Fatalf("unexpected key '%s'", k)
		}
	}
}

func TestSessionEnvironmentSet(t *testing.T) {
	setup(t, true, true)
	defer teardown(t)

	env, err := NewEnvironment(testEnvFile)
	if env == nil {
		t.Fatal("expected environment")
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if old := env.Set("people", "ok"); old != "shit" {
		t.Fatalf("unexpected old value: %s", old)
	} else if env.Data["people"] != "ok" {
		t.Fatalf("unexpected new value: %s", env.Data["people"])
	} else if old := env.Set("newkey", "nk"); old != "" {
		t.Fatalf("unexpected old value: %s", old)
	} else if env.Data["newkey"] != "nk" {
		t.Fatalf("unexpected new value: %s", env.Data["newkey"])
	}
}

func TestSessionEnvironmentSetWithCallback(t *testing.T) {
	setup(t, true, true)
	defer teardown(t)

	env, err := NewEnvironment(testEnvFile)
	if env == nil {
		t.Fatal("expected environment")
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cbCalled := false
	old := env.WithCallback("people", "ok", func(newValue string) {
		cbCalled = true
	})
	if old != "shit" {
		t.Fatalf("unexpected old value: %s", old)
	}

	cbCalled = false
	old = env.Set("people", "shitagain")
	if old != "ok" {
		t.Fatalf("unexpected old value: %s", old)
	} else if !cbCalled {
		t.Fatal("callback has not been called")
	}

	cbCalled = false
	env.Set("something", "else")
	if cbCalled {
		t.Fatal("callback should not have been called")
	}
}

func TestSessionEnvironmentGet(t *testing.T) {
	setup(t, true, true)
	defer teardown(t)

	env, err := NewEnvironment(testEnvFile)
	if env == nil {
		t.Fatal("expected environment")
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if len(env.Data) != len(testEnvData) {
		t.Fatalf("expected %d, found %d", len(testEnvData), len(env.Data))
	}

	for k, v := range testEnvData {
		if found, vv := env.Get(k); !found {
			t.Fatalf("should have found %s", k)
		} else if v != vv {
			t.Fatalf("unexpected value found: %s", vv)
		}
	}

	for _, k := range []string{"these", "keys", "should", "not", "be", "found"} {
		if found, _ := env.Get(k); found {
			t.Fatalf("should not have found %s", k)
		}
	}
}

func TestSessionEnvironmentGetInt(t *testing.T) {
	setup(t, true, true)
	defer teardown(t)

	env, err := NewEnvironment(testEnvFile)
	if env == nil {
		t.Fatal("expected environment")
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if len(env.Data) != len(testEnvData) {
		t.Fatalf("expected %d, found %d", len(testEnvData), len(env.Data))
	}

	for k := range testEnvData {
		if err, _ := env.GetInt(k); err == nil {
			t.Fatal("expected error")
		}
	}

	env.Data["num"] = "1234"
	if err, i := env.GetInt("num"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if i != 1234 {
		t.Fatalf("unexpected integer: %d", i)
	}

	if err, _ := env.GetInt("unknownint"); err == nil {
		t.Fatalf("expected error (unknown key): %v", err)
	}
}

func TestSessionEnvironmentSorted(t *testing.T) {
	setup(t, true, true)
	defer teardown(t)

	env, err := NewEnvironment(testEnvFile)
	if env == nil {
		t.Fatal("expected environment")
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if len(env.Data) != len(testEnvData) {
		t.Fatalf("expected %d, found %d", len(testEnvData), len(env.Data))
	} else if sorted := env.Sorted(); !reflect.DeepEqual(sorted, testEnvSorted) {
		t.Fatalf("unexpected sorted keys: %v", sorted)
	}
}
