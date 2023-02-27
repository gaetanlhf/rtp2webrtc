package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pion/webrtc/v3"
)

var (
	offer webrtc.SessionDescription
)

func initWebrtcTrack(mediaType string, codec string, trackName string) *webrtc.TrackLocalStaticRTP {
	track, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: codec}, mediaType, trackName)

	if err != nil {
		log.Panicln(err)
	}

	return track
}

func OnConnect(w http.ResponseWriter, r *http.Request) {
	if config.AllowCrossOrigin {
		header := w.Header()
		header.Add("Access-Control-Allow-Origin", "*")
		header.Add("Access-Control-Allow-Methods", "DELETE, POST, GET, OPTIONS")
		header.Add("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		header.Add("Access-Control-Allow-Private-Network", "Content-Type, Authorization, X-Requested-With")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	if r.Method != "POST" {
		log.Println("Invalid HTTP method!")
		http.Error(w, "Invalid HTTP method", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		http.Error(w, "Bad session description", http.StatusBadRequest)
		return
	}

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: iceServersArray,
			},
		},
	})

	if err != nil {
		log.Panicln(err)
	}

	if config.EnableVideo {
		_, err = peerConnection.AddTrack(videoTrack)
		if err != nil {
			log.Panicln(err)
		}
	}

	if config.EnableAudio {
		_, err = peerConnection.AddTrack(audioTrack)
		if err != nil {
			log.Panicln(err)
		}
	}

	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Printf("Connection state changed: %s", state)
		switch state {
		case webrtc.PeerConnectionStateDisconnected:
			fallthrough
		case webrtc.PeerConnectionStateFailed:
			log.Println("Explicitly closing connection")
			_ = peerConnection.Close()
		}
	})

	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		log.Panicln(err)
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	if err = peerConnection.SetLocalDescription(answer); err != nil {
		panic(err)
	}

	<-gatherComplete

	response, err := json.Marshal(peerConnection.LocalDescription())
	if err != nil {
		log.Panicln(err)
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(response); err != nil {
		log.Panicln(err)
	}
}
