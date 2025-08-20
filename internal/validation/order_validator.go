package validation

import (
	"L0/internal/kafka/dto"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

// OrderValidator contains validation logic for orders
type OrderValidator struct {
	validator *validator.Validate
}

// ValidationError represents validation errors with detailed information
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	return fmt.Sprintf("validation failed for %d field(s)", len(ve))
}

// NewOrderValidator creates a new order validator
func NewOrderValidator() *OrderValidator {
	v := validator.New()

	return &OrderValidator{
		validator: v,
	}
}

// ValidateOrder validates an order with both structural and business logic validation
func (ov *OrderValidator) ValidateOrder(order *dto.OrderDTO) error {
	if order == nil {
		return errors.New("order cannot be nil")
	}

	// Структура
	if err := ov.validator.Struct(order); err != nil {
		return ov.convertValidationErrors(err)
	}

	// Бизнес-логика
	if err := ov.validateBusinessLogic(order); err != nil {
		return err
	}

	return nil
}

// validateBusinessLogic performs custom business logic validation
func (ov *OrderValidator) validateBusinessLogic(order *dto.OrderDTO) error {
	var vErrors ValidationErrors

	// date format
	if _, err := time.Parse(time.RFC3339, order.DateCreated); err != nil {
		vErrors = append(vErrors, ValidationError{
			Field:   "date_created",
			Tag:     "datetime",
			Message: "date must be in RFC3339 format and valid",
			Value:   order.DateCreated,
		})
	}

	// Validate payment consistency
	if err := ov.validatePaymentConsistency(order); err != nil {
		var ve ValidationErrors
		if errors.As(err, &ve) {
			vErrors = append(vErrors, ve...)
		}
	}

	// Validate items consistency
	if err := ov.validateItemsConsistency(order); err != nil {
		var ve ValidationErrors
		if errors.As(err, &ve) {
			vErrors = append(vErrors, ve...)
		}
	}

	if len(vErrors) > 0 {
		return vErrors
	}

	return nil
}

// validatePaymentConsistency validates payment data consistency
func (ov *OrderValidator) validatePaymentConsistency(order *dto.OrderDTO) error {
	var vErrors ValidationErrors
	payment := order.Payment

	// payment amount
	calculatedTotal := payment.GoodsTotal + payment.DeliveryCost + payment.CustomFee
	if payment.Amount != calculatedTotal {
		vErrors = append(vErrors, ValidationError{
			Field: "payment.amount",
			Tag:   "consistency",
			Message: fmt.Sprintf("payment amount (%d) should equal goods_total + delivery_cost + custom_fee (%d)",
				payment.Amount, calculatedTotal),
		})
	}

	// transaction == orderUID
	if payment.Transaction != order.OrderUID {
		vErrors = append(vErrors, ValidationError{
			Field:   "payment.transaction",
			Tag:     "consistency",
			Message: "payment transaction should match order_uid",
		})
	}

	if len(vErrors) > 0 {
		return vErrors
	}

	return nil
}

// validateItemsConsistency validates items data consistency
func (ov *OrderValidator) validateItemsConsistency(order *dto.OrderDTO) error {
	var vErrors ValidationErrors

	totalItemsPrice := 0
	for i, item := range order.Items {
		// total_price
		expectedTotal := item.Price - (item.Price * item.Sale / 100)
		if item.TotalPrice != expectedTotal {
			vErrors = append(vErrors, ValidationError{
				Field: fmt.Sprintf("items[%d].total_price", i),
				Tag:   "consistency",
				Message: fmt.Sprintf("total_price (%d) should equal price - sale discount (%d)",
					item.TotalPrice, expectedTotal),
			})
		}

		// track_number
		if item.TrackNumber != order.TrackNumber {
			vErrors = append(vErrors, ValidationError{
				Field:   fmt.Sprintf("items[%d].track_number", i),
				Tag:     "consistency",
				Message: "item track_number should match order track_number",
			})
		}

		totalItemsPrice += item.TotalPrice
	}

	// goods_total
	if order.Payment.GoodsTotal != totalItemsPrice {
		vErrors = append(vErrors, ValidationError{
			Field: "payment.goods_total",
			Tag:   "consistency",
			Message: fmt.Sprintf("payment goods_total (%d) should equal sum of items total_price (%d)",
				order.Payment.GoodsTotal, totalItemsPrice),
		})
	}

	if len(vErrors) > 0 {
		return vErrors
	}

	return nil
}

// convertValidationErrors конвертирует ошибки валидатора в кастомный формат
func (ov *OrderValidator) convertValidationErrors(err error) error {
	var vErrors ValidationErrors

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, err := range validationErrors {
			vErrors = append(vErrors, ValidationError{
				Field:   err.Field(),
				Tag:     err.Tag(),
				Message: ov.getErrorMessage(err),
				Value:   fmt.Sprintf("%v", err.Value()),
			})
		}
	}

	return vErrors
}

// getErrorMessage возвращает читаемые ошибки для пользователя
func (ov *OrderValidator) getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "field is required"
	case "min":
		return fmt.Sprintf("field must be at least %s characters/value", err.Param())
	case "max":
		return fmt.Sprintf("field must be at most %s characters/value", err.Param())
	case "email":
		return "field must be a valid email address"
	case "oneof":
		return fmt.Sprintf("field must be one of: %s", err.Param())
	case "datetime":
		return "field must be a valid datetime in RFC3339 format"
	default:
		return fmt.Sprintf("field failed validation: %s", err.Tag())
	}
}
