package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	commHandler "sighthub-backend/internal/handlers/patient_handler/communication"
	fileHandler "sighthub-backend/internal/handlers/patient_handler/file"
	infoHandler "sighthub-backend/internal/handlers/patient_handler/info"
	insHandler "sighthub-backend/internal/handlers/patient_handler/insurance"
	invHandler "sighthub-backend/internal/handlers/patient_handler/invoice"
	recallHandler "sighthub-backend/internal/handlers/patient_handler/recall"
	reportHandler "sighthub-backend/internal/handlers/patient_handler/report"
	rxHandler "sighthub-backend/internal/handlers/patient_handler/rx"
	"sighthub-backend/internal/middleware"
	infoSvc "sighthub-backend/internal/services/patient_service/info"
	invSvc "sighthub-backend/internal/services/patient_service/invoice"
	recallSvc "sighthub-backend/internal/services/patient_service/recall"
	rxSvc "sighthub-backend/internal/services/patient_service/rx"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterPatientRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	// ─── Services ──────────────────────────────────────────────────────────────
	infoService := infoSvc.New(db, cfg.DBName)
	rxService   := rxSvc.New(db)
	invoiceService := invSvc.New(db)
	recallService := recallSvc.New(db)

	// ─── Handlers ──────────────────────────────────────────────────────────────
	ih  := infoHandler.New(infoService)
	rh  := rxHandler.New(rxService)
	ivh := invHandler.New(invoiceService)
	fh  := fileHandler.New(db)
	ish := insHandler.New(db)
	ch  := commHandler.New(db)
	rch := recallHandler.New(db)
	rlh := recallHandler.NewListHandler(recallService)
	rpt := reportHandler.New(db)

	// ─── Middleware ────────────────────────────────────────────────────────────
	jwtMW  := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	perm51 := middleware.ActivePermission(db, 51)
	perm52 := middleware.ActivePermission(db, 52)
	perm53 := middleware.ActivePermission(db, 53)
	perm54 := middleware.ActivePermission(db, 54)
	perm55 := middleware.ActivePermission(db, 55)
	perm56 := middleware.ActivePermission(db, 56)
	perm57 := middleware.ActivePermission(db, 57)
	perm62 := middleware.ActivePermission(db, 62)
	perm63 := middleware.ActivePermission(db, 63)
	perm64 := middleware.ActivePermission(db, 64)
	perm65 := middleware.ActivePermission(db, 65)
	perm68 := middleware.ActivePermission(db, 68)
	perm69 := middleware.ActivePermission(db, 69)
	perm70 := middleware.ActivePermission(db, 70)
	perm71 := middleware.ActivePermission(db, 71)
	perm72 := middleware.ActivePermission(db, 72)
	perm73 := middleware.ActivePermission(db, 73)
	perm74 := middleware.ActivePermission(db, 74)
	perm75 := middleware.ActivePermission(db, 75)
	perm76 := middleware.ActivePermission(db, 76)
	perm77 := middleware.ActivePermission(db, 77)
	perm78 := middleware.ActivePermission(db, 78)
	perm79 := middleware.ActivePermission(db, 79)

	// ─── Subrouter ─────────────────────────────────────────────────────────────
	api := r.PathPrefix("/api/patient").Subrouter()
	api.Use(jwtMW, perm51)

	// ─── Info routes ───────────────────────────────────────────────────────────

	api.HandleFunc("/languages", ih.GetLanguages).Methods("GET")
	api.HandleFunc("/country_codes", ih.GetCountryCodes).Methods("GET")

	api.HandleFunc("/add", ih.CreatePatient).Methods("POST")
	api.HandleFunc("/patient/search", ih.SearchPatients).Methods("GET")

	api.HandleFunc("/{patient_id:[0-9]+}", ih.GetPatient).Methods("GET")
	api.Handle("/{patient_id:[0-9]+}",
		perm52(http.HandlerFunc(ih.UpdatePatient)),
	).Methods("PUT")
	api.HandleFunc("/remove/{patient_id:[0-9]+}", ih.DeletePatient).Methods("DELETE")

	api.HandleFunc("/{patient_id:[0-9]+}/generate-filename-doc", ih.GenerateFilenameDoc).Methods("GET")
	api.HandleFunc("/{patient_id:[0-9]+}/generate-filename-prescription", ih.GenerateFilenamePrescription).Methods("GET")
	api.HandleFunc("/{patient_id:[0-9]+}/generate-filename-insurance-policy", ih.GenerateFilenameInsurancePolicy).Methods("GET")

	api.HandleFunc("/{patient_id:[0-9]+}/prescriptions", ih.GetPatientPrescriptions).Methods("GET")
	api.HandleFunc("/appointments", ih.GetPatientAppointments).Methods("GET")

	// ─── Communication history + send ──────────────────────────────────────────

	api.HandleFunc("/{patient_id:[0-9]+}/sms-history", ch.GetSMSHistory).Methods("GET")
	api.HandleFunc("/{patient_id:[0-9]+}/send-sms", ch.SendSMS).Methods("POST")
	api.HandleFunc("/{patient_id:[0-9]+}/email-history", ch.GetEmailHistory).Methods("GET")
	api.HandleFunc("/{patient_id:[0-9]+}/send-email", ch.SendEmail).Methods("POST")
	api.HandleFunc("/{patient_id:[0-9]+}/call-history", ch.GetCallHistory).Methods("GET")
	api.HandleFunc("/{patient_id:[0-9]+}/log-call", ch.LogCall).Methods("POST")

	// ─── Recall routes ─────────────────────────────────────────────────────────

	api.HandleFunc("/recall-list", rlh.GetRecallList).Methods("GET")
	api.HandleFunc("/recall/{recall_id:[0-9]+}/result", rlh.LogCallResult).Methods("POST")

	api.HandleFunc("/{patient_id:[0-9]+}/recall", rch.GetRecalls).Methods("GET")
	api.HandleFunc("/{patient_id:[0-9]+}/recall", rch.CreateRecall).Methods("POST")
	api.HandleFunc("/{patient_id:[0-9]+}/recall/{recall_id:[0-9]+}", rch.DeleteRecall).Methods("DELETE")

	// ─── File routes ───────────────────────────────────────────────────────────

	api.HandleFunc("/file", fh.GetPatientFiles).Methods("GET")
	api.Handle("/file",
		perm56(http.HandlerFunc(fh.UploadFile)),
	).Methods("POST")
	api.HandleFunc("/file/{id_file:[0-9]+}", fh.GetFile).Methods("GET")
	api.Handle("/file/{id_file:[0-9]+}",
		perm56(http.HandlerFunc(fh.UpdateFile)),
	).Methods("PUT")
	api.Handle("/file/{id_file:[0-9]+}",
		perm57(http.HandlerFunc(fh.DeleteFile)),
	).Methods("DELETE")

	// ─── Insurance routes ──────────────────────────────────────────────────────

	api.HandleFunc("/insurance", ish.GetInsurancePolicies).Methods("GET")
	api.HandleFunc("/insurance", ish.CreateInsurancePolicy).Methods("POST")
	api.HandleFunc("/insurance/coverage_types", ish.GetCoverageTypes).Methods("GET")
	api.HandleFunc("/insurance/companies", ish.GetCompanies).Methods("GET")
	api.HandleFunc("/insurance/{id_insurance:[0-9]+}/patient/{id_patient:[0-9]+}", ish.GetInsurancePolicyByID).Methods("GET")
	api.Handle("/insurance/{id_insurance:[0-9]+}",
		perm52(http.HandlerFunc(ish.UpdateInsurancePolicy)),
	).Methods("PUT")
	api.Handle("/insurance/{id_insurance:[0-9]+}/holders",
		perm52(http.HandlerFunc(ish.AddHolder)),
	).Methods("POST")
	api.Handle("/insurance/{id_insurance:[0-9]+}/holders/{patient_id:[0-9]+}",
		perm52(http.HandlerFunc(ish.UpdateHolder)),
	).Methods("PUT")
	api.Handle("/insurance/{id_insurance:[0-9]+}/holders/{patient_id:[0-9]+}",
		perm52(http.HandlerFunc(ish.DeleteHolder)),
	).Methods("DELETE")
	api.Handle("/insurance/{id_insurance:[0-9]+}/{id_patient:[0-9]+}",
		perm52(http.HandlerFunc(ish.DeleteInsurancePolicy)),
	).Methods("DELETE")

	// ─── Report routes ─────────────────────────────────────────────────────────

	api.HandleFunc("/report/all_patients/csv", rpt.AllPatientsCSV).Methods("GET")

	// ─── Rx routes ─────────────────────────────────────────────────────────────

	api.HandleFunc("/latest-rx", rh.GetLatestRx).Methods("GET")
	api.HandleFunc("/rx-list", rh.GetRxList).Methods("GET")
	api.HandleFunc("/rx/doctors", rh.GetDoctors).Methods("GET")
	api.HandleFunc("/rx", rh.GetRx).Methods("GET")
	api.Handle("/rx",
		perm54(http.HandlerFunc(rh.CreateRx)),
	).Methods("POST")
	api.Handle("/rx",
		perm53(http.HandlerFunc(rh.UpdateRx)),
	).Methods("PUT")
	api.Handle("/rx/{id_rx:[0-9]+}",
		perm55(http.HandlerFunc(rh.DeleteRx)),
	).Methods("DELETE")

	// ─── Invoice core routes ────────────────────────────────────────────────────

	// Statuses and payment-methods (no extra perm beyond perm51)
	api.HandleFunc("/invoice/statuses", ivh.GetInvoiceStatuses).Methods("GET")
	api.HandleFunc("/payment-methods", ivh.GetPaymentMethods).Methods("GET")
	api.HandleFunc("/invoices/search", ivh.SearchInvoices).Methods("GET")
	api.HandleFunc("/lookup", ivh.LookupBySKU).Methods("GET")

	// Invoice list (GET) and create (POST, perm 62)
	api.HandleFunc("/invoice", ivh.GetInvoiceList).Methods("GET")
	api.Handle("/invoice",
		perm62(http.HandlerFunc(ivh.CreateInvoice)),
	).Methods("POST")

	// Finalize / unfinalize (must be before /{invoice_id} to avoid mux conflict)
	api.Handle("/invoice/finalize/{invoice_id:[0-9]+}",
		perm77(http.HandlerFunc(ivh.FinalizeInvoice)),
	).Methods("PUT")
	api.Handle("/invoice/unfinalize/{invoice_id:[0-9]+}",
		perm78(http.HandlerFunc(ivh.UnfinalizeInvoice)),
	).Methods("PUT")

	// Invoice detail, update items, delete
	api.HandleFunc("/invoice/{invoice_id:[0-9]+}", ivh.GetInvoice).Methods("GET")
	api.Handle("/invoice/{invoice_id:[0-9]+}",
		perm62(http.HandlerFunc(ivh.AddItemsToInvoice)),
	).Methods("PUT")
	api.Handle("/invoice/{invoice_id:[0-9]+}",
		perm79(http.HandlerFunc(ivh.DeleteInvoice)),
	).Methods("DELETE")

	// Remake
	api.Handle("/invoice/{invoice_id:[0-9]+}/remake",
		perm62(http.HandlerFunc(ivh.CreateRemakeInvoice)),
	).Methods("POST")

	// HTML / print / PDF
	api.HandleFunc("/invoice/{invoice_id:[0-9]+}/html", ivh.GetInvoiceHTML).Methods("GET")
	api.HandleFunc("/invoice/{invoice_id:[0-9]+}/print", ivh.PrintInvoice).Methods("GET")
	api.HandleFunc("/invoice/{invoice_id:[0-9]+}/pdf", ivh.GetInvoicePDF).Methods("GET")

	// ─── Invoice payment routes ─────────────────────────────────────────────────

	api.HandleFunc("/insurance-payment-types", ivh.GetInsurancePaymentTypes).Methods("GET")

	api.Handle("/invoice/{invoice_id:[0-9]+}/patient-payment",
		perm64(http.HandlerFunc(ivh.AddPatientPayment)),
	).Methods("POST")
	api.Handle("/invoice/{invoice_id:[0-9]+}/credit/payment",
		perm64(http.HandlerFunc(ivh.PayWithCredit)),
	).Methods("POST")
	api.Handle("/invoice/{invoice_id:[0-9]+}/discount",
		perm65(http.HandlerFunc(ivh.AddDiscount)),
	).Methods("POST")
	api.Handle("/invoice/{invoice_id:[0-9]+}/insurance-payment",
		perm64(http.HandlerFunc(ivh.AddInsurancePayment)),
	).Methods("POST")
	api.Handle("/invoice/{invoice_id:[0-9]+}/transfer-credit",
		perm64(http.HandlerFunc(ivh.TransferCredit)),
	).Methods("POST")

	api.HandleFunc("/invoice/{invoice_id:[0-9]+}/insurance-payments", ivh.GetInsurancePayments).Methods("GET")
	api.HandleFunc("/invoice/{invoice_id:[0-9]+}/payment_history", ivh.GetPaymentHistory).Methods("GET")

	api.Handle("/invoice/{invoice_id:[0-9]+}/payment/{payment_id:[0-9]+}",
		perm74(http.HandlerFunc(ivh.UpdatePayment)),
	).Methods("PUT")
	api.Handle("/invoice/{invoice_id:[0-9]+}/payment/{payment_id:[0-9]+}",
		perm75(http.HandlerFunc(ivh.DeletePayment)),
	).Methods("DELETE")
	api.Handle("/invoice/{invoice_id:[0-9]+}/insurance-payment/{payment_id:[0-9]+}",
		perm75(http.HandlerFunc(ivh.DeleteInsurancePayment)),
	).Methods("DELETE")

	api.HandleFunc("/credit_balance/{patient_id:[0-9]+}", ivh.GetCreditBalance).Methods("GET")
	api.HandleFunc("/{patient_id:[0-9]+}/credit/payments", ivh.GetCreditPayments).Methods("GET")

	// ─── Invoice item routes ────────────────────────────────────────────────────

	api.Handle("/invoice/{invoice_id:[0-9]+}/item/{item_sale_id:[0-9]+}",
		perm63(http.HandlerFunc(ivh.UpdateItem)),
	).Methods("PUT")
	api.Handle("/invoice/{invoice_id:[0-9]+}/item/{item_sale_id:[0-9]+}",
		perm76(http.HandlerFunc(ivh.DeleteItem)),
	).Methods("DELETE")
	api.Handle("/invoice/{invoice_id:[0-9]+}/item/{item_sale_id:[0-9]+}/balance",
		perm62(http.HandlerFunc(ivh.SetLineBalance)),
	).Methods("PUT")

	api.Handle("/invoice/{invoice_id:[0-9]+}/insurance/add",
		perm62(http.HandlerFunc(ivh.AddInsurancePolicy)),
	).Methods("POST")
	api.Handle("/invoice/insurance/{invoice_id:[0-9]+}",
		perm70(http.HandlerFunc(ivh.DeleteInsuranceFromInvoice)),
	).Methods("DELETE")

	api.Handle("/invoice/{invoice_id:[0-9]+}/giftcard",
		perm68(http.HandlerFunc(ivh.AddGiftCard)),
	).Methods("POST")
	api.Handle("/invoice/gift-card/{invoice_id:[0-9]+}",
		perm69(http.HandlerFunc(ivh.DeleteGiftCard)),
	).Methods("DELETE")

	// ─── Invoice return routes ──────────────────────────────────────────────────

	api.Handle("/invoice/{invoice_id:[0-9]+}/return",
		perm71(http.HandlerFunc(ivh.ProcessReturn)),
	).Methods("POST")
	api.Handle("/return/{return_id:[0-9]+}/deny",
		perm71(http.HandlerFunc(ivh.DenyReturn)),
	).Methods("PUT")
	api.Handle("/return/{return_id:[0-9]+}/confirm",
		perm73(http.HandlerFunc(ivh.ConfirmReturn)),
	).Methods("PUT")
	api.Handle("/return/{return_id:[0-9]+}",
		perm72(http.HandlerFunc(ivh.DeleteReturn)),
	).Methods("DELETE")
	api.HandleFunc("/invoice/{invoice_id:[0-9]+}/returns", ivh.GetReturnsByInvoice).Methods("GET")
	api.HandleFunc("/return/{return_id:[0-9]+}", ivh.GetReturn).Methods("GET")

	// ── Invoice Price Book (source-filtered) ──────────────────────────────
	api.HandleFunc("/invoice/{invoice_id:[0-9]+}/price-book/lens/list", ivh.InvPBLensList).Methods("GET")
	api.HandleFunc("/invoice/{invoice_id:[0-9]+}/price-book/treatment/list", ivh.InvPBTreatmentList).Methods("GET")
	api.HandleFunc("/invoice/{invoice_id:[0-9]+}/price-book/additional/list", ivh.InvPBAddServiceList).Methods("GET")
}
