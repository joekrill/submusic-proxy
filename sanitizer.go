package sanitizer

import (
	"encoding/json"
	"fmt"
)

// TODO: sanitize `getPodcasts` response?

type PlaylistsResponse struct {
	SubsonicResponse struct {
		Status        string `json:"status"`
		Version       string `json:"version"`
		Type          string `json:"type"`
		ServerVersion string `json:"serverVersion"`
		OpenSubsonic  bool   `json:"openSubsonic"`
		Playlists     struct {
			Playlist []struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				SongCount int    `json:"songCount"`
			} `json:"playlist"`
		} `json:"playlists"`
	} `json:"subsonic-response"`
}

type PlaylistResponse struct {
	SubsonicResponse struct {
		Status        string `json:"status"`
		Version       string `json:"version"`
		Type          string `json:"type"`
		ServerVersion string `json:"serverVersion"`
		OpenSubsonic  bool   `json:"openSubsonic"`
		Playlist      struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			SongCount int    `json:"songCount"`
			Entry     []struct {
				ID          string `json:"id"`
				Title       string `json:"title"`
				Artist      string `json:"artist"`
				ContentType string `json:"contentType"`
				CoverArt    string `json:"coverArt"`
				Duration    int    `json:"duration"`
			} `json:"entry"`
		} `json:"playlist"`
	} `json:"subsonic-response"`
}

func CanSanitize(path string) bool {
	return path == "/rest/getPlaylists" || path == "/rest/getPlaylist"
}

// Unmarshals a JSON response based on the endpoint
func unmarshalResponse(endpoint string, body []byte) (any, error) {
	switch endpoint {
	case "/rest/getPlaylists":
		var data PlaylistsResponse
		err := json.Unmarshal(body, &data)
		return data, err
	case "/rest/getPlaylist":
		var data PlaylistResponse
		err := json.Unmarshal(body, &data)
		return data, err
	default:
		return nil, fmt.Errorf("unknown endpoint: %s", endpoint)
	}
}

// Remarshals the response body, which removes any json properties we haven't
// specified on our structs.
func sanitizeResponse(endpoint string, body []byte) ([]byte, error) {
	data, err := unmarshalResponse(endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s response: %w", endpoint, err)
	}

	updatedBody, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to remarshal reduced %s response: %w", endpoint, err)
	}

	return updatedBody, nil

}
