package model

type Course struct {
	ID            int    `gorm:"column:id;" json:"id"`
	CourseName    string `gorm:"column:course_name;" json:"course_name"`
	CourseContent string `gorm:"column:course_content;" json:"course_content"`
}

func (c *Course) TableName() string {
	return "course"
}