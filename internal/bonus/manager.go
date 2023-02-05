package bonus

type BonusDataContainer interface {

} 

type bonusManager struct {
	dataContainer BonusDataContainer
}

func NewManager(dc BonusDataContainer) *bonusManager {
	return &bonusManager{dataContainer: dc}
}