package order

import (
	"context"
	"crypto/rand"
	"math/big"

	"lazeez-core/internal/common"
)

const (
	orderNumberMin = 100000
	orderNumberMax = 999999
	maxAttempts    = 12
)

func randomOrderNumber() (int, error) {
	span := int64(orderNumberMax - orderNumberMin + 1)
	n, err := rand.Int(rand.Reader, big.NewInt(span))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + orderNumberMin, nil
}

func (s *orderService) assignOrderNumber(ctx context.Context, branchID string) (int, error) {
	for range maxAttempts {
		num, err := randomOrderNumber()
		if err != nil {
			return 0, err
		}
		exists, err := s.orderRepository.OrderNumberExists(ctx, branchID, num)
		if err != nil {
			return 0, err
		}
		if !exists {
			return num, nil
		}
	}
	return 0, common.ErrOrderNumberExhausted
}
