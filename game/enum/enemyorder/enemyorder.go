package enemyorder

type EnemyOrder string

const (
	Default EnemyOrder = "Default"
	F4      EnemyOrder = "F4"
)

func (eo EnemyOrder) String() string {
	return string(eo)
}
