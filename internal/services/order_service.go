package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"github.com/1kulture/1kulture-backend/internal/config"
	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/payments"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/idgen"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type orderService struct {
	orderRepo      repoInterfaces.OrderRepository
	orderItemRepo  repoInterfaces.OrderItemRepository
	ticketRepo     repoInterfaces.TicketRepository
	ttRepo         repoInterfaces.TicketTypeRepository
	eventRepo      repoInterfaces.EventRepository
	coOrgRepo      repoInterfaces.EventCoOrganizerRepository
	promoRepo      repoInterfaces.PromoCodeRepository
	redemptionRepo repoInterfaces.PromoCodeRedemptionRepository
	txnRepo        repoInterfaces.PaymentTransactionRepository
	commissionSvc  serviceInterfaces.CommissionService
	ledgerSvc      serviceInterfaces.LedgerService
	settingSvc     serviceInterfaces.SettingService
	provider       payments.PaymentProvider
	auditLogRepo   repoInterfaces.AuditLogRepository
	cfg            *config.Config
}

func NewOrderService(
	orderRepo repoInterfaces.OrderRepository,
	orderItemRepo repoInterfaces.OrderItemRepository,
	ticketRepo repoInterfaces.TicketRepository,
	ttRepo repoInterfaces.TicketTypeRepository,
	eventRepo repoInterfaces.EventRepository,
	coOrgRepo repoInterfaces.EventCoOrganizerRepository,
	promoRepo repoInterfaces.PromoCodeRepository,
	redemptionRepo repoInterfaces.PromoCodeRedemptionRepository,
	txnRepo repoInterfaces.PaymentTransactionRepository,
	commissionSvc serviceInterfaces.CommissionService,
	ledgerSvc serviceInterfaces.LedgerService,
	settingSvc serviceInterfaces.SettingService,
	provider payments.PaymentProvider,
	auditLogRepo repoInterfaces.AuditLogRepository,
	cfg *config.Config,
) serviceInterfaces.OrderService {
	return &orderService{
		orderRepo:      orderRepo,
		orderItemRepo:  orderItemRepo,
		ticketRepo:     ticketRepo,
		ttRepo:         ttRepo,
		eventRepo:      eventRepo,
		coOrgRepo:      coOrgRepo,
		promoRepo:      promoRepo,
		redemptionRepo: redemptionRepo,
		txnRepo:        txnRepo,
		commissionSvc:  commissionSvc,
		ledgerSvc:      ledgerSvc,
		settingSvc:     settingSvc,
		provider:       provider,
		auditLogRepo:   auditLogRepo,
		cfg:            cfg,
	}
}

// ---------- CreateOrder ----------

func (s *orderService) CreateOrder(ctx context.Context, userID uuid.UUID, req *requests.CreateOrderRequest) (*responses.OrderResponse, error) {
	// Idempotency: return existing order if same key was used
	if req.IdempotencyKey != "" {
		existing, err := s.orderRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			full, _ := s.orderRepo.FindWithItems(ctx, existing.ID)
			return s.toOrderResponse(full), nil
		}
	}

	eventID, err := uuid.Parse(req.EventID)
	if err != nil {
		return nil, fmt.Errorf("invalid event_id")
	}
	event, err := s.eventRepo.FindWithRelations(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}
	if event.Status != models.EventStatusPublished {
		return nil, fmt.Errorf("event is not on sale")
	}

	now := time.Now().UTC()
	var subtotal int64
	orderItems := make([]models.OrderItem, 0, len(req.Items))
	reserved := make([]struct {
		ttID uuid.UUID
		qty  int
	}, 0, len(req.Items))

	releaseAll := func() {
		for _, r := range reserved {
			_ = s.ttRepo.ReleaseReservation(ctx, r.ttID, r.qty)
		}
	}

	for _, input := range req.Items {
		ttID, err := uuid.Parse(input.TicketTypeID)
		if err != nil {
			releaseAll()
			return nil, fmt.Errorf("invalid ticket_type_id: %s", input.TicketTypeID)
		}
		tt, err := s.ttRepo.FindByID(ctx, ttID)
		if err != nil {
			releaseAll()
			return nil, err
		}
		if tt == nil || tt.EventID != eventID {
			releaseAll()
			return nil, fmt.Errorf("ticket type %s does not belong to this event", ttID)
		}
		if !tt.IsOnSale(now) {
			releaseAll()
			return nil, fmt.Errorf("ticket type '%s' is not on sale", tt.Name)
		}
		if tt.Remaining() < input.Quantity {
			releaseAll()
			return nil, fmt.Errorf("only %d tickets remaining for '%s'", tt.Remaining(), tt.Name)
		}

		// Per-user purchase limit
		if tt.PerUserLimit > 0 {
			myTickets, _, err := s.ticketRepo.FindByUser(ctx, userID, 1, 1000)
			if err != nil {
				releaseAll()
				return nil, err
			}
			count := 0
			for _, t := range myTickets {
				if t.TicketTypeID == tt.ID &&
					t.Status != models.TicketStatusCancelled &&
					t.Status != models.TicketStatusRefunded {
					count++
				}
			}
			if count+input.Quantity > tt.PerUserLimit {
				releaseAll()
				return nil, fmt.Errorf("purchase limit of %d reached for '%s'", tt.PerUserLimit, tt.Name)
			}
		}

		// Atomic stock reservation
		if err := s.ttRepo.ReserveQuantity(ctx, tt.ID, input.Quantity); err != nil {
			releaseAll()
			return nil, err
		}
		reserved = append(reserved, struct {
			ttID uuid.UUID
			qty  int
		}{tt.ID, input.Quantity})

		lineTotal := int64(input.Quantity) * tt.PriceMinor
		subtotal += lineTotal

		orderItems = append(orderItems, models.OrderItem{
			TicketTypeID:   tt.ID,
			Quantity:       input.Quantity,
			UnitPriceMinor: tt.PriceMinor,
			TotalMinor:     lineTotal,
			TicketTypeName: tt.Name,
		})
	}

	// Commission snapshot
	rateBps, err := s.commissionSvc.ResolveRateBps(ctx, event.OrganizerID)
	if err != nil {
		releaseAll()
		return nil, err
	}

	// Promo code
	var discount int64
	var promoID *uuid.UUID
	if req.PromoCode != "" {
		code := strings.ToUpper(strings.TrimSpace(req.PromoCode))
		pc, err := s.promoRepo.FindByCode(ctx, code)
		if err != nil {
			releaseAll()
			return nil, err
		}
		if pc == nil {
			releaseAll()
			return nil, fmt.Errorf("invalid promo code")
		}
		if pc.EventID != nil && *pc.EventID != eventID {
			releaseAll()
			return nil, fmt.Errorf("promo code is not valid for this event")
		}
		if !pc.IsValid(now) {
			releaseAll()
			return nil, fmt.Errorf("promo code is not active")
		}
		if pc.MinOrderMinor > 0 && subtotal < pc.MinOrderMinor {
			releaseAll()
			return nil, fmt.Errorf("order subtotal does not meet the minimum for this promo code")
		}
		if pc.PerUserLimit > 0 {
			count, err := s.promoRepo.CountUserRedemptions(ctx, pc.ID, userID)
			if err != nil {
				releaseAll()
				return nil, err
			}
			if int(count) >= pc.PerUserLimit {
				releaseAll()
				return nil, fmt.Errorf("you have reached the usage limit for this promo code")
			}
		}
		discount = calculateDiscount(pc, subtotal)
		if discount > subtotal {
			discount = subtotal
		}
		id := pc.ID
		promoID = &id
	}

	total := subtotal - discount
	if total < 0 {
		total = 0
	}
	commission := total * int64(rateBps) / 10000
	organizerNet := total - commission

	// Buyer details snapshot
	billing := map[string]interface{}{
		"name":  req.BuyerName,
		"email": req.BuyerEmail,
		"phone": req.BuyerPhone,
	}
	billingJSON, _ := jsonMarshal(billing)

	// Reservation hold
	holdMinutes := s.cfg.Payments.OrderHoldMinutes
	if holdMinutes <= 0 {
		holdMinutes = 15
	}
	expiresAt := now.Add(time.Duration(holdMinutes) * time.Minute)

	order := &models.Order{
		UserID:            userID,
		EventID:           eventID,
		Reference:         idgen.OrderReference(),
		Status:            models.OrderStatusPending,
		Currency:          event.Currency,
		SubtotalMinor:     subtotal,
		DiscountMinor:     discount,
		TotalMinor:        total,
		CommissionRateBps: rateBps,
		CommissionMinor:   commission,
		OrganizerNetMinor: organizerNet,
		PromoCodeID:       promoID,
		EscrowStatus:      models.EscrowStatusNone,
		ExpiresAt:         &expiresAt,
		IdempotencyKey:    req.IdempotencyKey,
		BillingInfo:       datatypes.JSON(billingJSON),
	}
	if err := s.orderRepo.Create(ctx, order); err != nil {
		releaseAll()
		return nil, err
	}

	for i := range orderItems {
		orderItems[i].OrderID = order.ID
	}
	if err := s.orderItemRepo.CreateMany(ctx, orderItems); err != nil {
		releaseAll()
		return nil, err
	}

	// Promo redemption tracking
	if promoID != nil {
		redemption := &models.PromoCodeRedemption{
			PromoCodeID:   *promoID,
			UserID:        userID,
			OrderID:       order.ID,
			DiscountMinor: discount,
		}
		if err := s.redemptionRepo.Create(ctx, redemption); err != nil {
			logger.Error("Failed to record promo redemption: ", err)
		}
		if err := s.promoRepo.IncrementUsage(ctx, *promoID, 1); err != nil {
			logger.Error("Failed to increment promo usage: ", err)
		}
	}

	s.audit(ctx, &userID, "ORDER_CREATED", order.ID.String())
	full, _ := s.orderRepo.FindWithItems(ctx, order.ID)
	return s.toOrderResponse(full), nil
}

// ---------- InitializePayment ----------

func (s *orderService) InitializePayment(ctx context.Context, userID, orderID uuid.UUID, req *requests.InitializePaymentRequest) (*responses.InitializePaymentResponse, error) {
	order, err := s.orderRepo.FindWithItems(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.UserID != userID {
		return nil, fmt.Errorf("forbidden")
	}
	if order.Status != models.OrderStatusPending {
		return nil, fmt.Errorf("order is not in a payable state")
	}
	if order.ExpiresAt != nil && time.Now().UTC().After(*order.ExpiresAt) {
		return nil, fmt.Errorf("order has expired")
	}

	buyerEmail := extractBillingStr(order.BillingInfo, "email")
	if buyerEmail == "" {
		return nil, fmt.Errorf("buyer email is required to initialize payment")
	}

	reference := order.Reference
	callback := req.CallbackURL
	if callback == "" {
		callback = s.cfg.Payments.CallbackURL
	}

	initRes, err := s.provider.Initialize(ctx, payments.InitializeRequest{
		AmountMinor: order.TotalMinor,
		Currency:    order.Currency,
		Email:       buyerEmail,
		Reference:   reference,
		CallbackURL: callback,
		Metadata: map[string]interface{}{
			"order_id": order.ID.String(),
			"user_id":  userID.String(),
			"event_id": order.EventID.String(),
		},
	})
	if err != nil {
		return nil, err
	}

	rawReq, _ := jsonMarshal(map[string]interface{}{
		"reference": reference,
		"amount":    order.TotalMinor,
		"currency":  order.Currency,
		"email":     buyerEmail,
	})
	txn := &models.PaymentTransaction{
		OrderID:     order.ID,
		UserID:      userID,
		Provider:    s.provider.Name(),
		Reference:   reference,
		ProviderRef: initRes.AccessCode,
		AmountMinor: order.TotalMinor,
		Currency:    order.Currency,
		Status:      models.PaymentTxnInitiated,
		RawRequest:  datatypes.JSON(rawReq),
	}
	if err := s.txnRepo.Create(ctx, txn); err != nil {
		logger.Error("Failed to record payment transaction: ", err)
	}

	order.PaymentProvider = s.provider.Name()
	order.PaymentReference = reference
	if err := s.orderRepo.Update(ctx, order); err != nil {
		logger.Error("Failed to update order with payment reference: ", err)
	}

	s.audit(ctx, &userID, "ORDER_PAYMENT_INITIALIZED", order.ID.String())

	return &responses.InitializePaymentResponse{
		OrderID:          order.ID,
		Reference:        reference,
		AuthorizationURL: initRes.AuthorizationURL,
		AccessCode:       initRes.AccessCode,
		ExpiresAt:        derefTime(order.ExpiresAt),
	}, nil
}

// ---------- HandleWebhook ----------

func (s *orderService) HandleWebhook(ctx context.Context, provider string, rawBody []byte, signature string) error {
	if provider != s.provider.Name() {
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	event, err := s.provider.ParseWebhook(rawBody, signature)
	if err != nil {
		return err
	}

	if event.EventType != "charge.success" {
		return nil
	}

	order, err := s.orderRepo.FindByReference(ctx, event.Reference)
	if err != nil {
		return err
	}
	if order == nil {
		return fmt.Errorf("order not found for reference %s", event.Reference)
	}

	// Idempotency: already paid → no-op
	if order.Status == models.OrderStatusPaid {
		return nil
	}

	// Double-check with provider
	verify, err := s.provider.Verify(ctx, event.Reference)
	if err != nil {
		return err
	}
	if verify.Status != "success" {
		s.markTransaction(ctx, order.ID, verify, models.PaymentTxnFailed, "verify status: "+verify.Status)
		return fmt.Errorf("payment verification failed: status=%s", verify.Status)
	}
	if verify.AmountMinor < order.TotalMinor {
		s.markTransaction(ctx, order.ID, verify, models.PaymentTxnFailed, "amount mismatch")
		return fmt.Errorf("paid amount (%d) is less than order total (%d)", verify.AmountMinor, order.TotalMinor)
	}

	paidAt := time.Unix(verify.PaidAt, 0).UTC()
	if verify.PaidAt == 0 {
		paidAt = time.Now().UTC()
	}
	order.Status = models.OrderStatusPaid
	order.PaidAt = &paidAt
	order.PaymentProvider = s.provider.Name()
	order.PaymentReference = event.Reference

	releaseDays := s.settingSvc.GetInt(ctx, models.SettingEscrowReleaseDays, 3)
	releaseAt := time.Now().UTC().AddDate(0, 0, releaseDays)
	order.EscrowStatus = models.EscrowStatusHeld
	order.EscrowReleaseAt = &releaseAt

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	full, err := s.orderRepo.FindWithItems(ctx, order.ID)
	if err != nil {
		return err
	}
	eventWithOrg, err := s.eventRepo.FindWithRelations(ctx, order.EventID)
	if err != nil {
		return err
	}

	// Commit reservations → sold
	for _, item := range full.Items {
		if err := s.ttRepo.CommitReservation(ctx, item.TicketTypeID, item.Quantity); err != nil {
			logger.Error("Failed to commit reservation: ", err)
			return err
		}
	}

	// Issue tickets
	tickets := make([]models.Ticket, 0)
	for _, item := range full.Items {
		for i := 0; i < item.Quantity; i++ {
			code := idgen.TicketCode()
			tickets = append(tickets, models.Ticket{
				OrderID:      order.ID,
				OrderItemID:  item.ID,
				TicketTypeID: item.TicketTypeID,
				EventID:      order.EventID,
				UserID:       order.UserID,
				Code:         code,
				QRPayload:    code,
				HolderName:   extractBillingStr(full.BillingInfo, "name"),
				HolderEmail:  extractBillingStr(full.BillingInfo, "email"),
				HolderPhone:  extractBillingStr(full.BillingInfo, "phone"),
				Status:       models.TicketStatusValid,
				IssuedAt:     time.Now().UTC(),
			})
		}
	}
	if err := s.ticketRepo.CreateMany(ctx, tickets); err != nil {
		return fmt.Errorf("failed to issue tickets: %w", err)
	}

	// Ledger: payment into escrow + platform fee
	order.Event = *eventWithOrg
	if err := s.ledgerSvc.RecordOrderPaid(ctx, order); err != nil {
		logger.Error("Failed to record ledger entries: ", err)
	}

	s.markTransaction(ctx, order.ID, verify, models.PaymentTxnSuccess, "")
	s.audit(ctx, &order.UserID, "ORDER_PAID", order.ID.String())
	return nil
}

func (s *orderService) markTransaction(ctx context.Context, orderID uuid.UUID, verify *payments.VerifyResult, status models.PaymentTransactionStatus, msg string) {
	txns, err := s.txnRepo.FindByOrder(ctx, orderID)
	if err != nil || len(txns) == 0 {
		return
	}
	txn := txns[0]
	txn.Status = status
	txn.StatusMessage = msg
	if verify != nil {
		txn.ProviderRef = verify.ProviderRef
		if verify.RawResponse != nil {
			txn.RawResponse = datatypes.JSON(verify.RawResponse)
		}
		now := time.Now().UTC()
		if status == models.PaymentTxnSuccess {
			txn.AuthorizedAt = &now
			txn.CompletedAt = &now
		}
	}
	if err := s.txnRepo.Update(ctx, &txn); err != nil {
		logger.Error("Failed to update payment transaction: ", err)
	}
}

// ---------- Reads ----------

func (s *orderService) GetOrder(ctx context.Context, viewerID, orderID uuid.UUID) (*responses.OrderResponse, error) {
	order, err := s.orderRepo.FindWithItems(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.UserID == viewerID {
		return s.toOrderResponse(order), nil
	}
	event, err := s.eventRepo.FindByID(ctx, order.EventID)
	if err != nil {
		return nil, err
	}
	if event != nil {
		if event.OrganizerID == viewerID {
			return s.toOrderResponse(order), nil
		}
		isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, event.ID, viewerID)
		if isCo {
			return s.toOrderResponse(order), nil
		}
	}
	return nil, fmt.Errorf("forbidden")
}

func (s *orderService) ListMyOrders(ctx context.Context, userID uuid.UUID, page, perPage int) ([]responses.OrderResponse, int64, error) {
	list, total, err := s.orderRepo.ListByUser(ctx, userID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]responses.OrderResponse, 0, len(list))
	for i := range list {
		out = append(out, *s.toOrderResponse(&list[i]))
	}
	return out, total, nil
}

func (s *orderService) ListEventOrders(ctx context.Context, actorID, eventID uuid.UUID, page, perPage int) ([]responses.OrderResponse, int64, error) {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, 0, err
	}
	if event == nil {
		return nil, 0, fmt.Errorf("event not found")
	}
	if event.OrganizerID != actorID {
		isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, eventID, actorID)
		if !isCo {
			return nil, 0, fmt.Errorf("forbidden")
		}
	}
	list, total, err := s.orderRepo.ListByEvent(ctx, eventID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]responses.OrderResponse, 0, len(list))
	for i := range list {
		out = append(out, *s.toOrderResponse(&list[i]))
	}
	return out, total, nil
}

// ---------- Expiry worker ----------

func (s *orderService) ExpirePendingOrders(ctx context.Context, limit int) (int, error) {
	if limit <= 0 {
		limit = 100
	}
	orders, err := s.orderRepo.FindExpiredPending(ctx, time.Now().UTC(), limit)
	if err != nil {
		return 0, err
	}
	expired := 0
	for i := range orders {
		o := &orders[i]
		items, err := s.orderItemRepo.FindByOrder(ctx, o.ID)
		if err != nil {
			logger.Error("Failed to load items for expired order: ", err)
			continue
		}
		for _, item := range items {
			if err := s.ttRepo.ReleaseReservation(ctx, item.TicketTypeID, item.Quantity); err != nil {
				logger.Error("Failed to release reservation: ", err)
			}
		}
		o.Status = models.OrderStatusExpired
		if err := s.orderRepo.Update(ctx, o); err != nil {
			logger.Error("Failed to mark order expired: ", err)
			continue
		}
		expired++
	}
	return expired, nil
}

// ---------- helpers ----------

func (s *orderService) toOrderResponse(o *models.Order) *responses.OrderResponse {
	if o == nil {
		return nil
	}
	r := &responses.OrderResponse{
		ID:                 o.ID,
		UserID:             o.UserID,
		EventID:            o.EventID,
		Reference:          o.Reference,
		Status:             string(o.Status),
		Currency:           o.Currency,
		SubtotalMinor:      o.SubtotalMinor,
		DiscountMinor:      o.DiscountMinor,
		PlatformFeeMinor:   o.PlatformFeeMinor,
		ProcessingFeeMinor: o.ProcessingFeeMinor,
		TotalMinor:         o.TotalMinor,
		CommissionRateBps:  o.CommissionRateBps,
		CommissionMinor:    o.CommissionMinor,
		OrganizerNetMinor:  o.OrganizerNetMinor,
		PaymentProvider:    o.PaymentProvider,
		PaymentReference:   o.PaymentReference,
		PaidAt:             o.PaidAt,
		EscrowStatus:       string(o.EscrowStatus),
		EscrowReleaseAt:    o.EscrowReleaseAt,
		EscrowReleasedAt:   o.EscrowReleasedAt,
		ExpiresAt:          o.ExpiresAt,
		CreatedAt:          o.CreatedAt,
		UpdatedAt:          o.UpdatedAt,
	}
	if len(o.Items) > 0 {
		items := make([]responses.OrderItemResponse, 0, len(o.Items))
		for i := range o.Items {
			items = append(items, responses.OrderItemResponse{
				ID:             o.Items[i].ID,
				TicketTypeID:   o.Items[i].TicketTypeID,
				TicketTypeName: o.Items[i].TicketTypeName,
				Quantity:       o.Items[i].Quantity,
				UnitPriceMinor: o.Items[i].UnitPriceMinor,
				TotalMinor:     o.Items[i].TotalMinor,
			})
		}
		r.Items = items
	}
	if len(o.Tickets) > 0 {
		tickets := make([]responses.TicketResponse, 0, len(o.Tickets))
		for i := range o.Tickets {
			tickets = append(tickets, *toTicketResponse(&o.Tickets[i]))
		}
		r.Tickets = tickets
	}
	return r
}

func (s *orderService) audit(ctx context.Context, actorID *uuid.UUID, action, resourceID string) {
	log := &models.AuditLog{
		UserID:     actorID,
		Action:     action,
		Resource:   "order",
		ResourceID: resourceID,
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to write audit log: ", err)
	}
}

func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func extractBilling(jsonBytes []byte, key string) (string, bool) {
	if len(jsonBytes) == 0 || string(jsonBytes) == "null" {
		return "", false
	}
	var m map[string]interface{}
	if err := jsonUnmarshal(jsonBytes, &m); err != nil {
		return "", false
	}
	v, ok := m[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func extractBillingStr(jsonBytes []byte, key string) string {
	v, _ := extractBilling(jsonBytes, key)
	return v
}

func toTicketResponse(t *models.Ticket) *responses.TicketResponse {
	return &responses.TicketResponse{
		ID:            t.ID,
		OrderID:       t.OrderID,
		OrderItemID:   t.OrderItemID,
		TicketTypeID:  t.TicketTypeID,
		EventID:       t.EventID,
		OccurrenceID:  t.OccurrenceID,
		UserID:        t.UserID,
		Code:          t.Code,
		HolderName:    t.HolderName,
		HolderEmail:   t.HolderEmail,
		HolderPhone:   t.HolderPhone,
		Status:        string(t.Status),
		IssuedAt:      t.IssuedAt,
		UsedAt:        t.UsedAt,
		ScanLocation:  t.ScanLocation,
		TransferCount: t.TransferCount,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
}
