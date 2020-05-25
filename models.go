package main

import (
	"regexp"
	"time"
)

type JobSpider struct {
}

const (
	JobSourceCjol  = "cjol"
	JobSource51job = "51job"
	JobSourceLagou = "lagou"
)

var digitsRegexp = regexp.MustCompile(`(\d+)`)

// moders

type BaseModel struct {
	Id        int `gorm:"AUTO_INCREMENT"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Company struct {
	BaseModel
	OutId   int    `gorm:"index"`
	Name    string `gorm:"index"`
	Source  string
	Site    string
	Address string
	WebSite string
	Descr   string
}

func (Company) TableName() string {
	return "company"
}

type Job struct {
	BaseModel
	OutId      int    `gorm:"index"`
	Name       string `gorm:"index"`
	Address    string
	Site       string
	Degree     string
	Experience int
	SalaryMin  int
	SalaryMax  int
	KeyWord    string
	Descr      string
	State      int
	Source     string
	PubTime    time.Time
	CompanyId  int
	Company    Company `gorm:"ForeignKey:CompanyId;AssociationForeignKey:Refer"`
}

func (Job) TableName() string {
	return "job"
}
