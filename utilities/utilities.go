package utilities

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
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
		fmt.Println(err)
	}

	defer jsonFile.Close()

	jsonByteArr, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Println("Error reading file: ", err)
		return
	}

	err = json.Unmarshal(jsonByteArr, d)
	if err != nil {
		fmt.Println("Error unmarshalling JSON: ", err)
 		return
	}

	fmt.Println("Loaded ", f)
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

func BuildFFstring(s StreamList, f FFoptions) (string, error) {
	// Check that required fields are filled in
	if s.FeedUrl == "" {
		return "", errors.New("Input path or URL is missing. Unable to build FFmpeg command string.")
	}

	var sb strings.Builder

	// Determine overwrite
	if s.Overwrite {
		sb.WriteString("-y ")
	} else {
		sb.WriteString("-n ")
	}

	// Write input feed location
	sb.WriteString("-i ")
	sb.WriteString(s.FeedUrl)

	// >>Handle custom ffmpeg options per stream at a later date<<

	// Video codec info
	sb.WriteString(" -c:v ")
	sb.WriteString(f.Vcodec)
	sb.WriteString(" -profile:v ")
	sb.WriteString(f.Vprofile)
	sb.WriteString(" -preset ")
	sb.WriteString(f.Preset)

	sb.WriteString(" -b:v ")
	var vbitrate int
	if s.Vbitrate > 0 {
		vbitrate = s.Vbitrate
	} else {
		vbitrate = f.DefaultVBitrate
	}
	sb.WriteString(strconv.Itoa(vbitrate))
	sb.WriteString("k")

	sb.WriteString(" -pix_fmt ")
	sb.WriteString(f.PixFmt)
	sb.WriteString(" -colorspace ")
	sb.WriteString(f.ColorFmt)
	sb.WriteString(" -color_primaries ")
	sb.WriteString(f.ColorFmt)
	sb.WriteString(" -color_trc ")
	sb.WriteString(f.ColorFmt)
	sb.WriteString( " -color_range" )
	sb.WriteString(f.ColorRange)

	sb.WriteString(" -c:a ")
	sb.WriteString(f.Acodec)
	sb.WriteString(" -b:a ")
	var abitrate int
	if s.Abitrate > 0 {
		abitrate = s.Abitrate
	} else {
		abitrate = f.DefaultABitrate
	}
	sb.WriteString(strconv.Itoa(abitrate))
	sb.WriteString("k")

	// Output (fix for proper output later)
	fmt.Fprintf(&sb, " stream-ch%d-%s.mp4", s.ChannelNum, s.Name)

	return sb.String(), nil
}
