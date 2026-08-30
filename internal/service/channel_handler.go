package service

import (
	"net/http"

	"github.com/user/media-manager/internal/db"
)

// ChannelHandler is a stub for TV-guide/channel routes. The type is
// referenced from HTTPServer, but the implementation is not in this tree.
type ChannelHandler struct {
	db       *db.Database
	tmdb     *TMDbService
	stream   *StreamHandler
	discover *DiscoverEndpoints
}

func NewChannelHandler(database *db.Database, tmdb *TMDbService, stream *StreamHandler, discover *DiscoverEndpoints) *ChannelHandler {
	return &ChannelHandler{
		db:       database,
		tmdb:     tmdb,
		stream:   stream,
		discover: discover,
	}
}

func (h *ChannelHandler) RegisterRoutes(mux *http.ServeMux) {}

func (h *ChannelHandler) SeedDefaultChannels() {}