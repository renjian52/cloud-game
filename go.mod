module github.com/giongto35/cloud-game

go 1.12

require (
        cloud.google.com/go v0.43.0
        github.com/disintegration/imaging v1.6.2
        github.com/gen2brain/x264-go v0.0.0-20200517120223-c08131f6fc8a
        github.com/go-gl/gl v0.0.0-20190320180904-bf2b1f2f34d7
        github.com/gofrs/uuid v3.2.0+incompatible
        github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
        github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
        github.com/gorilla/mux v1.7.3
        github.com/gorilla/websocket v1.4.0
        github.com/hajimehoshi/ebiten v1.11.3
        github.com/pion/webrtc/v2 v2.2.0
        github.com/prometheus/client_golang v1.1.0
        github.com/spf13/pflag v1.0.3
        github.com/veandco/go-sdl2 v0.4.4
        golang.org/x/crypto v0.0.0-20200128174031-69ecbb4d6d5d
        golang.org/x/image v0.0.0-20200119044424-58c23975cae1
        gopkg.in/hraban/opus.v2 v2.0.0-20180426093920-0f2e0b4fc6cd
)

replace github.com/hajimehoshi/ebiten => github.com/renjian52/ebiten v1.12.0-alpha.2.0.20200624163456-50a1cc10adff