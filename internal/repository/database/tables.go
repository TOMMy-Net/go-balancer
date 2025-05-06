package database

type Clients struct {
	ID              uint64 `gorm:"primaryKey"`
	IP              string `gorm:"not null"`
	Capacity        int    `gorm:"not null;default:'100'"`
	RatePerInterval int    `gorm:"not null;default:'5'"`
	Tokens          int    `gorm:"not null;default:'100'"`
}
