package db

import (
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&Certificate{}, &AuditEvent{}, &CASettings{}); err != nil {
		return nil, err
	}

	return &Database{DB: db}, nil
}

func (d *Database) CreateCertificate(cert *Certificate) error {
	return d.DB.Create(cert).Error
}

func (d *Database) GetCertificate(id string) (*Certificate, error) {
	var cert Certificate
	err := d.DB.Where("id = ?", id).First(&cert).Error
	return &cert, err
}

func (d *Database) ListCertificates(limit, offset int, status string) ([]Certificate, error) {
	var certs []Certificate
	query := d.DB.Order("created_at DESC")
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Find(&certs).Error
	return certs, err
}

func (d *Database) UpdateCertificate(cert *Certificate) error {
	return d.DB.Save(cert).Error
}

func (d *Database) DeleteCertificate(id string) error {
	return d.DB.Where("id = ?", id).Delete(&Certificate{}).Error
}

func (d *Database) LogAuditEvent(event *AuditEvent) error {
	return d.DB.Create(event).Error
}

func (d *Database) GetAuditEvents(certID string, limit int) ([]AuditEvent, error) {
	var events []AuditEvent
	query := d.DB.Order("timestamp DESC")
	
	if certID != "" {
		query = query.Where("cert_id = ?", certID)
	}
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&events).Error
	return events, err
}
