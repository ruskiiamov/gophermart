package bonus

import "context"

type BonusDataContainer interface {
}

type bonusManager struct {
	dataContainer BonusDataContainer
}

func NewManager(dc BonusDataContainer) *bonusManager {
	return &bonusManager{dataContainer: dc}
}

func (b *bonusManager) AddOrder(ctx context.Context, userID string, orderID int) error {
	//TODO
	return nil
}
