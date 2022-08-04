package main

import ( // Required Packages

	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

const LOGFILE = "log.txt"
const FETCHTIME = 1 // time in seconds to retrieve position and orientation from log file
const WRITETIME = 1 // time in seconds to simulate writing position and orientation in the log file

// TODO: Add headset IDs
// TODO: Ask 3d plan of office to test on office
// For the plan of the office: Check cloud point solution reconstruction.

type Location struct { // Structure containing all position and orientation variables
	p_x string
	p_y string
	p_z string
	o_x string
	o_y string
	o_z string
	o_w string
}

func main() {

	log.Println("Starting server")

	ch := make(chan Location) // Shared Channel (uses the struct defined before)

	light_1 := false

	go fetch_location(FETCHTIME, ch) // Retrieve position and orientation from log file
	go write_log_test(WRITETIME)     // Write position and orientation into log file (random numbers) -> Temporary

	// go start_realsense_camera() // roslaunch realsense2_camera rs_camera.launch align_depth:=true unite_imu_method:="linear_interpolation" enable_gyro:=true enable_accel:=true
	// go start_realsense_imu()    // rosrun imu_filter_madgwick imu_filter_node _use_mag:=false _publish_tf:=false _world_frame:="enu" /imu/data_raw:=/camera/imu /imu/data:=/rtabmap/imu
	// go start_mapping()          // roslaunch rtabmap_ros rtabmap.launch rtabmap_args:="--delete_db_on_start --Optimizer/GravitySigma 0.3" depth_topic:=/camera/aligned_depth_to_color/image_raw rgb_topic:=/camera/color/image_raw camera_info_topic:=/camera/color/camera_info approx_sync:=false wait_imu_to_init:=true imu_topic:=/rtabmap/imu
	// go start_localization()     // cp ~/.ros/rtabmap.db ~/Documents/THE_NAME_YOU_WANT_FOR_YOUR_MAP

	// for res := range ch {
	// 	fmt.Println(res.p_x, ",", res.p_y)
	// }

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("The IP %s accessed the server", r.RemoteAddr)
		fmt.Fprintln(w, "Welcome to Calvarina server, enjoy your stay")
	})

	http.HandleFunc("/position", func(w http.ResponseWriter, r *http.Request) {
		lastloc := <-ch // Reads from channel
		log.Printf("The IP %s tried to read the camera location", r.RemoteAddr)
		log.Printf("RealSense Location <px:%s,py:%s,pz:%s,ox:%s,oy:%s,oz:%s,ow:%s>", lastloc.p_x, lastloc.p_y, lastloc.p_z, lastloc.o_x, lastloc.o_y, lastloc.o_z, lastloc.o_w)
		fmt.Fprintln(w, "RealSense Location <px,py,pz,ox,oy,oz,ow>", lastloc.p_x, lastloc.p_y, lastloc.p_z, lastloc.o_x, lastloc.o_y, lastloc.o_z, lastloc.o_w)
		fmt.Fprintln(w, "Light Button <STATE> = ", light_1)
	})

	http.HandleFunc("/light_1", func(w http.ResponseWriter, r *http.Request) {
		if light_1 {
			light_1 = false
		} else {
			light_1 = true
		}
		log.Printf("The Node with IP %s pressed the light button!", r.RemoteAddr)
		log.Print("Light Button <STATE> = ", light_1)
		fmt.Fprintln(w, "Light Button <STATE> = ", light_1)
	})

	log.Fatal(http.ListenAndServe("0.0.0.0:8102", nil))
}

//http://10.0.2.2:8102/
//http://192.168.1.24:8102

func write_log_test(timer float32) { // Simulates writing in the log file (temporary function)
	c := time.Tick(time.Duration(timer) * time.Second)
	for _ = range c { // Periodically writes to the log file random position and orientation
		sampledata := []string{
			"---",
			"position:",
			"  x: " + fmt.Sprintf("%f", rand.Float64()),
			"  y: " + fmt.Sprintf("%f", rand.Float64()),
			"  z: " + fmt.Sprintf("%f", rand.Float64()),
			"orientation:",
			"  x: " + fmt.Sprintf("%f", rand.Float64()),
			"  y: " + fmt.Sprintf("%f", rand.Float64()),
			"  z: " + fmt.Sprintf("%f", rand.Float64()),
			"  w: " + fmt.Sprintf("%f", rand.Float64()),
		}

		file, err := os.OpenFile(LOGFILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

		if err != nil {
			log.Fatalf("failed creating file: %s", err)
		}

		datawriter := bufio.NewWriter(file)

		for _, data := range sampledata {
			_, _ = datawriter.WriteString(data + "\n")
		}

		datawriter.Flush()
		file.Close()
	}
}

func fetch_location(timer int, ch chan Location) {
	defer close(ch)
	c := time.Tick(time.Duration(timer) * time.Second)
	for _ = range c { // Periodically reads from the log file
		file, err := os.Open(LOGFILE)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		buf := make([]byte, 140) //Only reads last chunk of 140 bytes of data, not all file (effective reading)
		stat, err := os.Stat(LOGFILE)
		start := stat.Size() - 140
		_, err = file.ReadAt(buf, start)

		if err == nil {
			s := string(buf[:])
			// p_x, p_y, p_z, o_x, o_y, o_z, o_w := parse_location(s)
			loc := new(Location)
			loc.p_x, loc.p_y, loc.p_z, loc.o_x, loc.o_y, loc.o_z, loc.o_w = parse_location(s)
			// fmt.Println("p_x,p_y,p_z,o_x,o_y,o_z,o_w : ", loc.p_x, loc.p_y, loc.p_z, loc.o_x, loc.o_y, loc.o_z, loc.o_w)
			ch <- *loc // Writes to channel
		} else {
			fmt.Printf("ERROR")
		}
	}
}

func parse_location(s string) (string, string, string, string, string, string, string) {
	last_loc := strings.SplitAfter(s, "---")
	p := strings.Split(strings.ReplaceAll(last_loc[len(last_loc)-1], "\r\n", "\n"), "\n")
	p_x := strings.Split(p[2], "x: ")[1]
	p_y := strings.Split(p[3], "y: ")[1]
	p_z := strings.Split(p[4], "z: ")[1]
	o_x := strings.Split(p[6], "x: ")[1]
	o_y := strings.Split(p[7], "y: ")[1]
	o_z := strings.Split(p[8], "z: ")[1]
	o_w := strings.Split(p[9], "w: ")[1]
	return p_x, p_y, p_z, o_x, o_y, o_z, o_w
}

// func start_realsense_camera() {
//     cmd := exec.Command("roslaunch", "realsense2_camera", "rs_camera.launch", "align_depth:=true", "unite_imu_method:=\"linear_interpolation\"", "enable_gyro:=true", "enable_accel:=true")
// 	// cmd := exec.Command("ls", "-lah")
//     out, err := cmd.CombinedOutput()
//     if err != nil {
//         log.Fatalf("cmd.Run() failed with %s\n", err)
//     }
//     fmt.Printf("combined out:\n%s\n", string(out))
// }
