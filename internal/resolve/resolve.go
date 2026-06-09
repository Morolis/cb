package resolve

import (
	"fmt"

	"github.com/Morolis/cb/internal/api"
	"github.com/Morolis/cb/internal/models"
	"github.com/Morolis/cb/internal/storage"
)

type Source string

const (
	SourceLocal  Source = "local"
	SourceRemote Source = "remote"
)

type Resolver struct {
	local  *storage.LocalDB
	remote *api.Client
}

func NewResolver(local *storage.LocalDB, remote *api.Client) *Resolver {
	return &Resolver{local: local, remote: remote}
}

func (r *Resolver) GetByID(id string) (*models.Snippet, Source, error) {
	if models.IsLocalID(id) {
		// Try exact match first
		s, err := r.local.GetCached(id)
		if err == nil {
			return s, SourceLocal, nil
		}
		// Try prefix match in local DB
		s, err = r.local.GetByPrefix(id)
		if err == nil {
			return s, SourceLocal, nil
		}
		return nil, SourceLocal, fmt.Errorf("local snippet not found: %s", id)
	}

	// Short ID (< 36 chars, full UUID is 36) → try prefix search
	if len(id) < 36 {
		// Try remote prefix search first
		s, err := r.remote.GetSnippetByPrefix(id)
		if err == nil {
			r.local.CacheSnippet(s)
			return s, SourceRemote, nil
		}
		// Try local prefix
		s, err = r.local.GetByPrefix(id)
		if err == nil {
			return s, SourceLocal, nil
		}
		return nil, SourceRemote, fmt.Errorf("snippet not found: %s", id)
	}

	// Full ID → exact match
	s, err := r.remote.GetSnippet(id)
	if err == nil {
		r.local.CacheSnippet(s)
		return s, SourceRemote, nil
	}

	s, err = r.local.GetCached(id)
	if err == nil {
		return s, SourceLocal, nil
	}

	return nil, SourceRemote, fmt.Errorf("snippet not found: %s", id)
}

func (r *Resolver) GetByAlias(alias string) (*models.Snippet, Source, error) {
	// Try local first
	s, err := r.local.GetCachedByAlias(alias)
	if err == nil {
		return s, SourceLocal, nil
	}

	// Try remote
	s, err = r.remote.GetSnippetByAlias(alias)
	if err == nil {
		r.local.CacheSnippet(s)
		return s, SourceRemote, nil
	}

	return nil, SourceLocal, fmt.Errorf("snippet not found: %s", alias)
}

func (r *Resolver) GetLatest() (*models.Snippet, Source, error) {
	var localSnippet *models.Snippet
	var remoteSnippet *models.Snippet

	// Get latest local
	locals, _ := r.local.ListFiltered(1, 0, "", "")
	if len(locals) > 0 {
		localSnippet, _ = r.local.GetCached(locals[0].ID)
	}

	// Get latest remote
	remotes, err := r.remote.ListSnippets(1, 0)
	if err == nil && len(remotes) > 0 {
		remoteSnippet, _ = r.remote.GetSnippet(remotes[0].ID)
	}

	// Compare timestamps
	if localSnippet != nil && remoteSnippet != nil {
		if localSnippet.CreatedAt.After(remoteSnippet.CreatedAt) {
			return localSnippet, SourceLocal, nil
		}
		return remoteSnippet, SourceRemote, nil
	}
	if remoteSnippet != nil {
		return remoteSnippet, SourceRemote, nil
	}
	if localSnippet != nil {
		return localSnippet, SourceLocal, nil
	}

	return nil, SourceLocal, fmt.Errorf("no snippets found")
}

func (r *Resolver) Delete(id string) error {
	// Short ID (including short local IDs like loc_2ff8) → resolve to full ID
	if len(id) < 36 {
		s, source, err := r.GetByID(id)
		if err != nil {
			return err
		}
		if source == SourceLocal {
			return r.local.DeleteCached(s.ID)
		}
		// Remote: delete from server and local cache
		if err := r.remote.DeleteSnippet(s.ID); err != nil {
			return err
		}
		r.local.DeleteCached(s.ID)
		return nil
	}

	// Full ID
	if models.IsLocalID(id) {
		return r.local.DeleteCached(id)
	}

	// Full remote ID: delete from both
	if _, err := r.local.GetCached(id); err == nil {
		r.local.DeleteCached(id)
	}
	return r.remote.DeleteSnippet(id)
}

func (r *Resolver) DeleteByAlias(alias string) error {
	// Try local first
	err := r.local.DeleteByAlias(alias)
	if err == nil {
		return nil
	}

	// Try remote
	s, err := r.remote.GetSnippetByAlias(alias)
	if err != nil {
		return fmt.Errorf("snippet not found: %s", alias)
	}
	return r.remote.DeleteSnippet(s.ID)
}

type listEntry struct {
	preview models.SnippetPreview
	source  Source
}

func (r *Resolver) ListMerged(limit int, sourceFilter string) ([]models.SnippetPreview, []Source, error) {
	seen := make(map[string]bool)
	var entries []listEntry

	if sourceFilter != "remote" {
		locals, err := r.local.ListFiltered(limit, 0, "", "")
		if err == nil {
			for _, p := range locals {
				entries = append(entries, listEntry{preview: p, source: SourceLocal})
				seen[p.ID] = true
			}
		}
	}

	if sourceFilter != "local" {
		remotes, err := r.remote.ListSnippets(limit, 0)
		if err == nil {
			for _, p := range remotes {
				if seen[p.ID] {
					// Already listed as local cached, upgrade to remote
					for i, e := range entries {
						if e.preview.ID == p.ID {
							entries[i].source = SourceRemote
							break
						}
					}
					continue
				}
				entries = append(entries, listEntry{preview: p, source: SourceRemote})
				seen[p.ID] = true
			}
		}
	}

	// Sort by CreatedAt descending
	for i := 1; i < len(entries); i++ {
		for j := i; j > 0; j-- {
			if entries[j].preview.CreatedAt.After(entries[j-1].preview.CreatedAt) {
				entries[j], entries[j-1] = entries[j-1], entries[j]
			}
		}
	}

	if len(entries) > limit {
		entries = entries[:limit]
	}

	previews := make([]models.SnippetPreview, len(entries))
	sources := make([]Source, len(entries))
	for i, e := range entries {
		previews[i] = e.preview
		sources[i] = e.source
	}

	return previews, sources, nil
}

func (r *Resolver) ListRemote(limit, offset int) ([]models.SnippetPreview, error) {
	return r.remote.ListSnippets(limit, offset)
}

func (r *Resolver) ListLocal(limit, offset int, category, tag string) ([]models.SnippetPreview, error) {
	return r.local.ListFiltered(limit, offset, category, tag)
}

func (r *Resolver) CacheRemote(s *models.Snippet) {
	r.local.CacheSnippet(s)
}

func (r *Resolver) DeleteCached(id string) error {
	return r.local.DeleteCached(id)
}
