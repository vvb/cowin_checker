package main

import (
	"encoding/json"
	"fmt"
	mp3 "github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// Change - pincode and daysFromToday to experiment
func main() {
	// supports only pincode based search right now
	pincodes := []string{"560076", "560068"}

	// 0  - today
	// 1 - tomorrow and so on
	daysOffset := []int{0, 1}

	for _, pincode := range pincodes {
		for _, daysFromToday := range daysOffset {
			fmt.Printf("pincode: %s, DayOffset: %d\n", pincode, daysFromToday)
			s, err := get_availability_status(pincode, time.Now().AddDate(0, 0, daysFromToday).Format("02-01-2006"))
			if err != nil {
				//fmt.Println("error:", err.Error())
				continue
			}
			check_slots(s)
		}
	}
}

// get_availability_status - Gets the availability status for the pincode and returns as a nice structure
func get_availability_status(pincode string, dateString string) (AvailabilityStatus, error) {
	s := AvailabilityStatus{}

	url := "https://cdn-api.co-vin.in/api/v2/appointment/sessions/calendarByPin?pincode=" + pincode + "&date=" + dateString
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error: ", err.Error())
		return s, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err.Error())
	}
	// parse the response

	err = json.Unmarshal(body, &s)
	if err != nil {
		fmt.Println("failed to unmarshal body:", string(body))
		return s, err
	}

	return s, nil
}

// check_slots - checks for the slots
func check_slots(s AvailabilityStatus) {
	for _, center := range s.Centers {
		for _, session := range center.Sessions {
			if session.MinAgeLimit == 18 && session.AvailableCapacity > 0 {
				fmt.Println("Center:", center.Name, "Date:", session.Date, "Availability:", session.AvailableCapacity, "Vaccine:", session.Vaccine, "Slots:", session.Slots)
				alert_me()
			}
		}
	}
}

// alert_me - plays something to alerty you.
// configurable by changing the mp3 file :)
func alert_me() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Println(dir)

	f, err := os.Open("cowin_checker/alert.mp3")
	if err != nil {
		return err
	}
	defer f.Close()

	d, err := mp3.NewDecoder(f)
	if err != nil {
		return err
	}

	c, err := oto.NewContext(d.SampleRate(), 2, 2, 8192)
	if err != nil {
		return err
	}
	defer c.Close()

	p := c.NewPlayer()
	defer p.Close()

	if _, err := io.Copy(p, d); err != nil {
		return err
	}
	return nil
}
