# Insurance API — Coverage Type Decoupling

## Summary

Coverage type (`insurance_coverage_types`) is **no longer tied to the insurance company**.
It is now a property of the **insurance policy** only.

**Before:** `insurance_company` had `id_insurance_coverage_type` FK -> coverage was inherited from company to policy.
**After:** Coverage type lives only on `insurance_policy.insurance_coverage_type_id`. Frontend must pass it explicitly when creating/updating a policy.

---

## Database Changes

- **Dropped column:** `insurance_company.id_insurance_coverage_type`
- **Unchanged:** `insurance_policy.insurance_coverage_type_id` (FK to `insurance_coverage_types`)

---

## Unified Coverage Type Response Format

All endpoints that return coverage types now use the same JSON shape:

```json
{
  "id_insurance_coverage_type": 1,
  "coverage_name": "Vision"
}
```

---

## Endpoints by Module

### `/api/settings/insurance`

| Method | Path | Change |
|--------|------|--------|
| GET | `/companies` | No longer returns `id_insurance_coverage_type` or `coverage_name` |
| POST | `/companies` | No longer accepts `insurance_coverage_type_id` — only `company_name` |
| PUT | `/companies/{id}` | No longer accepts coverage type fields |
| DELETE | `/companies/{id}` | No change |
| GET | `/coverage_types` | No change — returns `id_insurance_coverage_type`, `coverage_name` |
| GET | `/types` | Response keys changed: `id_insurance_coverage_type`, `coverage_name` (was `id_type_insurance_policy`, `type_name`) |

#### POST `/companies` — new body:
```json
{
  "company_name": "Blue Cross"
}
```

---

### `/api/patient/insurance`

| Method | Path | Change |
|--------|------|--------|
| GET | `?id_patient=X` | Still returns `id_insurance_coverage_type` and `coverage_name` (from policy) |
| POST | `/` | Now accepts `insurance_coverage_type_id` in body (was auto-inherited from company) |
| GET | `/coverage_types` | No change |
| GET | `/companies` | Simplified — returns only `id_insurance_company` + `company_name` (no coverage prefix) |
| GET | `/{id_insurance}/patient/{id_patient}` | Removed `company_coverage_type_*` and `effective_coverage_type_*` fields. Only `insurance_coverage_type_id` and `insurance_coverage_type_name` remain |
| PUT | `/{id_insurance}` | Changing `insurance_company_id` no longer auto-sets coverage. Must pass `insurance_coverage_type_id` explicitly |

#### POST `/` — create policy body:
```json
{
  "id_patient": 123,
  "member_number": "ABC123",
  "insurance_company_id": 1,
  "insurance_coverage_type_id": 2,
  "group_number": "GRP001",
  "coverage_details": "Full coverage",
  "specify": "Some notes",
  "active": true,
  "front_photo": "path/to/front.jpg",
  "back_photo": "path/to/back.jpg",
  "policy_holder": "Self",
  "holder_id": null
}
```

#### GET `/{id_insurance}/patient/{id_patient}` — response:
```json
{
  "insurance_info": {
    "id_insurance": 1,
    "insurance_company_id": 1,
    "company_name": "Blue Cross",
    "group_number": "GRP001",
    "coverage_details": "Full coverage",
    "specify": null,
    "active": true,
    "front_photo": null,
    "back_photo": null,
    "insurance_coverage_type_id": 2,
    "insurance_coverage_type_name": "Vision"
  },
  "current_holder": { ... },
  "holders": [ ... ]
}
```

**Removed fields from response:**
- `company_coverage_type_id`
- `company_coverage_type_name`
- `effective_coverage_type_id`
- `effective_coverage_type_name`

---

### `/api/home`

| Method | Path | Response |
|--------|------|----------|
| GET | `/insurance/companies` | `[{ "insurance_company_id": 1, "company_name": "..." }]` — no change |
| GET | `/insurance/types` | Keys changed to `id_insurance_coverage_type`, `coverage_name` (was `id_type_insurance_policy`, `type_name`) |

---

### `/api/sale`

| Method | Path | Change |
|--------|------|--------|
| GET | `/insurance` | No change — returns `id_insurance_company`, `company_name` |
| GET | `/insurance/coverage_types` | **NEW** — returns `[{ "id_insurance_coverage_type": 1, "coverage_name": "Vision" }]` |

---

### `/api/claim`

| Method | Path | Change |
|--------|------|--------|
| GET | `/insurance-companies` | No change |
| GET | `/insurance-coverage-types` | **NEW** — returns `[{ "id_insurance_coverage_type": 1, "coverage_name": "Vision" }]` |

---

### Doctor Desk

`GET /api/doctor-desk/patient-info/{id}` now includes `coverage_name` in the insurance object:

```json
{
  "insurance": {
    "company_name": "Blue Cross",
    "coverage_name": "Vision",
    "group_number": "GRP001",
    "holder_type": "Self"
  }
}
```

---

## Coverage Types — Available Endpoints

For convenience, here is a list of all endpoints where frontend can fetch coverage types:

| Module | Endpoint |
|--------|----------|
| Patient | `GET /api/patient/insurance/coverage_types` |
| Settings | `GET /api/settings/insurance/coverage_types` |
| Settings | `GET /api/settings/insurance/types` |
| Home | `GET /api/home/insurance/types` |
| Sale | `GET /api/sale/insurance/coverage_types` |
| Claim | `GET /api/claim/insurance-coverage-types` |

All return the same format:
```json
[
  { "id_insurance_coverage_type": 1, "coverage_name": "Vision" },
  { "id_insurance_coverage_type": 2, "coverage_name": "Medical" }
]
```

---

## Migration Notes for Frontend

1. **Creating insurance policy:** Send `insurance_coverage_type_id` explicitly — it is no longer auto-inherited from the company.
2. **Insurance company dropdowns:** No longer include coverage type info. Use a separate dropdown for coverage type.
3. **Display logic:** Use `insurance_coverage_type_id` / `insurance_coverage_type_name` from the policy object. No more "effective" or "company" coverage fallback.
4. **Settings → Companies:** Create/update no longer handles coverage type. Coverage is a separate entity.
5. **Key naming:** All endpoints now use `id_insurance_coverage_type` and `coverage_name` consistently (legacy `id_type_insurance_policy` / `type_name` removed).
