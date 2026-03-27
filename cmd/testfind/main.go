package main

import (
	"encoding/json"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/location"
	invoiceSvc "sighthub-backend/internal/services/invoice_service"
)

func main() {
	dsn := "host=172.16.6.4 port=5432 user=icore password=Ri250Ca100w dbname=eyesync_db sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("DB ERROR:", err)
		return
	}
	svc := invoiceSvc.New(db)

	el := &invoiceSvc.EmpLocation{
		Employee: &employees.Employee{IDEmployee: 36},
		Location: &location.Location{IDLocation: 1},
	}
	
	result, err := svc.GetLocalTransfers(el, 1, 5, "", "")
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(b))
}
