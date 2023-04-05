
<h2 align="center">rtp2webrtc</h2>
<p align="center">A low latency WebRTC server in the form of a simple API for broadcasting RTP streams</p>
<p align="center">
    <a href="#about">About</a> •
    <a href="#features">Features</a> •
    <a href="#build">Build</a> •
    <a href="#configuration">Configuration</a> •
    <a href="#run">Run</a> •
    <a href="#usage">Usage</a> •
    <a href="#license">License</a>
</p>

## About

rtp2webrtc is a low latency WebRTC server in the form of a simple API for broadcasting RTP streams to any browser.

## Features

- ✅ A **single** statically compiled **binary** for each OS/architecture
- ✅ Easy **configuration** of the **IP address** and **port** of the **server**
- ✅ Easily **configure** the **RTP** **source(s)**
- ✅ Can receive a **video** and/or **audio** **RTP** stream
- ✅ **Define** **track names** without difficulty
- ✅ Effortlessly **define** your **WebRTC** stream **codecs** (must be the same as the RTP source)
- ✅ Easy **configurable** **ICE** server(s)
- ✅ **Choose** the **address** you want for the **API** **endpoint**
- ✅ **Option** to **allow** **cross origin requests**
- ✅ An **easily configurable** tool
- ✅ Can operate effortlessly as a **daemon**

## Build

First check that you have **Golang** installed on your machine.  
Then, **run**:  
```bash
make 
```
Quite simply!

## Configuration

Here is an example of a configuration (YAML): 
```yaml
# IP on which rtp2webrtc should be reachable
serve-ip: 127.0.0.1
# Port on which rtp2webrtc should be reachable
serve-port: 8080
# IP of the device sending the RTP streams
rtp-host: 127.0.0.1
# If you want to retrieve a video stream (true/false)
enable-video: true
# If you want to retrieve an audio stream (true/false)
enable-audio: true
# Port on which the video stream is to be received
rtp-video-port: 9903
# Port on which the audio stream is to be received
rtp-audio-port: 9904
# Video track name
video-track-name: default
# Video track name
audio-track-name: default
# Video codec (must be the same as the video RTP stream, either video/H264 or video/VP8)
video-codec: video/H264
# Audio codec (must be the same as the audio RTP stream, audio/opus)
audio-codec: audio/opus
# A list of ICE servers that can be used for the server
ice-servers:
  - ice-server: stun:stun.l.google.com:19302
# Customisable API endpoint
api-location: /rtp2webrtc/api/v1/offer
# To enable support for cross origin requests (for testing purposes for example)
allow-cross-origin: false
```

## Run
### Direct use

To be able to use rtp2webrtc directly, you must set the environment variable `RTP2WEBRTC_CONFIG_FILE_PATH` as the path to the configuration file.  
For example :
```
export RTP2WEBRTC_CONFIG_FILE_PATH=./config.yaml
```
Then you can run the program:
```
./rtp2webrtc
```

### As a systemd service

It is possible to easily use rtp2webrtc as a daemon with the provided systemd service.  
The systemd service provided can be adapted to your needs.

Steps to install rtp2webrtc as a systemd service:

- Create a group:

```
groupadd rtp2webrtc
```

 - Create an user:

```
useradd -r -s /sbin/nologin -g rtp2webrtc rtp2webrtc
```

- Copy the `rtp2webrtc` binary to `/usr/bin/`

```
cp rtp2webrtc /usr/bin/
```

- Create a `rtp2webrtc` folder in `/etc/` for the configuration file

```
mkdir /etc/rtp2webrtc
```

- Copy the `config.yaml` configuration file to `/etc/rtp2webrtc/`

```
cp config.yaml /etc/rtp2webrtc/
```

- Copy the `rtp2webrtc.service` systemd service file to `/etc/systemd/system/`

```
cp rtp2webrtc.service /etc/systemd/system/
```

- Start the systemd service

```
systemctl start rtp2webrtc
```

# Usage

## In a browser

Here is an example of an HTML page containing the elements needed to display WebRTC streams in a browser:

```html
<!DOCTYPE html>
<html>

<head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <title>WebRTC streaming</title>
    <meta name="description" content="">
    <meta name="viewport" content="width=device-width, initial-scale=1">
</head>

<body>
    <video id="feed"></video>

    <script>
        const defaultRtcConfig = new RTCPeerConnection({
            // ICE servers
            iceServers: [{
                urls: "stun:stun.l.google.com:19302"
            }]
        })

        function makeLivestream(videoElement, rtcConfig, endpoint, replyHandler, errorHandler) {
            let pc = new RTCPeerConnection(rtcConfig)

            pc.ontrack = function(event) {
                var el = videoElement
                el.srcObject = event.streams[0]
                el.autoplay = true
                el.controls = true
            }
            // If you want to receive video
            pc.addTransceiver("video", {
                "direction": "sendrecv"
            })
            // If you want to receive audio 
            pc.addTransceiver("audio", {
                "direction": "sendrecv"
            })
            pc.createDataChannel("noop")
            pc.createOffer()
                .then(offer => {
                    pc.setLocalDescription(offer)
                    return fetch(endpoint, {
                        method: "post",
                        headers: {
                            "Accept": "application/json, text/plain, */*",
                            "Content-Type": "application/json"
                        },
                        body: JSON.stringify(offer)
                    })
                })
                .then(res => res.json())
                .then(res => pc.setRemoteDescription(new RTCSessionDescription(res)))
                .then(replyHandler)
                .catch(err => console.log(err))
        }

        const videoElement = document.getElementById("feed")
        // Please update the URL according to your configuration
        makeLivestream(videoElement, defaultRtcConfig, "http://localhost:8080/rtp2webrtc/api/v1/offer", null, null)
    </script>
</body>

</html>
```

## Behind a reverse proxy

We recommend that the server runs behind a reverse proxy.  
Here is an example of a configuration for NGINX:

```nginx
events {}

http {
    sendfile on;
    sendfile_max_chunk 512k;
    tcp_nopush on;
    tcp_nodelay on;

    server {
        server_name streaming.domain.tld;
        listen 80;

        location / {
            root /var/www/html;
            index index.html;
        }

        location /reverse {
            rewrite /reverse/(.*) /$1  break;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection upgrade;
            proxy_redirect off;
            proxy_buffering off;
            proxy_request_buffering off;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_pass http://127.0.0.1:8080;
        }

        server_tokens off;
    }

}
```

## License

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see http://www.gnu.org/licenses/.
