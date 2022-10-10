package hangar

import (
	"embed"
	"errors"
	"github.com/gabriel-vasile/mimetype"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
)

//go:embed static
var UnderlayFS embed.FS

type Hangar struct {
	BaseFS      fs.FS
	Template    *template.Template
	SiteName    string
	NoHeader    bool
	Banner      string
	AppName     string
	Attribution template.HTML
}

func New(base fs.FS) (*Hangar, error) {
	t, err := template.New("directory").ParseFS(UnderlayFS, "static/directory.html.tmpl")
	if err != nil {
		return nil, err
	}

	return &Hangar{
		BaseFS:      base,
		Template:    t,
		SiteName:    "DLBase",
		NoHeader:    false,
		Banner:      "/static/banner.jpg",
		AppName:     "GoHangar",
		Attribution: "<a href=\"https://github.com/Unkn0wnCat/gohangar\" target=\"_blank\">GoHangar</a>",
	}, nil
}

func (hangar *Hangar) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	requestUri, err := url.Parse(req.RequestURI)
	if err != nil {
		res.WriteHeader(400)
		res.Write([]byte("400 - Invalid Request"))
		return
	}

	currentPath := strings.Trim(requestUri.Path, "/")
	if currentPath == "" {
		currentPath = "."
	}

	stat, err := fs.Stat(hangar.BaseFS, currentPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if stat, err := fs.Stat(UnderlayFS, currentPath); err == nil && !strings.HasSuffix(currentPath, ".tmpl") && !stat.IsDir() {
				writeFile(UnderlayFS, currentPath, res)
				return
			}

			res.WriteHeader(404)
			res.Write([]byte("404 - Not Found"))
			return
		}

		log.Println(err)
		res.WriteHeader(500)
		res.Write([]byte("500 - Internal Server Error"))
		return
	}

	if !stat.IsDir() {
		writeFile(hangar.BaseFS, currentPath, res)
		return
	}

	hangar.listDir(hangar.BaseFS, currentPath, res)
	return
}

type DirectoryListingData struct {
	SiteName    string
	AppName     string
	Attribution template.HTML
	NoHeader    bool
	Banner      string
	Path        []string
	Entries     []DirectoryListingEntry
	Readme      string
}

type DirectoryListingEntry struct {
	AbsolutePath string
	Name         string
	Icon         string
}

func getIcon(fsys fs.FS, filePath string) (string, error) {
	file, err := fsys.Open(filePath)
	if err != nil {
		return "", err
	}

	mime, err := mimetype.DetectReader(file)
	if err != nil {
		return "", err
	}

	mimeType := mime.String()

	switch mimeType {
	case "text/html":
		return "page_html.gif", nil
	case "application/x-shockwave-flash":
		return "application_flash.gif", nil
	case "application/pdf":
		return "file_acrobat.gif", nil
	}

	switch strings.Split(mimeType, "/")[0] {
	case "image":
		return "image.gif", nil
	case "audio":
		return "page_sound.gif", nil
	case "video":
		return "page_video.gif", nil
	case "text":
		return "page_text.gif", nil
	default:
		return "page.gif", nil
	}
}

func (hangar *Hangar) listDir(fsys fs.FS, currentPath string, res http.ResponseWriter) {
	dir, err := fs.ReadDir(fsys, currentPath)
	if err != nil {
		log.Println(err)
		res.WriteHeader(500)
		res.Write([]byte("500 - Internal Server Error"))
		return
	}

	var entries []DirectoryListingEntry
	var dirs []DirectoryListingEntry

	readme := ""

	for _, entry := range dir {
		if entry.Name() == ".readme.html" {
			readme = path.Join(currentPath, entry.Name())
		}

		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		if entry.IsDir() {
			dirs = append(dirs, DirectoryListingEntry{
				AbsolutePath: path.Join(currentPath, entry.Name()),
				Name:         entry.Name(),
				Icon:         "/static/images/folder.gif",
			})
			continue
		}

		icon, err := getIcon(fsys, path.Join(currentPath, entry.Name()))
		if err != nil {
			log.Println(err)
			continue
		}

		entries = append(entries, DirectoryListingEntry{
			AbsolutePath: path.Join(currentPath, entry.Name()),
			Name:         entry.Name(),
			Icon:         "/static/images/" + icon,
		})
	}

	res.Header().Set("Content-Type", "text/html")
	res.WriteHeader(200)
	err = hangar.Template.ExecuteTemplate(res, "directory.html.tmpl", DirectoryListingData{
		SiteName:    hangar.SiteName,
		AppName:     hangar.AppName,
		Attribution: hangar.Attribution,
		NoHeader:    hangar.NoHeader,
		Banner:      hangar.Banner,
		Path:        strings.Split(currentPath, "/"),
		Entries:     append(dirs, entries...),
		Readme:      readme,
	})
	if err != nil {
		log.Println(err)
		res.WriteHeader(500)
		res.Write([]byte("500 - Internal Server Error"))
		return
	}

	return
}

func writeFile(fsys fs.FS, currentPath string, res http.ResponseWriter) {
	file, err := fsys.Open(currentPath)
	if err != nil {
		log.Println(err)
		res.WriteHeader(500)
		res.Write([]byte("500 - Internal Server Error"))
		return
	}

	mime, err := mimetype.DetectReader(file)
	if err != nil {
		_ = file.Close()
		log.Println(err)
		res.WriteHeader(500)
		res.Write([]byte("500 - Internal Server Error"))
		return
	}

	mimeType := mime.String()

	if strings.HasSuffix(currentPath, ".css") {
		mimeType = "text/css"
	}

	_ = file.Close()
	file, err = fsys.Open(currentPath)
	defer file.Close()
	if err != nil {
		log.Println(err)
		res.WriteHeader(500)
		res.Write([]byte("500 - Internal Server Error"))
		return
	}

	res.Header().Set("Content-Type", mimeType)
	res.WriteHeader(200)
	io.Copy(res, file)
	return
}
