package tasks

type TasksResource struct {
	IDTasksResource    int64  `gorm:"column:id_tasks_resource;primaryKey;autoIncrement" json:"id_tasks_resource"`
	TaskID             int64  `gorm:"column:task_id;not null"                           json:"task_id"`
	InvoiceID          *int64 `gorm:"column:invoice_id"                                 json:"invoice_id,omitempty"`
	InvoiceItemSaleID  *int64 `gorm:"column:invoice_item_sale_id"                       json:"invoice_item_sale_id,omitempty"`
	PatientID          *int64 `gorm:"column:patient_id"                                 json:"patient_id,omitempty"`
	InventoryID        *int64 `gorm:"column:inventory_id"                               json:"inventory_id,omitempty"`
	AddPath            *string `gorm:"column:add_path;type:text"                        json:"add_path,omitempty"`
}

func (TasksResource) TableName() string { return "tasks_resource" }
