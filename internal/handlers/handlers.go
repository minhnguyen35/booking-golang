package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/minhnguyen/internal/config"
	"github.com/minhnguyen/internal/driver"
	"github.com/minhnguyen/internal/forms"
	"github.com/minhnguyen/internal/helper"
	"github.com/minhnguyen/internal/models"
	"github.com/minhnguyen/internal/render"
	"github.com/minhnguyen/internal/repository"
	"github.com/minhnguyen/internal/repository/dbrepo"
)

type Repository struct {
	AppConfig *config.AppConfig
	DB        repository.DatabaseRepo
}

var Repo *Repository

func NewRepository(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		AppConfig: a,
		DB:        dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the handler for the home page
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr
	m.AppConfig.Session.Put(r.Context(), "remote_ip", remoteIP)

	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "about.page.tmpl", &models.TemplateData{})
}

func (m *Repository) GeneralRoom(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "general.page.tmpl", &models.TemplateData{})
}

func (m *Repository) SuiteRoom(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "suite.page.tmpl", &models.TemplateData{})

}

func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{})

}

func (m *Repository) SearchAvailability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{})

}

func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	layout := "2006-01-02"
	
	startDate, err := time.Parse(layout, start)
	if err != nil {
		helper.ServerError(w, err)
	}

	endDate, err := time.Parse(layout, end)
	if err != nil {
		helper.ServerError(w, err)
	}

	rooms, err := m.DB.SearchAvailabilityAllRooms(startDate, endDate)

	if err != nil {
		helper.ServerError(w, err)
		return 
	}

	for _, i := range rooms {
		m.AppConfig.InfoLog.Println("ROOM: Available Room")
		m.AppConfig.InfoLog.Println("ROOM: ", i.ID, i.RoomName)
	}

	if len(rooms) == 0 {
		m.AppConfig.Session.Put(r.Context(), "error", "No Availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate: endDate,
	}

	m.AppConfig.Session.Put(r.Context(), "reservation", res)

	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	RoomID string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate string `json:"end_date"`
}

func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	
	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")
	dateFM := "2006-01-02"
	startDate, _ := time.Parse(dateFM, sd)
	endDate, _ := time.Parse(dateFM, ed)

	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	avalableRoom, _ := m.DB.SearchAvailability(startDate, endDate, roomID)

	resp := jsonResponse{
		OK:      avalableRoom,
		Message: "",
		StartDate: sd,
		EndDate: ed,
		RoomID: strconv.Itoa(roomID),
	}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		helper.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	w.Write(out)
}

func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {

	res, ok := m.AppConfig.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helper.ServerError(w, errors.New("error Parsing Reservation"))
		return
	}
	m.AppConfig.InfoLog.Println("Reservation: roomID ", res.RoomID)

	room, err := m.DB.GetRoomById(res.RoomID)
	m.AppConfig.InfoLog.Println("Reservation: room ", room)

	if err != nil {
		helper.ServerError(w, err)
		return
	}

	res.Room.RoomName = room.RoomName

	m.AppConfig.Session.Put(r.Context(), "reservation", res)

	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed
	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "make_reservation.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
		StringMap: stringMap,
	})

}

func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.AppConfig.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		helper.ServerError(w, errors.New("error Parsing Reservation"))
		return
	}

	err := r.ParseForm()
	if err != nil {
		helper.ServerError(w, err)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName =  r.Form.Get("last_name")
	reservation.Phone =    r.Form.Get("phone")
	reservation.Email =    r.Form.Get("email")
	

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 10)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		render.Template(w, r, "make_reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newID, err := m.DB.InsertReservation(reservation)

	if err != nil {
		helper.ServerError(w, err)
		return
	}
	m.AppConfig.Session.Put(r.Context(), "reservation", reservation)

	restriction := models.RoomRestriction{
		StartDate: reservation.StartDate,
		EndDate: reservation.EndDate,
		RoomID: reservation.RoomID,
		ReservationID: newID,
		RestrictionID: 1,
	}

	_, err = m.DB.InsertRoomRestriction(restriction)

	if err != nil {
		helper.ServerError(w, err)
		return
	}

	htmlMsg := fmt.Sprintf(`
		<strong>Reservation Confirmation</strong><br>
		Dear %s, <br>
		This is confirmation from %s to %s.
	`, reservation.FirstName, 
		reservation.StartDate.Format("2006-01-02"),reservation.EndDate.Format("2006-01-02"))

	msg := models.MailData{
		To: reservation.Email,
		From: "me@here.com",
		Subject: "Reservation Confirmation",
		Content: htmlMsg,
	}

	m.AppConfig.MailChan <- msg

	m.AppConfig.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.AppConfig.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		m.AppConfig.ErrorLog.Println("Cannot get item from session")
		m.AppConfig.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.AppConfig.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation
	dateFm := "2006-01-02"
	sd := reservation.StartDate.Format(dateFm)
	ed := reservation.EndDate.Format(dateFm)
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed
	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data,
		StringMap: stringMap,
	})
}


func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomId, err := strconv.Atoi(chi.URLParam(r, "id"))
	m.AppConfig.InfoLog.Println("ChooseRoom: roomID ", roomId)
	if err != nil {
		helper.ServerError(w, err)
		return
	}

	res, ok := m.AppConfig.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helper.ServerError(w, err)
		return 
	}

	res.RoomID = roomId
	m.AppConfig.Session.Put(r.Context(), "reservation", res)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	startDate := r.URL.Query().Get("s")
	endDate := r.URL.Query().Get("e")

	dateFm := "2006-01-02"
	sd, _ := time.Parse(dateFm, startDate)
	ed, _ := time.Parse(dateFm, endDate)

	var res models.Reservation

	room, err := m.DB.GetRoomById(id)

	if err != nil {
		helper.ServerError(w, err)
		return
	}

	res.Room.RoomName = room.RoomName
	res.RoomID = id
	res.StartDate = sd
	res.EndDate = ed
	m.AppConfig.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

func (m *Repository) Login(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Data: nil,
		Form: forms.New(nil),
	})
}

func (m* Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	_ = m.AppConfig.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		m.AppConfig.InfoLog.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	if !form.Valid() {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		m.AppConfig.InfoLog.Println(err)
		m.AppConfig.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.AppConfig.Session.Put(r.Context(), "user_id", id)
	m.AppConfig.Session.Put(r.Context(), "flash", "Login Successful")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.AppConfig.Session.Destroy(r.Context())
	_ = m.AppConfig.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllReservations()
	if err != nil {
		helper.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations


	render.Template(w, r, "admin-all-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) AdminCalendarReservations(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	if r.URL.Query().Get("y") != "" {
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	next := now.AddDate(0, 1, 0)
	last := now.AddDate(0, -1, 0)
	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")

	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	strMap := make(map[string]string)
	strMap["next_month"] = nextMonth
	strMap["next_month_year"] = nextMonthYear
	strMap["this_month"] = now.Format("01")
	strMap["this_month_year"] = now.Format("2006")
	strMap["last_month"] = lastMonth
	strMap["last_month_year"] = lastMonthYear

	data := make(map[string]interface{})
	data["now"] = now

	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	
	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	rooms, err := m.DB.AllRooms()
	if err != nil {
		helper.ServerError(w, err)
		return
	}

	data["rooms"] = rooms

	for _, x := range rooms {
		reservationMap := make(map[string]int)
		blocked := make(map[string]int)

		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006-01-2")] = 0
			blocked[d.Format("2006-01-2")] = 0
		}

		restrictions, err :=m.DB.GetRestrictionsForRoomByDate(x.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helper.ServerError(w, err)
			return
		}
		for _, y := range restrictions {
			log.Println("restriction: ", y.ID)
			if y.ReservationID > 0 {
				for d := y.StartDate; d.After(y.EndDate) == false;d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format("2006-01-2")] = y.ReservationID
				}
			} else {
				blocked[y.StartDate.Format("2006-01-2")] = y.ID
			}
		}
		log.Println("update success ", x.RoomName)
		data[fmt.Sprintf("reservation_map_%d", x.ID)] = reservationMap
		data[fmt.Sprintf("blocked_%d", x.ID)] = blocked
		
		m.AppConfig.Session.Put(r.Context(), fmt.Sprintf("blocked_%d", x.ID), blocked)
	}

	render.Template(w, r, "admin-calendar-reservations.page.tmpl", &models.TemplateData{
		StringMap: strMap,
		Data: data,
		IntMap: intMap,
	})

}

func (m *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	
	if err != nil {
		helper.ServerError(w, err)
		return
	}

	src := exploded[3]
	stringMap := make(map[string]string)
	stringMap["src"] = src

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	stringMap["month"] = month
	stringMap["year"] = year

	res, err := m.DB.GetReservationById(id)
	if err != nil {
		helper.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "admin-show-reservation.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data: data,
		Form: forms.New(nil),
	})

}

func (m *Repository) AdminPostShowReservation(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helper.ServerError(w, err)
		return
	}

	src := exploded[3]
	stringMap := make(map[string]string)
	stringMap["src"] = src
	res, err := m.DB.GetReservationById(id)
	if err != nil {
		helper.ServerError(w, err)
		return
	}
	
	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	err = m.DB.UpdateReservation(res)
	if err != nil {
		helper.ServerError(w, err)
		return
	}

	month := r.Form.Get("month")
	year := r.Form.Get("year")
	
	m.AppConfig.Session.Put(r.Context(), "flash", "Changed Saved!")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservation-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservation-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
	}
	
}

func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.NewReservations()
	if err != nil {
		helper.ServerError(w, err)
		return
	}
	data := make(map[string]interface{})
	data["reservations"] = reservations


	render.Template(w, r, "admin-new-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})

}

func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	err := m.DB.UpdateProcessedForReservation(id, 1)
	if err != nil {
		m.AppConfig.InfoLog.Println(err)
	}
	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")
	m.AppConfig.Session.Put(r.Context(), "flash", "Reservatio marked as processed")
	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservation-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservation-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
	}
}

func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	err := m.DB.DeleteReservation(id)
	if err != nil {
		m.AppConfig.InfoLog.Println(err)
	}
	
	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")
	
	m.AppConfig.Session.Put(r.Context(), "flash", "Reservation deleted")
	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservation-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservation-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
	}

}

func (m *Repository) AdminPostCalendarReservations(w http.ResponseWriter, r * http.Request) {
	err := r.ParseForm()
	if err != nil {
		helper.ServerError(w, err)
		return
	}

	year, _ := strconv.Atoi(r.Form.Get("y"))
	month, _ := strconv.Atoi(r.Form.Get("m"))

	rooms, err := m.DB.AllRooms()
	if err != nil {
		helper.ServerError(w, err)
		return
	}
	form := forms.New(r.PostForm)
	//remove blocks
	for _, x := range rooms {
		curMap := m.AppConfig.Session.Get(r.Context(), fmt.Sprintf("blocked_%d", x.ID)).(map[string]int)
		for name, value := range curMap {
			if val, ok := curMap[name]; ok {
				if val > 0 {
					if !form.Has(fmt.Sprintf("remove_block_%d_%s", x.ID, name)) {
						err := m.DB.DeleteBlockForRoom(value)
						if err != nil {
							log.Println(err)
						}
					}
				}
			}
		}
	}
	//add blocks
	for name, _ := range r.PostForm {
		if strings.HasPrefix(name, "add_block") {
			exploded := strings.Split(name, "_")
			roomID, _ := strconv.Atoi(exploded[2])
			t, _ := time.Parse("2006-01-2", exploded[3])
			err := m.DB.InsertBlockForRoom(roomID, t)
			if err != nil {
				log.Println(err)
			}
			log.Println("insert block for ", roomID, " for date ", exploded[3])
		}
	}

	m.AppConfig.Session.Put(r.Context(), "flash", "Changes Saved!")

	http.Redirect(w, r, fmt.Sprintf("/admin/reservation-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
}