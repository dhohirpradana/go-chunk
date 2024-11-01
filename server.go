package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/google/uuid"
)

const uploadDir = "./uploads"

var fileLocks sync.Map

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // Limit upload size to 10MB per chunk

	fileID := r.FormValue("fileID")
	chunkIndex, err := strconv.Atoi(r.FormValue("chunkIndex"))
	if err != nil {
		http.Error(w, "Invalid chunk index", http.StatusBadRequest)
		return
	}
	totalChunks, err := strconv.Atoi(r.FormValue("totalChunks"))
	if err != nil {
		http.Error(w, "Invalid total chunks", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("fileChunk")
	if err != nil {
		http.Error(w, "Could not read file chunk", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	tempDir := fmt.Sprintf("%s/%s", uploadDir, fileID)
	os.MkdirAll(tempDir, os.ModePerm)

	chunkPath := fmt.Sprintf("%s/chunk_%d", tempDir, chunkIndex)
	outFile, err := os.Create(chunkPath)
	if err != nil {
		http.Error(w, "Could not write chunk", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		http.Error(w, "Could not save chunk", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Chunk %d received\n", chunkIndex)

	go tryAssembleFile(fileID, totalChunks)
}

func tryAssembleFile(fileID string, totalChunks int) {
	lock, _ := fileLocks.LoadOrStore(fileID, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	defer lock.(*sync.Mutex).Unlock()

	tempDir := fmt.Sprintf("%s/%s", uploadDir, fileID)
	for i := 0; i < totalChunks; i++ {
		chunkPath := fmt.Sprintf("%s/chunk_%d", tempDir, i)
		if _, err := os.Stat(chunkPath); os.IsNotExist(err) {
			return
		}
	}

	newUUID := uuid.New().String()
	outputPath := fmt.Sprintf("%s/%s_%s", uploadDir, fileID, newUUID)
	finalFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Println("Could not create final file:", err)
		return
	}
	defer finalFile.Close()

	for i := 0; i < totalChunks; i++ {
		chunkPath := fmt.Sprintf("%s/chunk_%d", tempDir, i)
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			fmt.Println("Could not open chunk:", err)
			return
		}
		io.Copy(finalFile, chunkFile)
		chunkFile.Close()
		os.Remove(chunkPath)
	}

	os.Remove(tempDir)
	fmt.Printf("File %s assembled successfully\n", fileID)
}

func main() {
	os.MkdirAll(uploadDir, os.ModePerm)

	http.HandleFunc("/upload", UploadHandler)
	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
