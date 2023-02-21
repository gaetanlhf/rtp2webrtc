package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/pion/webrtc/v3"
	"gopkg.in/yaml.v3"
)

var (
	version         string
	buildTime       string
	config          ConfigStruct
	videoTrack      *webrtc.TrackLocalStaticRTP
	audioTrack      *webrtc.TrackLocalStaticRTP
	iceServersArray []string
)

type IceServersStruct struct {
	IceServer string `yaml:"ice-server"`
}

type ConfigStruct struct {
	ServeIp          string             `yaml:"serve-ip"`
	ServePort        string             `yaml:"serve-port"`
	RtpHost          string             `yaml:"rtp-host"`
	EnableVideo      bool               `yaml:"enable-video"`
	EnableAudio      bool               `yaml:"enable-audio"`
	RtpVideoPort     int                `yaml:"rtp-video-port"`
	RtpAudioPort     int                `yaml:"rtp-audio-port"`
	VideoTrackName   string             `yaml:"video-track-name"`
	AudioTrackName   string             `yaml:"audio-track-name"`
	VideoCodec       string             `yaml:"video-codec"`
	AudioCodec       string             `yaml:"audio-codec"`
	IceServers       []IceServersStruct `yaml:"ice-servers"`
	ApiLocation      string             `yaml:"api-location"`
	AllowCrossOrigin bool               `yaml:"allow-cross-origin"`
}

func main() {
	log.Printf("Starting rtp2webrtc %s build on %s", version, buildTime)

	configFilePath := os.Getenv("RTP2WEBRTC_CONFIG_FILE_PATH")
	log.Printf("Loading configuration file located at %s", configFilePath)
	configFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, iceServer := range config.IceServers {
		iceServersArray = append(iceServersArray, iceServer.IceServer)
	}

	if config.EnableVideo {
		videoTrack = initWebrtcTrack("video", config.VideoCodec, config.VideoTrackName)
		go listenRtp(config.RtpHost, config.RtpVideoPort, "video", videoTrack)
	}

	if config.EnableAudio {
		audioTrack = initWebrtcTrack("audio", config.AudioCodec, config.AudioTrackName)
		go listenRtp(config.RtpHost, config.RtpAudioPort, "audio", audioTrack)
	}

	http.HandleFunc(config.ApiLocation, OnConnect)
	log.Printf("Serving API on %s:%s", config.ServeIp, config.ServePort)
	log.Panicln(http.ListenAndServe(config.ServeIp+":"+config.ServePort, nil))
}
