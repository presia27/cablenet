package servercontrol

import (
	"cablenet/streamcontrol"
	"encoding/json"
	"net/http"
	"strconv"
)

var scInstance *streamcontrol.StreamControl;

type FeedInfo struct {
	ChannelNum			int			`json:"channel_num"`
	ChannelDisplay	string	`json:"channel_display"`
	Name						string	`json:"name"`
	Network					string	`json:"network"`
	IconImg					string	`json:"icon_img"`
	Owner						string	`json:"owner"`
	Enabled					bool		`json:"enabled"`

	Status					streamcontrol.StreamStatus	`json:"status"`
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
	metadata := scInstance.GetAllMetadata()
	var feedInfo []FeedInfo

	for i := 0; i < len(metadata); i++ {
		streamStatus, _ := scInstance.GetStreamStatus(metadata[i].ChannelNum)

		f := FeedInfo{
			ChannelNum: metadata[i].ChannelNum,
			ChannelDisplay: metadata[i].ChannelDisplay,
			Name: metadata[i].Name,
			Network: metadata[i].Network,
			IconImg: metadata[i].IconImg,
			Owner: metadata[i].Owner,
			Enabled: metadata[i].Enabled,
			Status: streamStatus,
			Playlist: metadata[i].GetPlaylistUrl(),
		}

		feedInfo = append(feedInfo, f)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(feedInfo)
}

func getFeed(w http.ResponseWriter, r *http.Request) {
	reqid, reqerr := strconv.Atoi(r.PathValue("id"))

	// Validate request (if the API were bigger, it would probably be better to have proper middleware)
	if reqerr != nil {
		responseData := FeedApiError{
			Code: "invalid_channel_id",
			Message: reqerr.Error(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(responseData)
	} else {
		// find channel, if it exists
		metadata, err := scInstance.GetMetadataByChannel(reqid)
		if err != nil {
			responseData := FeedApiError{
				Code: "feed_not_found",
				Message: err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(responseData)

		} else {
			streamStatus, _ := scInstance.GetStreamStatus(metadata.ChannelNum)

			f := FeedInfo{
				ChannelNum: metadata.ChannelNum,
				ChannelDisplay: metadata.ChannelDisplay,
				Name: metadata.Name,
				Network: metadata.Network,
				IconImg: metadata.IconImg,
				Owner: metadata.Owner,
				Enabled: metadata.Enabled,
				Status: streamStatus,
				Playlist: metadata.GetPlaylistUrl(),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(f)
		}
	}
}

func requestStart(w http.ResponseWriter, r *http.Request) {
	reqid, reqerr := strconv.Atoi(r.PathValue("id"))
	if reqerr != nil {
		responseData := FeedApiError{
			Code: "invalid_channel_id",
			Message: "Channel ID is invalid in type or form",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(responseData)
	} else {
		errcode, errdesc := scInstance.StartChannelWithId(reqid)
		if errcode != nil || errdesc != nil {
			// The main errors are going to be the conflict.
			if errcode.Error() == "feed_not_found" {
				responseData := FeedApiError{
					Code: errcode.Error(),
					Message: errdesc.Error(),
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(responseData)
				return
			} else if errcode.Error() == "feed_already_started" {
				responseData := FeedApiError{
					Code: errcode.Error(),
					Message: errdesc.Error(),
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(responseData)
				return
			} else {
				responseData := FeedApiError{
					Code: errcode.Error(),
					Message: "An internal error occurred. Please contact the administrator",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(responseData)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func requestStop(w http.ResponseWriter, r *http.Request) {
	reqid, reqerr := strconv.Atoi(r.PathValue("id")) // get route param
	if reqerr != nil {
		responseData := FeedApiError{
			Code: "invalid_channel_id",
			Message: "Channel ID is invalide in type or form",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(responseData)
	} else {
		errcode, errdesc := scInstance.StopChannelWithId(reqid)
		if errcode != nil || errdesc != nil {
			responseData := FeedApiError{ Code: errcode.Error() }
			w.Header().Set("Content-Type", "application/json")
			
			if errcode.Error() == "feed_not_found" {
				responseData.Message = errdesc.Error()
				w.WriteHeader(http.StatusNotFound)
			} else if errcode.Error() == "feed_already_started" {
				responseData.Message = errdesc.Error()
				w.WriteHeader(http.StatusConflict)
			} else {
				responseData.Message = "An internal error occurred. Please contact the administrator"
				w.WriteHeader(http.StatusInternalServerError)
			}

			json.NewEncoder(w).Encode(responseData)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func requestRestart(w http.ResponseWriter, r *http.Request) {


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(map[string]any{"todo": "restart"})
}
