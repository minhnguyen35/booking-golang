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
}