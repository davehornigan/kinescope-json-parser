package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

type PlaylistItem struct {
	Title string `json:"title"`
}

type Input struct {
	URL     string `json:"url"`
	Referer string `json:"referer,omitempty"`
	Options struct {
		Playlist []PlaylistItem `json:"playlist"`
	} `json:"options"`
}

func main() {
	http.HandleFunc("/", formHandler)
	http.HandleFunc("/upload", uploadHandler)

	fmt.Println("Server running at http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	html := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title>JSON Multi-Upload</title>
		</head>
		<body>
			<h1>Upload multiple JSON files</h1>
			<form enctype="multipart/form-data" action="/upload" method="post">
				<input type="file" name="files" multiple><br><br>
				<label for="path">Optional Path: </label>
				<input type="text" name="path" id="path" placeholder="Path (optional)"><br><br>
				<label>
					<input type="checkbox" name="use_filename" value="1">
					Use filename instead title
				</label><br><br>
				<input type="submit" value="Upload">
			</form>
		</body>
		</html>`
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(html))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	path := strings.TrimSpace(r.FormValue("path"))
	useFilename := r.FormValue("use_filename") == "1"

	var commands []string
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Error opening file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		var input Input
		body, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Error reading file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(body, &input); err != nil {
			http.Error(w, "Error decoding JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		cleanURL := trimURLParams(input.URL)

		var finalTitle = ""
		if useFilename {
			base := filepath.Base(fileHeader.Filename)
			finalTitle = strings.TrimSuffix(base, ".json") + ".mp4"
		} else {
			if len(input.Options.Playlist) > 0 {
				finalTitle = input.Options.Playlist[0].Title
			}
		}
		if path != "" {
			finalTitle = path + finalTitle
		}

		line := "kinescope-dl.exe"
		if strings.TrimSpace(input.Referer) != "" {
			line += " -r " + escapeArg(input.Referer)
		}
		line += " " + escapeArg(cleanURL)
		line += " " + escapeArg(finalTitle)
		commands = append(commands, line)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=\"commands.txt\"")
	for _, cmd := range commands {
		_, _ = fmt.Fprintln(w, cmd)
	}
}

func trimURLParams(url string) string {
	if idx := strings.Index(url, "?"); idx != -1 {
		return url[:idx]
	}
	return url
}

func escapeArg(arg string) string {
	if strings.ContainsAny(arg, " \t\"") {
		return "\"" + strings.ReplaceAll(arg, "\"", "\\\"") + "\""
	}
	return arg
}
