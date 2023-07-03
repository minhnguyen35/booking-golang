package dbrepo

import (
	"context"
	"time"

	"github.com/minhnguyen/internal/models"
)

func (m *pgDBRepo) InsertReservation(r models.Reservation) (int, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	var newID int

	stmt := `
		insert into reservation (first_name, last_name, email, phone, start_date, end_date, room_id, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6,$7, $8, $9)
	`

	err := m.DB.QueryRowContext(ctx, 
		stmt,
		r.FirstName, r.LastName, r.Email, r.Phone, r.StartDate, r.EndDate, r.RoomID, time.Now(), time.Now()).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (m *pgDBRepo) InsertRoomRestriction(r models.RoomRestriction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()


	stmt := `
		insert into reservation (first_name, last_name, email, phone, start_date, end_date, room_id, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6,$7, $8, $9)
	`

	var newID int

	err := m.DB.QueryRowContext(ctx, stmt,
		r.StartDate, r.EndDate, r.RoomID, r.ReservationID, time.Now(), time.Now(), r.RestrictionID).Scan(&newID)
	if err != nil{
		return 0, err
	}

	return newID, nil 
}

func (m *pgDBRepo) SearchAvailability(start, end time.Time, roomId int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	query := `
		select  count(id)
		from room_restriction
		where $1 < end_date and $2 > start_date and room_id = $3;
	`

	var numRows int
	row := m.DB.QueryRowContext(ctx, query, start, end, roomId)
	err := row.Scan(&numRows)

	if err != nil {
		return false, err
	}
	if numRows == 0 {
		return true, nil
	}
	return false, nil
}

func (m *pgDBRepo) SearchAvailabilityAllRooms(start, end time.Time)([]models.Room, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	query := `
		select r.id, r.room_name
		from room r
		where r.id not in 
			(select room_id from room_restriction rr where $1 < rr.end_date and $2 > rr.start_date)
	`
	var rooms []models.Room

	rows, err := m.DB.QueryContext(ctx, query, start, end)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var room models.Room

		err = rows.Scan(
			&room.ID,
			&room.RoomName)
		
		if err != nil {
			return nil, err
		}

		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}

func (m *pgDBRepo) GetRoomById(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	var room models.Room

	query := `
		select id, room_name
		from room
		where id = $1
	`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&room.ID,
		&room.RoomName,
	)

	if err != nil {
		return room, err
	}

	return room, nil
}