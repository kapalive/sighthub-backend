# Orders Dashboard & Notifications — API

## Orders Dashboard

### Locations
```
GET /api/orders/locations
```
Returns locations where current user has permission 62 (Add Item).
```json
[
  { "location_id": 1, "full_name": "Fabulous Optical", "current": true },
  { "location_id": 4, "full_name": "Bronx Vision Center", "current": false }
]
```

### Lab Status (Glasses tickets, g_or_c = "g")
```
GET /api/orders/lab-status
```

### Contact Status (Contact Lens tickets, g_or_c = "c")
```
GET /api/orders/contact-status
```

**Query params (both endpoints):**

| Param | Type | Description |
|---|---|---|
| `location_id` | int / `"all"` | Default = current location |
| `status_id` | int | Filter by ticket status ID |
| `search` | string | Search ticket number, patient, employee |
| `tray` | string | Filter by tray |
| `promised` | `today` / `tomorrow` / `overdue` | Filter by promise date |
| `complete` | `true` / `false` | Completed / not completed |
| `sort` | `ticket` / `late` / `date` / `patient` / `status` / `rep` | Sort column |
| `order` | `desc` | Sort direction (default: ASC) |
| `page` | int | Page (default: 1) |
| `per_page` | int | Per page (default: 50, max: 200) |
| `output` | `csv` | Download CSV |

**Response:**
```json
{
  "items": [
    {
      "id_lab_ticket": 78,
      "number_ticket": "T-273",
      "tray": null,
      "invoice_id": 273,
      "patient_id": 109,
      "inv_date": "2026-03-25",
      "date_promise": "2026-04-05",
      "date_complete": null,
      "late": -9,
      "status_id": 2,
      "status": "Lab Initiated",
      "employee": "Aleksandr Kapachinskikh",
      "patient": "zavulonov, manny",
      "phone": "+13475445138",
      "notified": null,
      "dashboard_note": null,
      "g_or_c": "g",
      "location_id": 1,
      "location_name": "Fabulous Optical"
    }
  ],
  "total": 27,
  "page": 1,
  "per_page": 50,
  "total_pages": 1
}
```

### Invoice Status
```
GET /api/orders/invoice-status
```
Shows only **patient**-type statuses (not internal/vendor).

| Param | Type | Description |
|---|---|---|
| `location_id` | int / `"all"` | Default = current |
| `status_id` | int | Filter by invoice status ID |
| `search` | string | Search invoice number, patient, employee |
| `sort` | `invoice` / `late` / `date` / `patient` / `status` / `rep` | Sort column |
| `order` | `asc` | Sort direction (default: DESC) |
| `page` / `per_page` | int | Pagination |
| `output` | `csv` | Download CSV |

---

## Status Dropdowns

### Ticket Statuses
```
GET /api/orders/ticket-statuses
```
```json
[
  { "id": 1, "name": "Complete - ready for pickup" },
  { "id": 2, "name": "Lab Initiated" }
]
```

### Invoice Statuses (patient-facing)
```
GET /api/orders/invoice-statuses
GET /api/orders/invoice-statuses?type=internal
```
Default type = `patient`. For internal dashboard use `?type=internal`.

---

## Dashboard Notes

### Ticket Note
```
POST /api/orders/ticket/{id}/dashboard-note
{ "dashboard_note": "Called patient" }
```
Clear: `{ "dashboard_note": null }`

### Invoice Note
```
POST /api/orders/invoice/{id}/dashboard-note
{ "dashboard_note": "Waiting insurance" }
```

---

## Invoice Status Update

```
PUT /api/patient/invoice/{invoice_id}/status
{ "status_invoice_id": 17 }
```
Only **patient** statuses allowed. Internal/vendor statuses are blocked.

Available statuses:
```
GET /api/patient/invoice/statuses
```

---

## Notifications

### Status → Notification Mapping
```
GET /api/orders/status-notification-map
GET /api/orders/status-notification-map?source=invoice
GET /api/orders/status-notification-map?source=ticket
```
Returns which SMS template and email category are linked to each status.

```json
[
  {
    "id": 1,
    "status_source": "invoice",
    "status_id": 17,
    "sms_template_id": 3,
    "sms_name": "ready_for_pickup",
    "sms_body": "Hi {{.patient_name}}, your order is ready for pickup at {{.location}}.",
    "email_category": "order",
    "auto_send": false
  }
]
```

**Frontend logic:**
1. When user changes status → check if `status_notification_map` has a mapping for that status
2. If mapping exists and `auto_send = false` → show "Notify Patient" button
3. Button click → `POST /api/orders/notify`

### Send Notification
```
POST /api/orders/notify
{
  "status_source": "invoice",
  "status_id": 17,
  "patient_id": 42,
  "send_sms": true,
  "send_email": true,
  "vars": {
    "patient_name": "Elon Musk",
    "location": "Fabulous Optical"
  }
}
```
- `patient_name` and `location` auto-filled from DB if not provided in `vars`
- `phone` and `email` auto-fetched from patient record if not provided
- Override with `"phone": "+1234567890"` or `"email": "test@test.com"`

Response:
```json
{
  "sms": { "status": "sent", "message": "Hi Elon Musk, your order is ready..." },
  "email": { "status": "sent" }
}
```

### Send Free-form SMS
```
POST /api/orders/send-sms
{
  "phone": "+13475445138",
  "patient_id": 42,
  "template_id": 3,
  "vars": { "patient_name": "Elon Musk", "location": "Fabulous Optical" }
}
```
Or free text (no template):
```json
{
  "phone": "+13475445138",
  "patient_id": 42,
  "message": "Hi, please call us back."
}
```

---

## SMS Templates (Settings)

**Permission required:** 80 (Settings)

### List
```
GET /api/sms-templates
GET /api/sms-templates?category=order
```
```json
[
  {
    "id": 3,
    "category": "order",
    "name": "ready_for_pickup",
    "body": "Hi {{.patient_name}}, your order is ready for pickup at {{.location}}.",
    "is_system": true,
    "active": true
  }
]
```

### Available Variables
```
GET /api/sms-templates/variables
```
```json
[
  { "var": "{{.patient_name}}", "description": "Patient full name" },
  { "var": "{{.location}}", "description": "Store/location name" },
  { "var": "{{.location_phone}}", "description": "Store phone number" },
  { "var": "{{.location_address}}", "description": "Store address" },
  { "var": "{{.doctor}}", "description": "Doctor name (with Dr. prefix)" },
  { "var": "{{.invoice_number}}", "description": "Invoice number" },
  { "var": "{{.ticket_number}}", "description": "Lab ticket number" }
]
```

### Create Custom Template
```
POST /api/sms-templates
{
  "category": "order",
  "name": "custom_followup",
  "body": "Hi {{.patient_name}}! Your order is complete. Please visit {{.location}} at {{.location_address}}"
}
```

### Edit Template
```
PUT /api/sms-templates/{id}
{
  "body": "Updated message text {{.patient_name}}",
  "active": true
}
```
- System templates: can edit `body` and `active`, cannot rename or delete
- Custom templates: can edit everything

### Delete Custom Template
```
DELETE /api/sms-templates/{id}
```
System templates cannot be deleted (returns 403).
