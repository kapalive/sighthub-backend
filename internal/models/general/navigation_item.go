// internal/models/general/navigation_item.go
package general

import "fmt"

type NavigationItem struct {
	IDNavigationItem  int     `gorm:"column:id_navigation_item;primaryKey"        json:"id_navigation_item"`
	NavigationGroupID *int    `gorm:"column:navigation_group_id"                  json:"navigation_group_id,omitempty"`
	Path              string  `gorm:"column:path;type:varchar(255);not null"     json:"path"`
	Label             string  `gorm:"column:label;type:varchar(100);not null"    json:"label"`
	Icon              *string `gorm:"column:icon;type:varchar(50)"               json:"icon,omitempty"`
	OnClick           *string `gorm:"column:on_click;type:varchar(100)"          json:"on_click,omitempty"`
}

func (NavigationItem) TableName() string { return "navigation_item" }

func (n *NavigationItem) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_navigation_item":  n.IDNavigationItem,
		"navigation_group_id": nil,
		"path":                n.Path,
		"label":               n.Label,
		"icon":                nil,
		"on_click":            nil,
	}
	if n.NavigationGroupID != nil {
		m["navigation_group_id"] = *n.NavigationGroupID
	}
	if n.Icon != nil {
		m["icon"] = *n.Icon
	}
	if n.OnClick != nil {
		m["on_click"] = *n.OnClick
	}
	return m
}

func (n *NavigationItem) String() string {
	return fmt.Sprintf("<NavigationItem id=%d label='%s' path='%s'>", n.IDNavigationItem, n.Label, n.Path)
}
