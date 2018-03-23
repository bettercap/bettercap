// +build linux

package serial

import (
	"os"
	"testing"
	"time"
)

func TestConnection(t *testing.T) {
	port0 := os.Getenv("PORT0")
	port1 := os.Getenv("PORT1")
	if port0 == "" || port1 == "" {
		t.Skip("Skipping test because PORT0 or PORT1 environment variable is not set")
	}
	c0 := &Config{Name: port0, Baud: 115200}
	c1 := &Config{Name: port1, Baud: 115200}

	s1, err := OpenPort(c0)
	if err != nil {
		t.Fatal(err)
	}

	s2, err := OpenPort(c1)
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan int, 1)
	go func() {
		buf := make([]byte, 128)
		var readCount int
		for {
			n, err := s2.Read(buf)
			if err != nil {
				t.Fatal(err)
			}
			readCount++
			t.Logf("Read %v %v bytes: % 02x %s", readCount, n, buf[:n], buf[:n])
			select {
			case <-ch:
				ch <- readCount
				close(ch)
			default:
			}
		}
	}()

	if _, err = s1.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if _, err = s1.Write([]byte(" ")); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
	if _, err = s1.Write([]byte("world")); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second / 10)

	ch <- 0
	s1.Write([]byte(" ")) // We could be blocked in the read without this
	c := <-ch
	exp := 5
	if c >= exp {
		t.Fatalf("Expected less than %v read, got %v", exp, c)
	}
}
