package main

import (
	"fmt"
	"os"
	"encoding/csv"
	"time"
	"runtime"
	"strconv"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"github.com/xlab/closer"
	"github.com/pariahGH/roomview_desktop/util"
	)
	
const (
	winHeight = 900
	winWidth = 300
	maxVertexBuffer  = 512 * 1024
	maxElementBuffer = 128 * 1024
)

func init() {
	runtime.LockOSThread()
}

func main(){
	if err := glfw.Init(); err != nil {
		fmt.Println(err)
	}
	glfw.WindowHint(glfw.Resizable, glfw.False)
	win, err := glfw.CreateWindow(winWidth, winHeight, "Roomview Desktop - Custom", nil, nil)
	if err != nil {
		fmt.Println(err)
	}
	win.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		closer.Fatalln("opengl: init failed:", err)
	}
	
	ctx := nk.NkPlatformInit(win, nk.PlatformInstallCallbacks)
	atlas := nk.NewFontAtlas()
	nk.NkFontStashBegin(&atlas)
	sansfont := nk.NkFontAtlasAddDefault(atlas, 14,nil)
	nk.NkFontStashEnd()
	if sansfont != nil {
		nk.NkStyleSetFont(ctx, sansfont.Handle())
	}
	exitC := make(chan struct{}, 1)
	doneC := make(chan struct{}, 1)
	closer.Bind(func() {
		close(exitC)
		<-doneC
	})
	file, err := os.Open("rooms.csv")
	if err != nil {
		fmt.Println(err)
	}
	csvReader := csv.NewReader(file)
	if err != nil {
		fmt.Println(err)
	}
	
	roomList, err := csvReader.ReadAll()
	if err != nil {
		fmt.Println(err)
	}
	rooms := make([]util.Room,0)
	for _,room := range roomList[1:] {
		rooms = append(rooms, util.Room{room[0], room[1], false, false})
	}
	alertChan := make(chan util.Update)
	soundChan := make(chan string)
	go util.PlaySound(soundChan)
	
	util.LoadRooms(rooms, alertChan)
	state := &util.State{
		Alert: false,
		AlertRooms: make([]string, 1),
		Playing: false,
	}
	fpsTicker := time.NewTicker(time.Second / 10)
	for {
		select {
		case <-exitC:
			nk.NkPlatformShutdown()
			glfw.Terminate()
			fpsTicker.Stop()
			close(doneC)
			return
		case alert := <- alertChan:
			if alert.Type == "alert"{
				state.Alert = true
				state.AlertRooms = append(state.AlertRooms, alert.Room)
			}
			if alert.Type == "connection"{
				rooms[alert.Index].Connected = alert.Connected
			}
		case <-fpsTicker.C:
			if win.ShouldClose() {
				close(exitC)
				continue
			}
			glfw.PollEvents()
			gfxMain(win, ctx, state, rooms)
			checkState(state, soundChan)
		}
	}
}
func checkState(state *util.State, soundChan chan string){
	if state.Alert && !state.Playing{
		soundChan <- "start"
		state.Playing = true
	}
	if !state.Alert && state.Playing{
		soundChan <- "stop"
		state.Playing = false
	}
}

func gfxMain(win *glfw.Window, ctx *nk.Context, state *util.State, rooms []util.Room){
	nk.NkPlatformNewFrame()

	bounds := nk.NkRect(0, 0, winWidth, winHeight)
	roomString := ""
	for _, room := range state.AlertRooms {
		roomString += room
	}
	if nk.NkBegin(ctx, "", bounds, nk.WindowBorder) > 0 {
		nk.NkLayoutRowDynamic(ctx, 30, 1)
		nk.NkLabel(ctx, "Alerts: " + roomString, nk.TextLeft)
		if nk.NkButtonLabel(ctx, "Clear Alerts") > 0 {
			state.AlertRooms = make([]string, 1)
			state.Alert = false
		}
		for _,room := range rooms{
			nk.NkLayoutRowDynamic(ctx, 30, 1)
			nk.NkLabel(ctx, "Room: "+room.Room+" Connected: "+strconv.FormatBool(room.Connected), nk.TextLeft)
		}
	}
	
	nk.NkEnd(ctx)
	bg := make([]float32, 4)
	width, height := win.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.ClearColor(bg[0], bg[1], bg[2], bg[3])
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
	win.SwapBuffers()
}

