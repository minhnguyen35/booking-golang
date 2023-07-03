package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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