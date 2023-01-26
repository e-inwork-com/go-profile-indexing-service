package grpc

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"
	"testing"
)

func (app *Application) testFormProfile(t *testing.T) (io.Reader, string) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// Add team name
	bodyWriter.WriteField("profile_name", "John Doe")

	// Add team picture
	filename := "./test/images/profile.jpg"
	fileWriter, err := bodyWriter.CreateFormFile("profile_picture", filename)
	if err != nil {
		t.Fatal(err)
	}

	// Open file
	fileHandler, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}

	// Copy file
	_, err = io.Copy(fileWriter, fileHandler)
	if err != nil {
		t.Fatal(err)
	}

	// Put on body
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	return bodyBuf, contentType
}
