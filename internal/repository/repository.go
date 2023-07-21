package repository

import (
	"time"

	"github.com/minhnguyen/internal/models"
)

type DatabaseRepo interface {
	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction(r models.RoomRestriction) (int, error)
	SearchAvailability(start, end time.Time, roomId int) (bool, error)
	SearchAvailabilityAllRooms(start, end time.Time)([]models.Room, error) 
	GetRoomById(id int) (models.Room, error)

	Authenticate(email, testPassword string) (int, string, error)
	UpdateUser(u models.User) (error)
	GetUserById(id int) (models.User, error)

	AllReservations() ([]models.Reservation, error)
	NewReservations() ([]models.Reservation, error)
	GetReservationById(id int) (models.Reservation, error)
	UpdateReservation(u models.Reservation) (error)
	DeleteReservation(id int) error
	UpdateProcessedForReservation(id, processed int) error
	
	AllRooms() ([]models.Room, error)
	GetRestrictionsForRoomByDate(roomId int, start, end time.Time) ([]models.RoomRestriction, error)

	InsertBlockForRoom(id int, startDate time.Time) error 
	DeleteBlockForRoom(id int) error
}