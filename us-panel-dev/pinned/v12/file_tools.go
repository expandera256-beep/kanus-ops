package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Root-confined file tools for the agent/editor paths. Every path is
// resolved relative to one workspace root; "..", absolute paths outside the
// root and any symlink component are refused. Writes are atomic
// (temp file + fsync + rename + dir fsync) and size-limited.

const (
	fileToolMaxRead  = 8 << 20
	fileToolMaxWrite = maxBackupBytes // restores of backed-up files must still fit
)

var errFileToolEscape = errors.New("yol çalışma klasörünün dışında")

type fileTools struct {
	Root     string
	MaxRead  int64
	MaxWrite int64
	// root is pinned once at construction (unix: O_NOFOLLOW dir fd); all
	// Write/Delete/Read go through it and never re-resolve Root.
	root *rootHandle
}

func newFileTools(root string) (*fileTools, error) {
	if strings.TrimSpace(root) == "" || refusedFolder(root) {
		return nil, errors.New("geçerli proje klasörü yok")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(abs)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("geçersiz çalışma klasörü")
	}
	h, err := pinRoot(abs)
	if err != nil {
		return nil, errors.New("geçersiz çalışma klasörü")
	}
	return &fileTools{Root: abs, MaxRead: fileToolMaxRead, MaxWrite: fileToolMaxWrite, root: h}, nil
}

// resolve maps rel to a full path inside Root or returns errFileToolEscape.
func (ft *fileTools) resolve(rel string) (string, error) {
	rel = strings.TrimSpace(rel)
	if rel == "" || strings.ContainsRune(rel, 0) {
		return "", errFileToolEscape
	}
	if strings.HasPrefix(rel, `\\`) || strings.HasPrefix(rel, "//") {
		return "", errFileToolEscape // \\?\, \\.\ and UNC prefixes
	}
	for _, part := range strings.Split(strings.ReplaceAll(rel, `\`, "/"), "/") {
		if part == ".." {
			return "", errFileToolEscape
		}
	}
	full := rel
	if !filepath.IsAbs(full) {
		full = filepath.Join(ft.Root, filepath.FromSlash(rel))
	}
	full = filepath.Clean(full)
	if full == ft.Root || !insideRoot(ft.Root, full) {
		return "", errFileToolEscape
	}
	relInside, err := filepath.Rel(ft.Root, full)
	if err != nil || !windowsSafeRel(relInside) {
		return "", errFileToolEscape
	}
	if err := noSymlinkPath(ft.Root, full); err != nil {
		return "", errFileToolEscape
	}
	// Defence in depth: the deepest existing ancestor must resolve inside.
	probe := full
	for {
		if _, err := os.Lstat(probe); err == nil {
			break
		}
		next := filepath.Dir(probe)
		if next == probe {
			break
		}
		probe = next
	}
	realRoot, err := filepath.EvalSymlinks(ft.Root)
	if err != nil {
		return "", errFileToolEscape
	}
	realProbe, err := filepath.EvalSymlinks(probe)
	if err != nil || (realProbe != realRoot && !insideRoot(realRoot, realProbe)) {
		return "", errFileToolEscape
	}
	return full, nil
}

func (ft *fileTools) Read(rel string) ([]byte, error) {
	full, err := ft.resolve(rel)
	if err != nil {
		return nil, err
	}
	rel2, err := filepath.Rel(ft.Root, full)
	if err != nil {
		return nil, errFileToolEscape
	}
	f, err := openBeneathAt(ft.root, ft.Root, rel2)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() > ft.MaxRead {
		return nil, errors.New("dosya okuma sınırını aşıyor")
	}
	body, err := io.ReadAll(io.LimitReader(f, ft.MaxRead+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > ft.MaxRead {
		return nil, errors.New("dosya okuma sınırını aşıyor")
	}
	return body, nil
}

func (ft *fileTools) Write(rel string, data []byte) error {
	if int64(len(data)) > ft.MaxWrite {
		return errors.New("yazma boyut sınırı aşıldı")
	}
	full, err := ft.resolve(rel)
	if err != nil {
		return err
	}
	if info, err := os.Lstat(full); err == nil && (!info.Mode().IsRegular() || fileLinkCount(info) > 1) {
		return errors.New("normal dosya değil veya hard link")
	}
	rel2, err := filepath.Rel(ft.Root, full)
	if err != nil || rel2 == "." || strings.HasPrefix(rel2, "..") {
		return errFileToolEscape
	}
	return writeBeneathAt(ft.root, ft.Root, rel2, data)
}

func (ft *fileTools) Delete(rel string) error {
	full, err := ft.resolve(rel)
	if err != nil {
		return err
	}
	rel2, err := filepath.Rel(ft.Root, full)
	if err != nil {
		return errFileToolEscape
	}
	return deleteBeneathAt(ft.root, rel2)
}

// atomicWriteFile writes via a synced temp file in the same directory and an
// atomic rename. Readers see either the old or the new content, never a
// partial file. writeTestHook lets tests simulate a crash mid-write.
var atomicWriteTestHook func(tmp string) error

func atomicWriteFile(root, full string, data []byte) error {
	if int64(len(data)) > fileToolMaxWrite {
		return errors.New("yazma boyut sınırı aşıldı")
	}
	rel, err := filepath.Rel(root, full)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
		return errFileToolEscape
	}
	return writeBeneath(root, rel, data)
}

func syncDir(dir string) {
	if d, err := os.Open(dir); err == nil {
		_ = d.Sync()
		_ = d.Close()
	}
}

// windowsSafeRel refuses path forms that are dangerous or ambiguous on
// Windows regardless of the host OS: ':' (ADS "a.txt:stream", drive-relative
// "C:x"), reserved device names (CON, PRN, AUX, NUL, COM1-9, LPT1-9, with any
// extension) and components ending in '.' or ' '.
func windowsSafeRel(rel string) bool {
	for _, part := range strings.Split(strings.ReplaceAll(rel, `\`, "/"), "/") {
		if part == "" || part == "." {
			continue
		}
		if strings.ContainsAny(part, ":<>\"|?*") || strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") {
			return false
		}
		base := strings.ToUpper(part)
		if i := strings.IndexByte(base, '.'); i >= 0 {
			base = base[:i]
		}
		switch base {
		case "CON", "PRN", "AUX", "NUL", "CONIN$", "CONOUT$":
			return false
		}
		if len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '1' && base[3] <= '9' {
			return false
		}
	}
	return true
}
