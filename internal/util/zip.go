// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ZipFolder zips the contents of the specified folder and saves the zip file to the specified destination.
func ZipFolder(source, destination string) error {
	// Create the zip file
	zipfile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	// Create a new zip archive.
	zipWriter := zip.NewWriter(zipfile)
	defer zipWriter.Close()

	// Walk the source folder
	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create a local path for the file or folder in the zip archive
		localPath := strings.TrimPrefix(path, source)
		localPath = strings.TrimPrefix(localPath, string(filepath.Separator))

		// If the file is a directory, create a corresponding directory entry in the zip archive
		if info.IsDir() {
			_, err := zipWriter.Create(localPath + "/")
			return err
		}

		// Skip .zip files
		if strings.HasSuffix(localPath, ".zip") {
			return nil
		}

		// Create a file entry in the zip archive
		file, err := zipWriter.Create(localPath)
		if err != nil {
			return err
		}

		// Open the source file
		sourceFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		// Copy the contents of the file to the zip archive
		_, err = io.Copy(file, sourceFile)
		return err
	})

	return err
}

func Unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			CreateFolder(path)
		} else {
			if err := CreateFolder(filepath.Dir(path)); err != nil {
				return err
			}

			outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, rc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
