/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

// Package img provide image related methods

package img

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/cockroachdb/errors"
	"github.com/nfnt/resize"
)

// Compress image compress
func Compress(file io.Reader, fileFormat string) (io.Reader, error) {
	// Reader not reset default, only support read once, convert to byte
	body, err := io.ReadAll(file)
	if err != nil {
		err = errors.Wrap(err, "file transfer to byte failed")
		return nil, err
	}
	var img image.Image
	switch fileFormat {
	case ".jpeg":
		img, err = jpeg.Decode(bytes.NewReader(body))
	case ".jpg":
		img, err = jpeg.Decode(bytes.NewReader(body))
	case ".png":
		img, err = png.Decode(bytes.NewReader(body))
	case ".gif":
		img, err = gif.Decode(bytes.NewReader(body))
	default:
		return nil, errors.New("this file is not image type")
	}
	if err != nil {
		// wrong png format, try to parse jpg format
		if err.Error() == "png: invalid format: not a PNG file" {
			img, err = jpeg.Decode(bytes.NewReader(body))
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	//width := img.Bounds().Dx()
	//height := img.Bounds().Dy()
	m := resize.Resize(0, 0, img, resize.Lanczos3)
	var writer = bytes.NewBuffer(nil)
	err = jpeg.Encode(writer, m, nil)

	// compress fail, return origin file
	if len(body) < writer.Len() {
		return file, nil
	}
	// compress success, return compress file
	newFile := writer
	return newFile, err
}
