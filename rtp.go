package main

import (
	"log"
	"net"

	"github.com/pion/webrtc/v3"
)

func listenRtp(ip string, port int, mediaType string, track *webrtc.TrackLocalStaticRTP) {
	client, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	})

	if err != nil {
		log.Panicln(err)
	}

	defer client.Close()

	log.Printf("Listening for UDP RTP packets (%s) on %s:%d", mediaType, ip, port)

	buf := make([]byte, 1500)

	for {
		n, _, err := client.ReadFrom(buf)

		if err != nil {
			log.Println(err)
			continue
		}

		track.Write(buf[:n])
	}
}
