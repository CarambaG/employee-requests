package domain

import "time"

type Department struct {
	ID   int64
	Name string
}

type Position struct {
	ID   int64
	Name string
}

type Employee struct {
	ID         int64
	FullName   string
	Department Department
	Position   Position
}

type Request struct {
	Number      int64
	CreatedAt   time.Time
	Author      Employee
	Assignee    Employee
	Description string
	DueAt       time.Time
	Status      RequestStatus
	UpdatedAt   time.Time
}
