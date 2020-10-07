package config

import "golang.org/x/xerrors"

var (
	// SessionToken is the file containing the session token.
	SessionToken File = "session"
)

// File is a thin wrapper around os.File for conveniently interacting
// with the FS.
type File string

// Write writes to the file with 0600 perms.
func (f File) Write(s string) error {
	return write(string(f), 0600, []byte(s))
}

// Read reads the file, returning the string representation.
func (f File) Read() (string, error) {
	b, err := read(string(f))
	return string(b), err
}

// Delete deletes the file.
func (f File) Delete() error {
	return rm(string(f))
}

// ReadFiles reads and returns the contents of all the provided
// Files. It should not be used with large files.
// TODO: maybe we just make a JSON config at this point...
func ReadFiles(files ...File) (map[File]string, error) {
	m := make(map[File]string, len(files))
	for _, file := range files {
		val, err := file.Read()
		if err != nil {
			return nil, xerrors.Errorf("read %s", file)
		}
		m[file] = val
	}

	return m, nil
}
