package contextbroker

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
)

const (
	contentCacheMaxBytes      = 8 << 20
	contentCacheMaxEntryBytes = 512 << 10
	contentCacheMaxEntries    = 8
)

type contentCacheEntry struct {
	result   Result
	size     int
	sequence uint64
}

func contentCacheKey(q Query) tokenHash {
	// Quoted fields preserve boundaries even when a caller includes separators.
	// Request IDs identify separate audits, not separate representations.
	encoded := fmt.Sprintf("%q", []string{q.AgentID, q.AssignmentID, q.TaskID, q.LeaseToken, q.View, q.Target, q.Cursor, q.Since, fmt.Sprint(q.MaxBytes)})
	return sha256.Sum256([]byte(encoded))
}

func copyContentResult(v Result) Result {
	v.Data = slices.Clone(v.Data)
	v.Sources = slices.Clone(v.Sources)
	return v
}

func (r *Registry) cachedContent(token string, key tokenHash) (Result, bool) {
	hash := tokenHash(sha256.Sum256([]byte(token)))
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, active := r.byHash[hash]; !active {
		return Result{}, false
	}
	entry, ok := r.cache[hash][key]
	return copyContentResult(entry.result), ok
}

func (r *Registry) forgetContent(token string, key tokenHash) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.removeContentLocked(sha256.Sum256([]byte(token)), key)
}

func (r *Registry) storeContent(token string, key tokenHash, result Result) {
	// Account for metadata and key/map overhead as well as the raw body.
	size := len(result.Data) + len(result.ContentDigest) + 256
	for _, source := range result.Sources {
		size += len(source.Kind) + len(source.ID) + len(source.Part) + len(source.Version) + len(source.UpdatedAt) + 128
	}
	hash := tokenHash(sha256.Sum256([]byte(token)))
	r.mu.Lock()
	defer r.mu.Unlock()
	r.removeContentLocked(hash, key)
	if _, active := r.byHash[hash]; !active || size > contentCacheMaxEntryBytes {
		return
	}
	if r.cache == nil {
		r.cache = make(map[tokenHash]map[tokenHash]contentCacheEntry)
	}
	for len(r.cache[hash]) >= contentCacheMaxEntries {
		r.evictContentLocked(&hash)
	}
	for r.cacheBytes+size > contentCacheMaxBytes {
		r.evictContentLocked(nil)
	}
	if r.cache[hash] == nil {
		r.cache[hash] = make(map[tokenHash]contentCacheEntry)
	}
	r.cacheSequence++
	r.cache[hash][key] = contentCacheEntry{result: copyContentResult(result), size: size, sequence: r.cacheSequence}
	r.cacheBytes += size
}

func (r *Registry) removeContentLocked(hash, key tokenHash) {
	if entry, ok := r.cache[hash][key]; ok {
		r.cacheBytes -= entry.size
		delete(r.cache[hash], key)
	}
	if len(r.cache[hash]) == 0 {
		delete(r.cache, hash)
	}
}

func (r *Registry) dropContentCacheLocked(hash tokenHash) {
	for key := range r.cache[hash] {
		r.removeContentLocked(hash, key)
	}
}

func (r *Registry) evictContentLocked(only *tokenHash) {
	var oldestHash, oldestKey tokenHash
	var oldest uint64
	for hash, entries := range r.cache {
		if only != nil && hash != *only {
			continue
		}
		for key, entry := range entries {
			if oldest == 0 || entry.sequence < oldest {
				oldestHash, oldestKey, oldest = hash, key, entry.sequence
			}
		}
	}
	if oldest != 0 {
		r.removeContentLocked(oldestHash, oldestKey)
	}
}

func validContentDigest(value string) bool {
	if len(value) != len("sha256:")+sha256.Size*2 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	for _, ch := range value[len("sha256:"):] {
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
			return false
		}
	}
	return true
}

func sameContentMetadata(a, b Result) bool {
	return a.SnapshotID == b.SnapshotID && a.CapturedAt.Equal(b.CapturedAt) && a.Truncated == b.Truncated && a.Live == b.Live && slices.Equal(a.Sources, b.Sources)
}

func (b *Broker) fetchContent(ctx context.Context, token string, q Query) (Result, error) {
	key := contentCacheKey(q)
	cached, found := b.Registry.cachedContent(token, key)
	q.CacheEnabled = q.View == viewContent
	if q.CacheEnabled && found {
		q.KnownDigest = cached.ContentDigest
	}
	result, err := fetchContent(ctx, b.Fetcher, q)
	if err != nil {
		b.Registry.forgetContent(token, key)
		return Result{}, err
	}
	if result.ContentDigest == "" && !result.NotModified {
		b.Registry.forgetContent(token, key)
		return result, nil
	}
	invalid := !q.CacheEnabled || result.Live || result.CapturedAt.IsZero() || !validContentDigest(result.ContentDigest)
	if result.NotModified {
		body := bytes.TrimSpace(result.Data)
		invalid = invalid || !found || result.ContentDigest != q.KnownDigest || !sameContentMetadata(result, cached) || !bytes.Equal(body, []byte("null"))
		if !invalid {
			result.Data = slices.Clone(cached.Data)
			result.NotModified = false
		}
	} else {
		invalid = invalid || !json.Valid(result.Data) || bytes.Equal(bytes.TrimSpace(result.Data), []byte("null"))
	}
	if invalid {
		b.Registry.forgetContent(token, key)
		return Result{}, errors.New("invalid assignment content cache response")
	}
	b.Registry.storeContent(token, key, result)
	return result, nil
}
