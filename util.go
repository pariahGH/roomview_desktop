package main

import(
	"os"
	"time"
	"fmt"
	"io"
	"net"
	"github.com/youpy/go-wav"
	"github.com/gen2brain/malgo"
)

func playSound(stopChan chan string){
	file, err := os.Open("alert.wav")
	stat, err := file.Stat()
	if err != nil {
		fmt.Println(err)
	}
	fileSize := stat.Size()
	if err != nil {
		fmt.Println(err)
	}

	defer file.Close()

	reader := wav.NewReader(file)
	f, err := reader.Format()
	if err != nil {
		fmt.Println(err)
	}
	channels := uint32(f.NumChannels)
	sampleRate := f.SampleRate
	playbackLength := fileSize/int64(f.ByteRate)
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>\n", message)
	})
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	deviceConfig := malgo.DefaultDeviceConfig()
	deviceConfig.Format = malgo.FormatS16
	deviceConfig.Channels = channels
	deviceConfig.SampleRate = sampleRate
	deviceConfig.Alsa.NoMMap = 1

	sampleSize := uint32(malgo.SampleSizeInBytes(deviceConfig.Format))
	// This is the function that's used for sending more data to the device for playback.
	onSendSamples := func(frameCount uint32, samples []byte) uint32 {
		n, _ := io.ReadFull(reader, samples)
		return uint32(n) / uint32(channels) / sampleSize
	}

	deviceCallbacks := malgo.DeviceCallbacks{
		Send: onSendSamples,
		Stop: func(){fmt.Println("done")},
	}
	device, err := malgo.InitDevice(ctx.Context, malgo.Playback, nil, deviceConfig, deviceCallbacks)
	if err != nil {
		fmt.Println(err)
	}
	defer device.Uninit()
	
	for{
		select {
			case cmd <-stopChan:
			if cmd == "start"{
				err = device.Start()
				if err != nil {
					fmt.Println(err)
				}
				time.Sleep(time.Second * time.Duration(playbackLength))
			}else{
				err = device.Stop()
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	
}

func loadRooms(rooms *[]Room,alertChan chan string){
	for index, room := range rooms {
		go roomListener(&room, index, alertChan)
	}
}

//systems send raw number of lamp hours in their status update
//handles updating the room, emits a string when an alert happens
func roomListener(room *Room, index int, alertChan chan string){
	bytes := [0x01,0x00,0x0b,0x0a,0xa6,0xca,0x37,0x00,0x72,0x40,0x00,0x00,0xf1,0x01]
	var conn;TCPConnection
	defer conn.Close()
	for{
		conn, err := net.Dial("tcp",room.Ip+":41794")
		if err != nil {
			alertChan <- Alert{index, "connection", False}
		}else{
			alert.Connected = true
			alertChan <- Alert{index, "connection", True}
			conn.Write(bytes)
			msg := make([]byte, 20)
			for {
				io.Copy(&msg, conn)
				fmt.Println(msg)
				if strings.Contains(string(msg), "need help"){
					roomChan <- Alert{index, "alert", True}
				}
				msg = make([]byte, 20)
			}
		}
		time.Sleep(time.Second * time.Duration(5))
	}
	
	
	
}

type State struct {
	Alert bool
	AlertRooms []string
	Playing bool
}

//Alert and Connected here are used when drawing - our state keeps track of 
//if we need to play sound and what rooms need to be displayed as needing assistance
type Room struct {
	Room string
	Ip string
	Alert bool
	Connected bool
}

type Alert struct{
	Index int
	Type string
	Connected bool
}