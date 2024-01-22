/*
 * LURE - Linux User REpository
 * Copyright (C) 2023 Elara Musayelyan
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package dl

import (
	"bytes"
	"context"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/mholt/archiver/v4"
	"github.com/schollz/progressbar/v3"
	"lure.sh/lure/internal/shutils/handlers"
)

// FileDownloader downloads files using HTTP
type FileDownloader struct{}

// Name always returns "file"
func (FileDownloader) Name() string {
	return "file"
}

// MatchURL always returns true, as FileDownloader
// is used as a fallback if nothing else matches
func (FileDownloader) MatchURL(string) bool {
	return true
}

// Download downloads a file using HTTP. If the file is
// compressed using a supported format, it will be extracted
func (FileDownloader) Download(opts Options) (Type, string, error) {
	u, err := url.Parse(opts.URL)
	if err != nil {
		return 0, "", err
	}

	query := u.Query()

	name := query.Get("~name")
	query.Del("~name")

	archive := query.Get("~archive")
	query.Del("~archive")

	u.RawQuery = query.Encode()

	var r io.ReadCloser
	var size int64
	if u.Scheme == "local" {
		localFl, err := os.Open(filepath.Join(opts.LocalDir, u.Path))
		if err != nil {
			return 0, "", err
		}
		fi, err := localFl.Stat()
		if err != nil {
			return 0, "", err
		}
		size = fi.Size()
		if name == "" {
			name = fi.Name()
		}
		r = localFl
	} else {
		res, err := http.Get(u.String())
		if err != nil {
			return 0, "", err
		}
		size = res.ContentLength
		if name == "" {
			name = getFilename(res)
		}
		r = res.Body
	}
	defer r.Close()

	opts.PostprocDisabled = archive == "false"

	path := filepath.Join(opts.Destination, name)
	fl, err := os.Create(path)
	if err != nil {
		return 0, "", err
	}
	defer fl.Close()

	var bar io.WriteCloser
	if opts.Progress != nil {
		bar = progressbar.NewOptions64(
			size,
			progressbar.OptionSetDescription(name),
			progressbar.OptionSetWriter(opts.Progress),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(10),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionShowCount(),
			progressbar.OptionOnCompletion(func() {
				_, _ = io.WriteString(opts.Progress, "\n")
			}),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionFullWidth(),
			progressbar.OptionSetRenderBlankState(true),
		)
		defer bar.Close()
	} else {
		bar = handlers.NopRWC{}
	}

	h, err := opts.NewHash()
	if err != nil {
		return 0, "", err
	}

	var w io.Writer
	if opts.Hash != nil {
		w = io.MultiWriter(fl, h, bar)
	} else {
		w = io.MultiWriter(fl, bar)
	}

	_, err = io.Copy(w, r)
	if err != nil {
		return 0, "", err
	}
	r.Close()

	if opts.Hash != nil {
		sum := h.Sum(nil)
		if !bytes.Equal(sum, opts.Hash) {
			return 0, "", ErrChecksumMismatch
		}
	}

	if opts.PostprocDisabled {
		return TypeFile, name, nil
	}

	_, err = fl.Seek(0, io.SeekStart)
	if err != nil {
		return 0, "", err
	}

	format, ar, err := archiver.Identify(name, fl)
	if err == archiver.ErrNoMatch {
		return TypeFile, name, nil
	} else if err != nil {
		return 0, "", err
	}

	err = extractFile(ar, format, name, opts)
	if err != nil {
		return 0, "", err
	}

	err = os.Remove(path)
	return TypeDir, "", err
}

// extractFile extracts an archive or decompresses a file
func extractFile(r io.Reader, format archiver.Format, name string, opts Options) (err error) {
	fname := format.Name()

	switch format := format.(type) {
	case archiver.Extractor:
		err = format.Extract(context.Background(), r, nil, func(ctx context.Context, f archiver.File) error {
			fr, err := f.Open()
			if err != nil {
				return err
			}
			defer fr.Close()
			fi, err := f.Stat()
			if err != nil {
				return err
			}
			fm := fi.Mode()

			path := filepath.Join(opts.Destination, f.NameInArchive)

			err = os.MkdirAll(filepath.Dir(path), 0o755)
			if err != nil {
				return err
			}

			if f.IsDir() {
				err = os.Mkdir(path, 0o755)
				if err != nil {
					return err
				}
			} else {
				outFl, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fm.Perm())
				if err != nil {
					return err
				}
				defer outFl.Close()

				_, err = io.Copy(outFl, fr)
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	case archiver.Decompressor:
		rc, err := format.OpenReader(r)
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(opts.Destination, name)
		path = strings.TrimSuffix(path, fname)

		outFl, err := os.Create(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(outFl, rc)
		if err != nil {
			return err
		}
	}

	return nil
}

// getFilename attempts to parse the Content-Disposition
// HTTP response header and extract a filename. If the
// header does not exist, it will use the last element
// of the path.
func getFilename(res *http.Response) (name string) {
	_, params, err := mime.ParseMediaType(res.Header.Get("Content-Disposition"))
	if err != nil {
		return path.Base(res.Request.URL.Path)
	}
	if filename, ok := params["filename"]; ok {
		return filename
	} else {
		return path.Base(res.Request.URL.Path)
	}
}
