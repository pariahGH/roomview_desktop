package util

import (
	"net"
	"fmt"
	"os"
	"io"
	"time"
	"strings"
	"github.com/youpy/go-wav"
	"github.com/gen2brain/malgo"
)

func PlaySound(stopChan chan string){
	file, err := os.Open("alert.wav")
	if err != nil {
		fmt.Println(err) 
		os.Exit(1)
	}
	stat, err := file.Stat()
	if err != nil {
		fmt.Println(err) 
		os.Exit(1)
	}
	fileSize := stat.Size()
	if err != nil {
		fmt.Println(err) 
		os.Exit(1)
	}

	defer file.Close()

	reader := wav.NewReader(file)
	f, err := reader.Format()
	if err != nil {
		fmt.Println(err) 
		os.Exit(1)
	}
	channels := uint32(f.NumChannels)
	sampleRate := f.SampleRate
	playbackLength := fileSize/int64(f.ByteRate)
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>\n", message)
	})
	if err != nil {
		fmt.Println(err) 
		os.Exit(1)
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
	onSendSamples := func(frameCount uint32, samples []byte) uint32 {
		n, _ := io.ReadFull(reader, samples)
		return uint32(n) / uint32(channels) / sampleSize
	}

	deviceCallbacks := malgo.DeviceCallbacks{
		Send: onSendSamples,
	}
	device, err := malgo.InitDevice(ctx.Context, malgo.Playback, nil, deviceConfig, deviceCallbacks)
	if err != nil {
		fmt.Println(err) 
	}
	defer device.Uninit()
	playing := false
	lastPlayed :=  time.Now()
	for{
		select {
			case cmd := <-stopChan:
				if cmd == "start"{
					err = device.Start()
					if err != nil {
						fmt.Println(err) 
					}
					playing = true
				}else if cmd == "stop"{
					err = device.Stop()
					if err != nil {
						fmt.Println(err) 
					}
					playing = false
				}
			default:
				if playing && time.Since(lastPlayed) > (time.Second * time.Duration(playbackLength)){
					reader = wav.NewReader(file)
					lastPlayed = time.Now()
				}
		}
	}
	
}

func LoadRooms(rooms []*Room,alertChan chan Update){
	for index, room := range rooms {
		go roomListener(index, room, alertChan)
	}
}

//handles updating the room and alerting to help requests
func roomListener(index int, room *Room, alertChan chan Update){
	var conn net.Conn
	for {
		if !room.Connected {
			conn = connect(room.Ip)
			room.Connected = conn != nil
		}
		for room.Connected {
			msg := make([]byte, 20)
			_, err := conn.Read(msg)
			if err != nil {
				fmt.Println(err) 
				room.Connected = false
				continue
			}
			fmt.Println("Received from "+room.Room+":", msg)
			if strings.Contains(string(msg), "need help"){
				alertChan <- Update{room.Room}
			}
		}
	}
	conn.Close()
}

func connect(ip string) (conn net.Conn){
	var bytes = []byte{0x01,0x00,0x0b,0x0a,0xa6,0xca,0x37,0x00,0x72,0x40,0x00,0x00,0xf1,0x01}
	conn, err := net.Dial("tcp",ip+":41794")
	if err != nil {
		fmt.Println(err)
		time.Sleep(time.Second*time.Duration(1))
		return nil
	}
	conn.Write(bytes)
	return conn
}