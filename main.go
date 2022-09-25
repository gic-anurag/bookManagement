package main

import (
	"BookManagement/dao"
	"BookManagement/model"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var bed = dao.BookDao{}

func getBookByName(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != "GET" {
		respondWithError(w, http.StatusBadRequest, "Method not allowed")
		return
	}

	name := strings.Split(r.URL.Path, "/")[2]

	books, err := bed.FindByBookName(name)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJson(w, http.StatusOK, books)
}

func createNewBook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != "POST" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
		return
	}

	var book model.Book

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if err := bed.Insert(book); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	} else {
		respondWithJson(w, http.StatusAccepted, map[string]string{
			"message": "Record inserted successfully",
		})
	}

}

func deleteBookByName(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != "DELETE" {
		respondWithError(w, http.StatusBadRequest, "Method not allowed")
		return
	}
	var reqBody map[string]string

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	name := reqBody["name"]

	err := bed.DeleteBook(name)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJson(w, http.StatusOK, map[string]string{
		"message": "Record deleted successfully",
	})
}

func updateBookByName(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != "PUT" {
		respondWithError(w, http.StatusBadRequest, "Method not allowed")
		return
	}
	var book model.Book
	err := json.NewDecoder(r.Body).Decode(&book)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	name := book.Name

	_ = bed.UpdateBook(name, book)

	respondWithJson(w, http.StatusOK, map[string]string{
		"message": "Record updated successfully",
	})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJson(w, code, map[string]string{"error": msg})
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func init() {
	bed.Server = "mongodb://localhost:27017/"
	bed.Database = "quickstart"
	bed.Collection = "book"
	bed.Connect()
}

func main() {
	http.HandleFunc("/add-book", createNewBook)
	http.HandleFunc("/get-book/", getBookByName)
	http.HandleFunc("/delete-book", deleteBookByName)
	http.HandleFunc("/update-book", updateBookByName)
	http.HandleFunc("/download-book", downloadFile)
	http.HandleFunc("/upload-book", uploadHandler)
	//	http.HandleFunc("/minio-setup", mac)
	fmt.Println("starting server at 9090")
	log.Fatal(http.ListenAndServe(":9090", nil))
}

var (
	fileName    string
	fullURLFile string
)

type FileDown struct {
	fullURLFile string
}

func downloadFile(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		respondWithError(w, http.StatusBadRequest, "Method not allowed")
		return
	}

	var downFile FileDown

	if err := json.NewDecoder(r.Body).Decode(&downFile); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	//	fullURLFile = "https://www.golang-book.com/public/pdf/gobook.pdf"

	// Build fileName from fullPath
	fileURL, err := url.Parse(downFile.fullURLFile)
	if err != nil {
		log.Fatal(err)
	}
	path := fileURL.Path
	segments := strings.Split(path, "/")
	fileName = segments[len(segments)-1]

	// Create blank file
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	// Put content on file
	resp, err := client.Get(downFile.fullURLFile)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)

	defer file.Close()

	fmt.Printf("Downloaded a file %s with size %d", fileName, size)

}

const MAX_SIZE = 10 * 1024 * 1024

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 32 MB is the default used by FormFile()
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get a reference to the fileHeaders.
	// They are accessible only after ParseMultipartForm is called
	files := r.MultipartForm.File["file"]

	for _, fileHeader := range files {
		// Restrict the size of each uploaded file to 1MB.
		// To prevent the aggregate size from exceeding
		// a specified value, use the http.MaxBytesReader() method
		// before calling ParseMultipartForm()
		if fileHeader.Size > MAX_SIZE {
			http.Error(w, fmt.Sprintf("The uploaded image is too big: %s. Please use an image less than 1MB in size", fileHeader.Filename), http.StatusBadRequest)
			return
		}

		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer file.Close()

		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//	File ristriction
		filetype := http.DetectContentType(buff)
		if filetype != "image/jpeg" && filetype != "image/png" {
			http.Error(w, "The provided file format is not allowed. Please upload a JPEG or PNG image", http.StatusBadRequest)
			return
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = os.MkdirAll("./uploads", os.ModePerm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		f, err := os.Create(fmt.Sprintf("./uploads/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	fmt.Fprintf(w, "Upload successful")
}

/*func mac(w http.ResponseWriter, r *http.Request) {
	r.Body.Close()

	if r.Method != "POST" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Get a reference to the fileHeaders.
	// They are accessible only after ParseMultipartForm is called
	files := r.MultipartForm.File["file"]

	for _, fileHeader := range files {
		// Restrict the size of each uploaded file to 1MB.
		// To prevent the aggregate size from exceeding
		// a specified value, use the http.MaxBytesReader() method
		// before calling ParseMultipartForm()
		if fileHeader.Size > MAX_SIZE {
			http.Error(w, fmt.Sprintf("The uploaded image is too big: %s. Please use an image less than 1MB in size", fileHeader.Filename), http.StatusBadRequest)
			return
		}
		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer file.Close()

		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//	File ristriction
		filetype := http.DetectContentType(buff)
		if filetype != "image/jpeg" && filetype != "image/json" {
			http.Error(w, "The provided file format is not allowed. Please upload a JPEG or PNG image", http.StatusBadRequest)
			return
		}
		Minio(filetype)
		fmt.Printf("r.Response.Header: %v\n", r.Response.Header)
	}
}*/
