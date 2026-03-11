package tasks

type TaskResourceRelation struct {
	IDTaskResourceRelation int64 `gorm:"column:id_task_resource_relation;primaryKey;autoIncrement" json:"id_task_resource_relation"`
	TaskID                 int64 `gorm:"column:task_id;not null"                                   json:"task_id"`
	TasksResourceID        int64 `gorm:"column:tasks_resource_id;not null"                         json:"tasks_resource_id"`
}

func (TaskResourceRelation) TableName() string { return "task_resource_relation" }
