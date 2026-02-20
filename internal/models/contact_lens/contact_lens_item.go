// internal/models/contact_lens/contact_lens_item.go
package contactlens

import (
	"fmt"
	"sighthub-backend/internal/models/vendors"
)

// ContactLensItem ⇄ contact_lens_item
type ContactLensItem struct {
	IDContactLensItem  int      `gorm:"column:id_contact_lens_item;primaryKey"            json:"id_contact_lens_item"`
	NameContact        string   `gorm:"column:name_contact;type:varchar(255);not null"    json:"name_contact"`
	ContactWType       *string  `gorm:"column:contact_w_type"                             json:"contact_w_type,omitempty"`
	Model              *string  `gorm:"column:model;type:varchar(255)"                    json:"model,omitempty"`
	InvoiceDesc        *string  `gorm:"column:invoice_desc;type:varchar(255)"             json:"invoice_desc,omitempty"`
	Material           *string  `gorm:"column:material;type:varchar(255)"                 json:"material,omitempty"`
	WaterPer           *string  `gorm:"column:water_per;type:varchar(20)"                 json:"water_per,omitempty"`
	GasPerm            *bool    `gorm:"column:gas_perm"                                   json:"gas_perm,omitempty"`
	Cost               *float64 `gorm:"column:cost;type:numeric(10,2)"                    json:"cost,omitempty"`
	SellingPrice       *float64 `gorm:"column:selling_price;type:numeric(10,2)"           json:"selling_price,omitempty"`
	ContactType        *string  `gorm:"column:contact_type;type:varchar(255)"             json:"contact_type,omitempty"`
	Colors             *string  `gorm:"column:colors;type:varchar(255)"                   json:"colors,omitempty"`
	Duration           *string  `gorm:"column:duration;type:varchar(255)"                 json:"duration,omitempty"`
	InsVCode           *string  `gorm:"column:ins_v_code;type:varchar(50)"                json:"ins_v_code,omitempty"`
	InsVCodeClass      *string  `gorm:"column:ins_v_code_class;type:varchar(50)"          json:"ins_v_code_class,omitempty"`
	Commission         *bool    `gorm:"column:commission"                                 json:"commission,omitempty"`
	CanLookup          *bool    `gorm:"column:can_lookup"                                 json:"can_lookup,omitempty"`
	MfrNumber          *string  `gorm:"column:mfr_number;type:varchar(255)"               json:"mfr_number,omitempty"`
	Sort1              *string  `gorm:"column:sort1;type:varchar(10)"                     json:"sort1,omitempty"`
	Sort2              *string  `gorm:"column:sort2;type:varchar(10)"                     json:"sort2,omitempty"`
	Note               *string  `gorm:"column:note;type:text"                             json:"note,omitempty"`
	ImageLink          *string  `gorm:"column:image_link;type:text"                       json:"image_link,omitempty"`
	Description        *string  `gorm:"column:description;type:text"                      json:"description,omitempty"`
	BrandContactLensID *int     `gorm:"column:brand_contact_lens_id"                      json:"brand_contact_lens_id,omitempty"`
	VendorID           int      `gorm:"column:vendor_id;not null"                         json:"vendor_id"`

	// Relation (optional preloading)
	Brand *vendors.BrandContactLens `gorm:"foreignKey:BrandContactLensID;references:IDBrandContactLens" json:"-"`
}

func (ContactLensItem) TableName() string { return "contact_lens_item" }

func (c *ContactLensItem) ToMap() map[string]interface{} {
	out := map[string]interface{}{
		"id_contact_lens_item": c.IDContactLensItem,
		"item_number":          c.NameContact, // как в Python to_dict()
		"contact_w_type":       c.ContactWType,
		"model":                c.Model,
		"invoice_desc":         c.InvoiceDesc,
		"material":             c.Material,
		"water_per":            c.WaterPer,
		"gas_perm":             c.GasPerm,
		"cost":                 c.Cost,
		"selling_price":        c.SellingPrice,
		"contact_type":         c.ContactType,
		"colors":               c.Colors,
		"duration":             c.Duration,
		"ins_v_code":           c.InsVCode,
		"ins_v_code_class":     c.InsVCodeClass,
		"commission":           c.Commission,
		"can_lookup":           c.CanLookup,
		"mfr_number":           c.MfrNumber,
		"sort1":                c.Sort1,
		"sort2":                c.Sort2,
		"note":                 c.Note,
		"image_link":           c.ImageLink,
		"description":          c.Description,
	}

	// brand (как в Python: self.brand.to_dict() if self.brand else None)
	if c.Brand != nil {
		out["brand"] = c.Brand.ToMap()
	} else {
		out["brand"] = nil
	}

	return out
}

func (c *ContactLensItem) String() string {
	return fmt.Sprintf("<ContactLensItem %s>", c.NameContact)
}
