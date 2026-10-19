package servercontrol

import (
	"cablenet/streamcontrol"
	"encoding/json"
	"net/http"
	"strconv"
)

var scInstance *streamcontrol.StreamControl;

type StreamStatus string
const (
	Stopped StreamStatus = "stopped"
	Started	StreamStatus = "started"
	Failed	StreamStatus = "failed"
)

type FeedInfo struct {
	ChannelNum			int			`json:"channel_num"`
	ChannelDisplay	string	`json:"channel_display"`
	Name						string	`json:"name"`
	Network					string	`json:"network"`
	IconImg					string	`json:"icon_img"`
	Owner						string	`json:"owner"`
	Enabled					bool		`json:"enabled"`

	Status					StreamStatus	`json:"status"`
	Playlist			string	`json:"playlist"`
}

type ServerStatus struct {
	Status					string		`json:"status"`
	UptimeSeconds		int				`json:"uptime_seconds"`
	FeedsTotal			int				`json:"feeds_total"`
	FeedsStarted		int				`json:"feeds_started"`
	FeedsStopped		int				`json:"feeds_stopped"`
	FeedsFailed			int				`json:"feeds_failed"`
}

type FeedApiError struct {
	Code			string		`json:"code"`
	Message		string		`json:"message"`
}

func NewMux(sc *streamcontrol.StreamControl) *http.ServeMux {
	scInstance = sc

	mux := http.NewServeMux()

	mux.HandleFunc("GET /feeds", getAllFeeds)
	mux.HandleFunc("GET /feeds/{id}", getFeed)
	mux.HandleFunc("POST /feeds/{id}/start", requestStart)
	mux.HandleFunc("POST /feeds/{id}/stop", requestStop)
	mux.HandleFunc("POST /feeds/{id}/restart", requestRestart)

	return mux
}

func getAllFeeds(w http.ResponseWriter, r *http.Request) {


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"whatis": "all_feeds"})
}

func getFeed(w http.ResponseWriter, r *http.Request) {
	reqid, reqerr := strconv.Atoi(r.PathValue("id"))

	if reqerr != nil {
		responseData := FeedApiError{
			Code: "Error",
			Message: reqerr.Error(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(responseData)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"whatis": strconv.Itoa(reqid)})
	}
}

func requestStart(w http.ResponseWriter, r *http.Request) {


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(map[string]any{"todo": "start"})
}

func requestStop(w http.ResponseWriter, r *http.Request) {


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(map[string]any{"todo": "stop"})
}

func requestRestart(w http.ResponseWriter, r *http.Request) {


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(map[string]any{"todo": "restart"})
}
