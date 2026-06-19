package adapter

import (
	"context"
	"sort"

	"github.com/ghpages/mobagent/backend/internal/models"
)

type Adapter interface {
	Type() models.AgentType
	Name() string
	Run(ctx context.Context, task *models.Task, emit func(line string)) error
}

type Registry struct {
	adapters map[models.AgentType]Adapter
}

func NewRegistry(items ...Adapter) *Registry {
	r := &Registry{adapters: make(map[models.AgentType]Adapter)}
	for _, a := range items {
		r.adapters[a.Type()] = a
	}
	return r
}

func (r *Registry) Get(t models.AgentType) (Adapter, bool) {
	a, ok := r.adapters[t]
	return a, ok
}

func (r *Registry) All() []Adapter {
	out := make([]Adapter, 0, len(r.adapters))
	for _, a := range r.adapters {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return string(out[i].Type()) < string(out[j].Type()) })
	return out
}

func (r *Registry) List() []models.AgentType {
	out := make([]models.AgentType, 0, len(r.adapters))
	for k := range r.adapters {
		out = append(out, k)
	}
	return out
}
