package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"wsock"
)

var TemplateDir string
var StaticDir string

func GetAppDir() string {
	apppath, err := exec.LookPath(os.Args[0])
	if err != nil {
		panic(err)
	}
	apppath, err = filepath.Abs(apppath)
	if err != nil {
		panic(err)
	}
	dir, _ := path.Split(apppath)
	return dir
}

const indexHtml = `<script src="/socket.io/socket.io.js"></script>
<script>
  var socket = io.connect('http://localhost');
  socket.on('news', function (data) {
	console.log(data);
	socket.send('O hai!');
	socket.on('message', function(data) { console.log('Get message ' + data); });
  });
</script>`

func handleStatic(w http.ResponseWriter, r *http.Request) {
	re := regexp.MustCompile(`/static/(.*)`)
	results := re.FindStringSubmatch(r.URL.Path)
	filename := results[1]
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, path.Join(StaticDir, filename))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tfile := path.Join(TemplateDir, "index.html")
	templ, err := template.ParseFiles(tfile)
	if err != nil {
		fmt.Printf("Could not parse template: %s\n", err.Error())
	}
	templ.Execute(w, nil)
}

func socketConnectionHandler(inst *wsock.SocketInstance) error {
	err := inst.ReadLoop()
	fmt.Printf("connection closed %q\n", err)
	return err
}

func socketHandler(inst *wsock.SocketInstance, length int64, data io.Reader) error{
	d := make([]byte, length)
	data.Read(d)
	fmt.Printf("Message: %q\n", d)
	inst.Write(d)
	return nil
}

var comm chan []byte

func scanStdin() {
	buf := make([]byte, 1024)
	for {
		nr, _ := os.Stdin.Read(buf)
		if nr > 0 {
			comm <- buf[0:nr]
			fmt.Printf("read %s\n", buf[0:nr])
		}
	}
}
func main() {
	comm = make(chan []byte, 5)
	appDir := GetAppDir()
	TemplateDir = path.Join(appDir, "..", "templates")
	StaticDir = path.Join(appDir, "..", "static")
	http.HandleFunc(`/`, handleIndex)
	http.HandleFunc(`/static/`, handleStatic)
	http.Handle(`/news`, wsock.SocketHandlerFuncs(socketConnectionHandler, socketHandler))
	fmt.Printf("listening on localhost:8082\n")
	go scanStdin()
	http.ListenAndServe(":8082", nil)
}
