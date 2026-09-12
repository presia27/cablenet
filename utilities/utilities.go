package utilities

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type StreamList struct {
	ChannelNum			int			`json:"channel_num"`
	ChannelDisplay	string	`json:"channel_display"`
	Name						string	`json:"name"`
	Network					string	`json:"network"`
	IconImg					string	`json:"icon_img"`
	Owner						string	`json:"owner"`
	FeedUrl					string	`json:"feed_url"`
	EpgUrl					string	`json:"epg_url"`
	Enabled					bool		`json:"enabled"`
	Overwrite				bool		`json:"overwrite"`
	Vbitrate				int			`json:"vbitrate_kb"`
	Abitrate				int			`json:"abitrate_kb"`
	Deinterlace			bool		`json:"deinterlace"`
	AutostartEncode	bool		`json:"autostart_encode"`
	FFmpegOptions		any			`json:"ffmpeg_options"`
}

type FFoptions struct {
	Vcodec					string	`json:"vcodec"`
	Vprofile				string	`json:"vprofile"`
	Preset					string	`json:"preset"`
	DefaultVBitrate	int			`json:"defaultvbitratekb"`
	PixFmt					string	`json:"pixfmt"`
	ColorFmt				string	`json:"colorfmt"`
	ColorRange			string	`json:"colorrange"`
	Acodec					string	`json:"acodec"`
	DefaultABitrate	int			`json:"defaultabitratekb"`
}

// f - filename string
// d - destination struct
func LoadJsonFile(f string, d any) {
	jsonFile, err := os.Open(f)

	if err != nil {
		log.Println(err)
	}

	defer jsonFile.Close()

	jsonByteArr, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Println("Error reading file: ", err)
		return
	}

	err = json.Unmarshal(jsonByteArr, d)
	if err != nil {
		log.Println("Error unmarshalling JSON: ", err)
 		return
	}

	log.Println("Loaded ", f)
}

func GetStreamListInfo(streams []StreamList) string {
	var sb strings.Builder

	for _, s := range streams {
		fmt.Fprintf(&sb, "Stream Channel %d (%s): %s\n", s.ChannelNum, s.ChannelDisplay, s.Name)
		fmt.Fprintf(&sb, "  Network:      %s\n", s.Network)
		fmt.Fprintf(&sb, "  Icon:         %s\n", s.IconImg)
		fmt.Fprintf(&sb, "  Feed URL:     %s\n", s.FeedUrl)
		fmt.Fprintf(&sb, "  EPG URL:      %s\n", s.EpgUrl)
		fmt.Fprintf(&sb, "  Enabled:      %t\n", s.Enabled)
		fmt.Fprintf(&sb, "  Overwrite:    %t\n", s.Overwrite)
		fmt.Fprintf(&sb, "  Video Bitrate: %d kb\n", s.Vbitrate)
		fmt.Fprintf(&sb, "  Audio Bitrate: %d kb\n", s.Abitrate)
		fmt.Fprintf(&sb, "  Deinterlace:  %t\n", s.Deinterlace)
		fmt.Fprintf(&sb, "  Autostart:    %t\n", s.AutostartEncode)
		fmt.Fprintf(&sb, "  FFmpeg Opts:  %v\n", s.FFmpegOptions)
		sb.WriteString("\n")
	}

	return sb.String()
}

func BuildFFparams(s StreamList, f FFoptions) ([]string, error) {
	// Check that required fields are filled in
	if s.FeedUrl == "" {
		return nil, errors.New("Input path or URL is missing. Unable to build FFmpeg command string.")
	}

	var params []string

	// Determine overwrite
	if s.Overwrite {
		params = append(params, "-y")
	} else {
		params = append(params, "-n")
	}

	// Write input feed location
	params = append(params, "-i")
	params = append(params, s.FeedUrl)

	// >>Handle custom ffmpeg options per stream at a later date<<

	// Video codec info
	params = append(params, "-c:v")
	params = append(params, f.Vcodec)
	params = append(params, "-profile:v")
	params = append(params, f.Vprofile)
	params = append(params, "-preset")
	params = append(params, f.Preset)

	params = append(params, "-b:v")
	var vbitrate int
	if s.Vbitrate > 0 {
		vbitrate = s.Vbitrate
	} else {
		vbitrate = f.DefaultVBitrate
	}
	params = append(params, strconv.Itoa(vbitrate) + "k")

	params = append(params, "-pix_fmt")
	params = append(params, f.PixFmt)
	params = append(params, "-colorspace")
	params = append(params, f.ColorFmt)
	params = append(params, "-color_primaries")
	params = append(params, f.ColorFmt)
	params = append(params, "-color_trc")
	params = append(params, f.ColorFmt)
	params = append(params,  "-color_range" )
	params = append(params, f.ColorRange)

	params = append(params, "-c:a")
	params = append(params, f.Acodec)
	params = append(params, "-b:a")
	var abitrate int
	if s.Abitrate > 0 {
		abitrate = s.Abitrate
	} else {
		abitrate = f.DefaultABitrate
	}
	params = append(params, strconv.Itoa(abitrate) + "k")

	// Output (fix for proper output later)
	params = append(params, fmt.Sprintf("stream-ch%d-%s.mp4", s.ChannelNum, s.Name))

	return params, nil
}
