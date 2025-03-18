package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	goruntime "runtime"
	"testing"

	"github.com/mholt/archives"
)

func newTestApp() *App {
	return &App{
		ctx:   context.Background(),
		cache: NewFileLoader(),
	}
}

func seqReader(ctx context.Context, filename string) map[int][]byte {
	book := make(map[int][]byte)
	var counter int
	fsys, err := archives.FileSystem(ctx, filename, nil)
	if err != nil {
		return nil
	}

	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}

		if err != nil {
			return err
		}

		if isImage(path) {
			f, err := fsys.Open(path)
			if err != nil {
				return err
			}

			defer f.Close()
			var data bytes.Buffer

			if _, err = io.Copy(&data, f); err != nil {
				return err
			}

			book[counter] = data.Bytes()
			counter++
		}

		return nil
	})
	return book
}

var mockHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
})

func TestBookOpen(t *testing.T) {
	app := newTestApp()

	_, err := app.LoadBook("./lib/Chew.cbr")
	if err != nil {
		t.Errorf("An error occurred opening this book")
	}

	length := app.Length()

	fmt.Println(length)

	if length != 28 {
		t.Errorf("Incorrect page number")
	}
}

func TestBadBookOpen(t *testing.T) {
	app := newTestApp()

	_, err := app.LoadBook("./lib/book(some bad).cbr")
	if err == nil {
		t.Fatalf("Expected an error")
	}

	if err.Error() != "invalid cbr file" {
		t.Fatalf("Expected an invalid cbr file error")
	}

	gophers := goruntime.NumGoroutine()

	if gophers >= 3 {
		t.Fatalf("zombie goroutines lurking")
	}
}

func TestEmptyBookOpen(t *testing.T) {
	app := newTestApp()

	_, err := app.LoadBook("./lib/empty.cbr")
	if err == nil {
		t.Fatalf("Expected an error")
	}

	if err.Error() != "cannot open an empty cbr" {
		t.Fatalf("Expected a \"cannot open an empty cbr\" error")
	}
}

func TestBookOrder(t *testing.T) {
	app := newTestApp()
	cache := app.cache
	_, err := app.LoadBook("./lib/Batman - Year One.cbr")
	if err != nil {
		fmt.Println(err)
		t.Errorf("An error occurred opening this book")
	}

	seq := seqReader(app.ctx, "./lib/Batman - Year One.cbr")
	if seq == nil {
		fmt.Println(err)
		t.Errorf("Sequetial reader has failed")
	}

	for k, _ := range seq {
		if eq := bytes.Equal(seq[k], cache.cache[k]); eq == false {
			t.Errorf("Page %d does not match", k)
		}
	}

}

func TestTraverseBook(t *testing.T) {
	app := newTestApp()
	_, err := app.LoadBook("./lib/Batman - Year One.cbr")
	length := app.Length()

	if err != nil {
		fmt.Println(err)
		t.Errorf("An error occurred opening this book")
	}

	var route string

	middleware := app.cache.TraverseBook(mockHandler)

	for i := 0; i < length; i++ {
		route = fmt.Sprintf("/image/%d", i)
		req := httptest.NewRequest(http.MethodGet, route, nil)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			fmt.Println(rr.Code)
			t.Fatal("An error has occurred")
		}

		page := rr.Body.Bytes()

		if !bytes.Equal(page, app.cache.cache[i]) {
			t.Fatal("An error has occurred page do not match")
		}

	}

}
