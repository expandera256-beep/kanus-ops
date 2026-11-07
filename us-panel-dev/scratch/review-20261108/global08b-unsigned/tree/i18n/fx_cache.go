package i18n

import "errors"

// ReadRaw / WriteRaw give other small panel files (the FX cache) the same
// rules as the language preference: pinned root handle, no symlink or
// reparse-point traversal, regular file only, at most maxPrefSize bytes,
// temp file + atomic rename. A memory-only root (unavailable config dir)
// returns ErrNotPersisted; callers then simply do not cache.
func (p *PrefStore) ReadRaw() ([]byte, error) {
	_, r := p.root()
	if r == nil {
		return nil, ErrNotPersisted
	}
	defer p.release(r)
	if r.err != nil {
		return nil, ErrNotPersisted
	}
	parts, err := p.name()
	if err != nil {
		return nil, err
	}
	return readPrefFile(r.h, parts)
}

// WriteRaw atomically replaces the file with raw (<= maxPrefSize bytes).
func (p *PrefStore) WriteRaw(raw []byte) error {
	if len(raw) == 0 || len(raw) > maxPrefSize {
		return errors.New("i18n: raw size")
	}
	_, r := p.root()
	if r == nil {
		return ErrNotPersisted
	}
	defer p.release(r)
	if r.err != nil {
		return ErrNotPersisted
	}
	parts, err := p.name()
	if err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return writePrefFile(r.h, parts, raw)
}
