package utils

import "mime/multipart"

// MultipartFileHeaderToBytes reads an uploaded multipart file into memory.
func MultipartFileHeaderToBytes(fileHeader *multipart.FileHeader) ([]byte, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileBytes := make([]byte, fileHeader.Size)
	if _, err := file.Read(fileBytes); err != nil {
		return nil, err
	}
	return fileBytes, nil
}
