package http

import (
	"net/http"
	"strconv"
	"time"

	"glovo-backend/services/payment-service/internal/domain"
	"glovo-backend/shared/auth"
	"glovo-backend/shared/middleware"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService domain.PaymentService
}

func NewPaymentHandler(paymentService domain.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

func (h *PaymentHandler) SetupRoutes(router *gin.RouterGroup) {
	// Public payment processing (for order service internal calls)
	public := router.Group("/payments")
	{
		public.POST("/process", h.processPayment)
		public.POST("/validate", h.validatePaymentMethod)
	}

	// Customer wallet and payment methods
	customer := router.Group("/wallet")
	customer.Use(middleware.AuthMiddleware())
	customer.Use(middleware.RequireRole(auth.RoleCustomer))
	{
		customer.GET("/balance", h.getWalletBalance)
		customer.POST("/add-funds", h.addFunds)
		customer.GET("/transactions", h.getTransactionHistory)

		// Payment methods
		customer.POST("/payment-methods", h.addPaymentMethod)
		customer.GET("/payment-methods", h.getPaymentMethods)
		customer.PUT("/payment-methods/:id", h.updatePaymentMethod)
		customer.DELETE("/payment-methods/:id", h.deletePaymentMethod)
		customer.PUT("/payment-methods/:id/default", h.setDefaultPaymentMethod)
	}

	// Merchant earnings and payouts
	merchant := router.Group("/merchant")
	merchant.Use(middleware.AuthMiddleware())
	merchant.Use(middleware.RequireRole(auth.RoleMerchant))
	{
		merchant.GET("/earnings", h.getMerchantEarnings)
		merchant.GET("/transactions", h.getMerchantTransactions)
		merchant.POST("/withdraw", h.requestPayout)
		merchant.GET("/payouts", h.getPayoutHistory)
	}

	// Driver earnings
	driver := router.Group("/driver")
	driver.Use(middleware.AuthMiddleware())
	driver.Use(middleware.RequireRole(auth.RoleDriver))
	{
		driver.GET("/earnings", h.getDriverEarnings)
		driver.GET("/transactions", h.getDriverTransactions)
		driver.POST("/withdraw", h.requestDriverPayout)
		// driver.GET("/payouts", h.getDriverPayoutHistory) // TODO: Fix domain interface mismatch
	}

	// Admin payment management
	admin := router.Group("/admin/payments")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.RequireRole(auth.RoleAdmin))
	{
		// admin.GET("/transactions", h.getAllTransactions) // TODO: Fix domain interface mismatch
		admin.GET("/transactions/:id", h.getTransaction)
		// admin.POST("/refund", h.processRefund) // TODO: Fix domain interface mismatch
		// admin.GET("/commissions", h.getCommissions) // TODO: Fix domain interface mismatch
		// admin.POST("/commissions", h.createCommission) // TODO: Fix domain interface mismatch
		// admin.PUT("/commissions/:id", h.updateCommission) // TODO: Fix domain interface mismatch
		// admin.GET("/payouts/pending", h.getPendingPayouts) // TODO: Fix domain interface mismatch
		// admin.PUT("/payouts/:id/approve", h.approvePayout) // TODO: Fix domain interface mismatch
		// admin.PUT("/payouts/:id/reject", h.rejectPayout) // TODO: Fix domain interface mismatch
	}
}

// @Summary Process payment
// @Description Process a payment for an order
// @Tags payments
// @Accept json
// @Produce json
// @Param request body domain.ProcessPaymentRequest true "Payment data"
// @Success 200 {object} domain.Transaction
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/payments/process [post]
func (h *PaymentHandler) processPayment(c *gin.Context) {
	var req domain.ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transaction, err := h.paymentService.ProcessPayment(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// @Summary Validate payment method
// @Description Validate if a payment method can be used for a transaction
// @Tags payments
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Validation data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Router /api/v1/payments/validate [post]
func (h *PaymentHandler) validatePaymentMethod(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Placeholder validation - always return true for now
	// In a real implementation, this would validate credit card format, etc.
	c.JSON(http.StatusOK, gin.H{"valid": true})
}

// Customer endpoints

// @Summary Get wallet balance
// @Description Get the authenticated customer's wallet balance
// @Tags wallet
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Wallet
// @Failure 500 {object} map[string]string
// @Router /api/v1/wallet/balance [get]
func (h *PaymentHandler) getWalletBalance(c *gin.Context) {
	userID, _ := c.Get("user_id")

	wallet, err := h.paymentService.GetWallet(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// @Summary Add funds to wallet
// @Description Add funds to customer wallet
// @Tags wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.TopUpRequest true "Add funds data"
// @Success 200 {object} domain.Transaction
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/wallet/add-funds [post]
func (h *PaymentHandler) addFunds(c *gin.Context) {
	var req domain.TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	req.UserID = userID.(string)

	transaction, err := h.paymentService.ProcessTopUp(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// @Summary Get transaction history
// @Description Get customer's transaction history
// @Tags wallet
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {array} domain.Transaction
// @Failure 500 {object} map[string]string
// @Router /api/v1/wallet/transactions [get]
func (h *PaymentHandler) getTransactionHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	transactions, err := h.paymentService.GetTransactionHistory(userID.(string), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// @Summary Add payment method
// @Description Add a new payment method for the customer
// @Tags payment-methods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.AddPaymentMethodRequest true "Payment method data"
// @Success 201 {object} domain.PaymentMethod
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/wallet/payment-methods [post]
func (h *PaymentHandler) addPaymentMethod(c *gin.Context) {
	var req domain.AddPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	paymentMethod, err := h.paymentService.AddPaymentMethod(userID.(string), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, paymentMethod)
}

// @Summary Get payment methods
// @Description Get all payment methods for the customer
// @Tags payment-methods
// @Produce json
// @Security BearerAuth
// @Success 200 {array} domain.PaymentMethod
// @Failure 500 {object} map[string]string
// @Router /api/v1/wallet/payment-methods [get]
func (h *PaymentHandler) getPaymentMethods(c *gin.Context) {
	userID, _ := c.Get("user_id")

	paymentMethods, err := h.paymentService.GetPaymentMethods(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, paymentMethods)
}

// @Summary Set default payment method
// @Description Set a payment method as the default
// @Tags payment-methods
// @Produce json
// @Security BearerAuth
// @Param id path string true "Payment Method ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/wallet/payment-methods/{id} [put]
func (h *PaymentHandler) updatePaymentMethod(c *gin.Context) {
	paymentMethodID := c.Param("id")
	userID, _ := c.Get("user_id")

	err := h.paymentService.SetDefaultPaymentMethod(userID.(string), paymentMethodID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Default payment method updated successfully"})
}

// @Summary Delete payment method
// @Description Delete a payment method
// @Tags payment-methods
// @Produce json
// @Security BearerAuth
// @Param id path string true "Payment Method ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/wallet/payment-methods/{id} [delete]
func (h *PaymentHandler) deletePaymentMethod(c *gin.Context) {
	paymentMethodID := c.Param("id")
	userID, _ := c.Get("user_id")

	err := h.paymentService.RemovePaymentMethod(userID.(string), paymentMethodID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment method deleted successfully"})
}

// @Summary Set default payment method
// @Description Set a payment method as default
// @Tags payment-methods
// @Produce json
// @Security BearerAuth
// @Param id path string true "Payment Method ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/wallet/payment-methods/{id}/default [put]
func (h *PaymentHandler) setDefaultPaymentMethod(c *gin.Context) {
	paymentMethodID := c.Param("id")
	userID, _ := c.Get("user_id")

	err := h.paymentService.SetDefaultPaymentMethod(paymentMethodID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Default payment method set successfully"})
}

// Merchant endpoints

// @Summary Get merchant earnings
// @Description Get merchant earnings report
// @Tags merchant
// @Produce json
// @Security BearerAuth
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} domain.EarningsReport
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/merchant/earnings [get]
func (h *PaymentHandler) getMerchantEarnings(c *gin.Context) {
	userID, _ := c.Get("user_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	_, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format"})
		return
	}

	_, err = time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
		return
	}

	// Placeholder earnings report - replace with actual implementation
	earnings := map[string]interface{}{
		"user_id":        userID.(string),
		"total_earnings": 0.0,
		"period":         map[string]string{"start": startDateStr, "end": endDateStr},
		"transactions":   []interface{}{},
		"message":        "Earnings report feature coming soon",
	}

	c.JSON(http.StatusOK, earnings)
}

// @Summary Get merchant transactions
// @Description Get merchant transaction history
// @Tags merchant
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {array} domain.Transaction
// @Failure 500 {object} map[string]string
// @Router /api/v1/merchant/transactions [get]
func (h *PaymentHandler) getMerchantTransactions(c *gin.Context) {
	userID, _ := c.Get("user_id")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	transactions, err := h.paymentService.GetTransactionHistory(userID.(string), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// @Summary Request payout
// @Description Request a payout for merchant
// @Tags merchant
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]float64 true "Payout data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/merchant/withdraw [post]
func (h *PaymentHandler) requestPayout(c *gin.Context) {
	var req map[string]float64
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	amount, exists := req["amount"]
	if !exists || amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid amount is required"})
		return
	}

	userID, _ := c.Get("user_id")

	payout, err := h.paymentService.ProcessMerchantPayout(userID.(string), amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, payout)
}

// @Summary Get payout history
// @Description Get merchant payout history
// @Tags merchant
// @Produce json
// @Security BearerAuth
// @Success 200 {array} domain.Payout
// @Failure 500 {object} map[string]string
// @Router /api/v1/merchant/payouts [get]
func (h *PaymentHandler) getPayoutHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")

	// Placeholder payout history - replace with actual implementation
	payouts := []map[string]interface{}{
		{
			"message": "Payout history feature coming soon",
			"user_id": userID.(string),
		},
	}

	c.JSON(http.StatusOK, payouts)
}

// Driver endpoints (similar to merchant)

// @Summary Get driver earnings
// @Description Get driver earnings report
// @Tags driver
// @Produce json
// @Security BearerAuth
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} domain.EarningsReport
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/earnings [get]
func (h *PaymentHandler) getDriverEarnings(c *gin.Context) {
	userID, _ := c.Get("user_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	_, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format"})
		return
	}

	_, err = time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
		return
	}

	// Placeholder driver earnings report - replace with actual implementation
	earnings := map[string]interface{}{
		"user_id":        userID.(string),
		"total_earnings": 0.0,
		"period":         map[string]string{"start": startDateStr, "end": endDateStr},
		"deliveries":     []interface{}{},
		"message":        "Driver earnings report feature coming soon",
	}

	c.JSON(http.StatusOK, earnings)
}

func (h *PaymentHandler) getDriverTransactions(c *gin.Context) {
	userID, _ := c.Get("user_id")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	transactions, err := h.paymentService.GetTransactionHistory(userID.(string), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

func (h *PaymentHandler) getTransaction(c *gin.Context) {
	transactionID := c.Param("id")

	transaction, err := h.paymentService.GetTransaction(transactionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

func (h *PaymentHandler) requestDriverPayout(c *gin.Context) {
	var req struct {
		Amount float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount must be greater than 0"})
		return
	}

	userID, _ := c.Get("user_id")

	payout, err := h.paymentService.ProcessDriverPayout(userID.(string), req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, payout)
}

// TODO: Fix domain interface mismatch
/*
func (h *PaymentHandler) getDriverPayoutHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")

	payouts, err := h.paymentService.GetPayoutHistory(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payouts)
}
*/

// Admin endpoints
// TODO: The following functions are commented out due to domain interface mismatches
// They need to be fixed to match the actual domain.PaymentService interface

/*
func (h *PaymentHandler) getAllTransactions(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")
	status := c.Query("status")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	req := domain.AdminTransactionRequest{
		Limit:  limit,
		Offset: offset,
		Status: status,
	}

	transactions, err := h.paymentService.GetAllTransactions(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transactions)
}



func (h *PaymentHandler) processRefund(c *gin.Context) {
	var req domain.RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminID, _ := c.Get("user_id")
	req.ProcessedBy = adminID.(string)

	refund, err := h.paymentService.ProcessRefund(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, refund)
}

func (h *PaymentHandler) getCommissions(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	commissions, err := h.paymentService.GetCommissions(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, commissions)
}

func (h *PaymentHandler) createCommission(c *gin.Context) {
	var req domain.CreateCommissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminID, _ := c.Get("user_id")
	req.CreatedBy = adminID.(string)

	commission, err := h.paymentService.CreateCommission(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, commission)
}

func (h *PaymentHandler) updateCommission(c *gin.Context) {
	commissionID := c.Param("id")

	var req domain.UpdateCommissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminID, _ := c.Get("user_id")
	req.UpdatedBy = adminID.(string)

	commission, err := h.paymentService.UpdateCommission(commissionID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, commission)
}

func (h *PaymentHandler) getPendingPayouts(c *gin.Context) {
	payouts, err := h.paymentService.GetPendingPayouts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payouts)
}

func (h *PaymentHandler) approvePayout(c *gin.Context) {
	payoutID := c.Param("id")
	adminID, _ := c.Get("user_id")

	err := h.paymentService.ApprovePayout(payoutID, adminID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payout approved successfully"})
}

func (h *PaymentHandler) rejectPayout(c *gin.Context) {
	payoutID := c.Param("id")
	adminID, _ := c.Get("user_id")

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.paymentService.RejectPayout(payoutID, adminID.(string), req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payout rejected successfully"})
}
*/
