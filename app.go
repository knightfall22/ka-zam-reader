package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"

	goRuntime "runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/mholt/archives"
	"github.com/wailsapp/mimetype"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sync/errgroup"
)

// App struct
type App struct {
	ctx   context.Context
	cache *FileLoader
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		cache: NewFileLoader(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

type FileLoader struct {
	http.Handler
	mu      sync.Mutex
	cache   map[int][]byte
	counter int
}

type bookOrder struct {
	id   int
	path string
}
type result struct {
	id   int
	data []byte
}

func NewFileLoader() *FileLoader {
	return &FileLoader{
		cache: make(map[int][]byte),
	}
}

func writebook(ctx context.Context, order chan bookOrder, fsys fs.FS, limits chan bool, resultChan chan result) error {
	limits <- true

	defer func() {
		<-limits
	}()

	page := <-order

	f, err := fsys.Open(page.path)
	if err != nil {
		return err
	}

	typ, file, err := recycleReader(f)
	if err != nil {
		return err
	}

	if !mimetype.EqualsAny(typ, "image/png", "image/jpeg", "image/jpg") {
		return fmt.Errorf("invalid cbr file")
	}

	defer f.Close()
	var data bytes.Buffer

	if _, err = io.Copy(&data, file); err != nil {
		return err
	}

	select {
	case resultChan <- result{
		id:   page.id,
		data: data.Bytes(),
	}:
	case <-ctx.Done():
		return ctx.Err()

	}

	return nil
}

func (h *FileLoader) StoreBook(ctx context.Context, filename string) (id int, err error) {
	workers := 2 * goRuntime.GOMAXPROCS(0)
	limits := make(chan bool, workers)
	order := make(chan bookOrder, workers)
	resultChan := make(chan result)
	counter := 0

	h.cache = make(map[int][]byte)

	go func() {
		for d := range resultChan {
			h.mu.Lock()
			h.cache[d.id] = d.data
			h.mu.Unlock()
		}
	}()

	fsys, err := archives.FileSystem(ctx, filename, nil)
	if err != nil {
		return 0, err
	}

	g, ctx := errgroup.WithContext(ctx)

	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}

		if err != nil {
			return err
		}

		if isImage(path) {
			order <- bookOrder{
				id:   counter,
				path: path,
			}
			counter++

			g.Go(func() error { return writebook(ctx, order, fsys, limits, resultChan) })

		}

		return nil
	})

	err = g.Wait()
	close(resultChan)

	if err != nil {
		return 0, err
	}

	if len(h.cache) == 0 {
		return 0, fmt.Errorf("cannot open an empty cbr")
	}

	log.Printf("[INFO] This is the cache length: %d\n", len(h.cache))

	return 0, nil
}

func (h *FileLoader) GetImage(id int) ([]byte, bool) {
	h.mu.Lock()
	data, ok := h.cache[id]
	defer h.mu.Unlock()

	if !ok {
		return nil, false
	}

	return data, ok
}

func (a *App) Length() int {
	return len(a.cache.cache)
}

func (a *App) LoadBook(filename string) (id int, err error) {
	id, err = a.cache.StoreBook(a.ctx, filename)
	log.Print(filename)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (h *FileLoader) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	requestedFileName := strings.TrimPrefix(req.URL.Path, "/")
	println("Requesting File:", requestedFileName)
	file, err := os.ReadFile(requestedFileName)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(fmt.Sprintf("Could not load file %s", requestedFileName)))
	}

	res.Write(file)
}

func (h *FileLoader) TraverseBook(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if !strings.HasPrefix(req.URL.Path, "/image") {
			log.Println("Hello")
			next.ServeHTTP(res, req)
			return
		}

		strId := strings.TrimPrefix(req.URL.Path, "/image/")

		if strId == "" {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte(fmt.Sprintf("Could not load file id does not exist")))
			return
		}

		id, err := strconv.Atoi(strId)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte(fmt.Sprintf("Could not load file id does not exist")))
			return
		}

		data, ok := h.GetImage(id)
		if !ok {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte(fmt.Sprintf("Could not load file id does not exist")))
			return
		}

		res.WriteHeader(http.StatusOK)
		res.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate;")
		res.Header().Set("Pragma", "no-cache")
		res.Header().Set("Expires", "0")
		res.Write(data)

	})
}

func (a *App) SelectFile() (string, error) {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select a file",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Comics (*.rar;*.cbr;*.cbz)",
				Pattern:     "*.rar;*.cbr;*.cbz",
			},
		},
	})

	if err != nil {
		log.Println("Error opening file dialog:", err)
		return "", err
	}
	return file, nil
}

func isImage(name string) bool {
	return strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".jpeg")
}

func recycleReader(input io.Reader) (mimeType string, recycled io.Reader, err error) {
	header := bytes.NewBuffer(nil)

	mtype, err := mimetype.DetectReader(io.TeeReader(input, header))
	if err != nil {
		return
	}

	recycled = io.MultiReader(header, input)

	return mtype.String(), recycled, nil
}
